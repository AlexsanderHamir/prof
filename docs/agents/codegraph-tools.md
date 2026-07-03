# Codegraph MCP tools — agent guide

Codegraph is a per-project SQLite index of symbols, call edges, and file metadata. Every MCP tool on the `user-codegraph` server reads from that index rather than scanning the tree on each request, which makes symbol lookup and call-graph traversal faster and more structured than ad-hoc Grep and Read loops. The index lives under `.codegraph/` at the project root and is built with `codegraph init`; until it exists, none of the tools below apply and you should fall back to Read, Grep, and Glob. All tools accept `projectPath` (an absolute path to the repo or any subdirectory inside it) because the server has no default project when the workspace root itself is unindexed — pass the repo root path or equivalent on every call.

---

## Starting with the index

Before any other tool, confirm the index is present and healthy with `codegraph_status`. On a freshly initialized prof-polish repo this returned 135 indexed files, 1,351 nodes, and 2,937 edges along with a breakdown by symbol kind (functions, methods, imports, and so on) and by language. The tool adds no semantic value during normal exploration — its job is diagnostics when a query returns empty results, after a large refactor, or when you suspect the index is stale. **Pros:** one call tells you whether Codegraph is usable at all. **Cons:** no code, no graph, no file listing. **When to use:** once at session start if unsure, or when other tools fail unexpectedly. **Combine with:** nothing routinely; if status shows zero files, run `codegraph init` and retry.

Once the index is confirmed, orient yourself in the tree with `codegraph_files`. Filtering `path=engine/collect` and `format=tree` produced the same 21 Go files that Glob finds, but each entry included a symbol count (for example `entry.go` with 11 symbols) and the output was already hierarchical. The `grouped` format rolls files up by language, and `pattern` accepts globs such as `*.go`. **Pros:** faster than Glob for layout questions because results come from the index and carry symbol density; supports depth limits and flat/tree/grouped views. **Cons:** only lists indexed source files — configs, markdown, and other non-parsed paths are absent, with no file timestamps or sizes. **When to use:** "what lives under this package?" or "how big is this area?" before diving into symbols. **Combine with:** `codegraph_search` next, using the file paths you just discovered as disambiguation hints.

---

## Finding symbols

`codegraph_search` resolves a name (full or partial) to definition sites. Searching `RunAuto` with `kind=function` returned a single canonical hit at `engine/collect/entry.go:15` with its signature, while Grep for `func RunAuto` found only that one definition but missed six interface methods and test stubs that share the name, and a plain Grep for `RunAuto` returned twenty-five noisy hits spanning comments, call sites, and documentation. **Pros:** definition-oriented, typed results with file, line, and signature; optional `kind` filter narrows to functions, methods, classes, routes, components, and other node kinds. **Cons:** locations only — no source bodies; requires knowing (or guessing) a symbol name rather than a concept. **When to use:** you know what something is called and need to find where it is defined. **Combine with:** `codegraph_node` on the hit you care about, passing `file` when the name is overloaded.

---

## Reading source

`codegraph_node` has two modes and is the workhorse for anything that would otherwise use Read. In **file mode**, pass `file` alone (path or basename). Reading `engine/collect/entry.go` returned the same line-numbered source as the Read tool, prefixed with a header noting the file is used by two others (`manual_test.go` and `internal/app/defaults.go`). Setting `symbolsOnly=true` instead returned just the three symbol signatures and line numbers — a cheap structural preview. `offset` and `limit` paginate exactly like Read. A request for a non-existent `auto.go` failed with a clear message to use Read for non-indexed paths. **Pros:** drop-in Read replacement with dependents attached; index-backed so repeated reads are cheap. **Cons:** capped at 2,000 lines; wrong filenames error rather than falling back.

In **symbol mode**, pass `symbol` (optionally with `file` and `line` to pin one overload). Querying `RunAuto` with `includeCode=true` and no disambiguation returned all seven definitions — interface methods, test fakes, and the real function — each with its full body and a caller/callee trail. Pinning with `file=engine/collect/entry.go` and `line=15` collapsed that to the single production definition plus its trail: calls to `applyAutoSkipPNG`, `config.Load`, `setupDirectories`, and `runBenchAndGetProfiles`; called by `defaultCollect.RunAuto` and a test. **Pros:** replaces search + Read + manual trace for one symbol; disambiguates overloaded names without opening every matching file. **Cons:** ambiguous names with `includeCode=true` produce very large responses; trails are one hop unless you follow them manually.

**When to use file mode:** any time you would Read a source file, especially when you also want to know who imports or depends on it. **When to use symbol mode:** before editing a function, method, or type — always prefer pinning with `file` when the name appears in tests or interfaces. **Combine with:** `codegraph_callers` and `codegraph_callees` when you need the full caller or callee list rather than the inline trail.

---

## Traversing the call graph

