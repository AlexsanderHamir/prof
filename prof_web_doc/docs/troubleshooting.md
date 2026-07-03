# Troubleshooting

This page lists common failure modes when using Prof and what to change before opening an issue.

### prof ui or prof tui fails in CI or IDEs {#prof-ui-or-prof-tui-fails-in-ci-or-ides}

Menus never appear, or you see an error about stdin or stdout not being a TTY.

`prof ui` and `prof tui` expect an interactive terminal.

Use non-interactive commands instead: `prof auto` and `prof manual`. See [Quickstart](quickstart.md) Path B and [CLI reference](cli-reference.md).

### Wrong directory or no bench folder {#wrong-directory-or-no-bench-folder}

Collect succeeds elsewhere, or `bench/` appears in the wrong repo.

Prof writes `bench/` under the current working directory, which must be your module root (where `go.mod` lives).

`cd` to the same directory you use for `go test`, then run Prof again. See [Working directory and paths](workspace.md).

### Graphviz or PNG errors {#graphviz-png-errors}

Failure when generating call-graph PNGs.

PNG generation uses Graphviz `dot` when installed. Without Graphviz, prof warns during collection and still writes text profiles.

Install [Graphviz](https://graphviz.org/) for call-graph PNGs. See [Collect profiling data](collect.md).

## Next steps

- [CLI reference](cli-reference.md) for flags and defaults

## Related

- [Home](index.md)