The three graph tools slice the same edge data along different axes. `codegraph_callers` lists who invokes a symbol; without a `file` filter on `RunAuto` it grouped seven distinct definitions (interface declarations, test stubs, and the real function), but adding `file=engine/collect/entry.go` narrowed to two direct callers: `TestRunAuto_validation` and `defaultCollect.RunAuto`. Dynamic interface dispatch is labeled explicitly (for example `RunAuto (internal/app/services.go:13) [dynamic: interface → impl @…]`). **Pros:** precise inbound edges, grouped per definition, with interface resolution noted. **Cons:** common names require `file`; default limit is twenty; no source code.

`codegraph_callees` mirrors this for outbound calls. The production `RunAuto` at `entry.go:15` calls nine functions and references several types in one hop. **Pros:** immediate dependency list for a single symbol. **Cons:** one level deep only; type references mixed with function calls can clutter the picture.

`codegraph_impact` walks multiple hops to answer "what breaks if I change this?" Analyzing `GetFunctionListEntriesV2` at `depth=2` surfaced four affected symbols across `parser/facade.go` and `parser/coverage_more_test.go`. The `file` parameter disambiguates when several symbols share a name; `depth` controls how far the traversal reaches (default 2). **Pros:** refactor planning in one call instead of chaining callers manually. **Cons:** output is a flat symbol list with no ranking by risk and no code context.

**When to use callers:** before changing a function signature or deleting an export — verify every inbound site. **When to use callees:** understanding what a function orchestrates without reading its entire body. **When to use impact:** scoping a rename, signature change, or extraction across packages. **Combine with:** `codegraph_node` with `includeCode=true` on the highest-risk symbols impact returns, then `codegraph_callers` again if you need the complete inbound set beyond impact's depth window.

---

## The missing aggregate tool

Server documentation and several tool descriptions reference `codegraph_explore` as a single-call replacement for search + node + callers + callees when understanding an area. **It is not registered on the MCP server** (calls return "tool not found" as of this benchmark). Until it ships, reproduce its intent with a fixed four-step recipe: `codegraph_search` to locate the entry symbol, `codegraph_node` with `includeCode=true` and a `file` pin to read the body and inline trail, `codegraph_callers` with the same `file` pin for the full inbound set, and `codegraph_callees` for outbound dependencies. That sequence understood the collect pipeline's `RunAuto` entry point in four round trips versus six or more with Grep and Read.

---

## Named workflows

**New area onboarding.** Run `codegraph_files` on the package directory to see file count and symbol density, then `codegraph_search` for the entry function or route name you expect, then `codegraph_node` in symbol mode with `includeCode=true` and a `file` pin. This replaces Glob → Grep → Read with three indexed calls and gives you dependents and a one-hop trail for free.

**Refactor prep.** After `codegraph_search` locates the symbol, call `codegraph_impact` with an appropriate `depth` (2 for local changes, higher for API surface changes). For each returned symbol, call `codegraph_node` with `includeCode=true` to inspect bodies, and `codegraph_callers` with `file` if impact's depth missed a caller you care about.

**Pin an ambiguous symbol.** Call `codegraph_node` with just `symbol` to list all definitions (twelve matches for `Run` in prof-polish). Pick the `file:line` from the list, then re-query with `symbol`, `file`, and `line` to get one body and trail. This avoids the token cost of `includeCode=true` on every overload upfront.

**Read a file before editing.** Prefer `codegraph_node` with `file` over Read: same source format, plus a dependents note that tells you whether the edit might ripple. Use `symbolsOnly=true` first if the file is long and you only need to know what symbols it exports.

---

## When to fall back to built-in tools

| Situation | Use instead |
| --- | --- |
| No `.codegraph/` index or `codegraph_status` shows zero files | Read, Grep, Glob; run `codegraph init` |
| Non-source files (markdown, YAML configs, `.env`, CI) | Read or Grep — only parsed languages are indexed |
| Wrong or stale path after a rename | Read directly; re-run `codegraph init` if the index is old |
| Concept search ("where is auth handled?") | SemanticSearch or Grep — Codegraph needs a symbol name |
| String literal or comment search | Grep — the index tracks symbols and call edges, not all text |
| `codegraph_explore`-style area query | Four-tool recipe above until explore is available |

---

## Tool summary

| Tool | Best for | Avoid when |
| --- | --- | --- |
| `codegraph_status` | Index health check | Normal exploration |
| `codegraph_files` | Package layout, symbol counts | Non-indexed paths |
| `codegraph_search` | Find definitions by name | You need source or call graph |
| `codegraph_node` (file) | Read source + dependents | File not in index |
| `codegraph_node` (symbol) | Body + trail before edit | Ambiguous name without `file` pin |
| `codegraph_callers` | Full inbound call list | Name is overloaded without `file` |
| `codegraph_callees` | Outbound dependencies | You need multi-hop reach |
| `codegraph_impact` | Refactor blast radius | You need code bodies inline |
| `codegraph_explore` | *(unavailable)* | — |

All observations above were measured against the prof-polish Go codebase (135 files, 1,351 nodes) using symbols along the collect→parser call chain (`RunAuto`, `GetFunctionListEntriesV2`, and related paths). Adjust `depth`, `limit`, and disambiguation habits for larger monorepos where name collisions are more frequent.
