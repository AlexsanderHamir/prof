package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/workspace"

	"github.com/AlexsanderHamir/prof/parser"
)

// fixtureProfile pairs a profile kind ("cpu", "memory") with the fixture
// filename committed under tests/assets/fixtures/.
type fixtureProfile struct {
	kind     string
	fileName string
}

// allFixtureProfiles is the canonical set of committed pprof binaries the
// in-process filter driver replays. Update alongside fixtures_regen.go when
// adding a new profile kind.
var allFixtureProfiles = []fixtureProfile{
	{kind: cpuProfile, fileName: fixtureCPUFile},
	{kind: memProfile, fileName: fixtureMemFile},
}

// runFilterInProcess returns, per profile kind, the list of function names
// that survive cfg.FunctionFilter[benchName] when applied to the committed
// pprof fixture. This is the exact set `prof auto` would produce a
// per-function .txt file for; collapsing the test to the filtered name list
// avoids ~100 `go tool pprof -list` shellouts per scenario without losing
// coverage — the GetFunctionsOutput round-trip is already covered by the
// collector unit tests.
func runFilterInProcess(t *testing.T, cfg config.Config) map[string][]string {
	t.Helper()

	root, err := getProjectRoot()
	if err != nil {
		t.Fatalf("getProjectRoot: %v", err)
	}
	fixturesPath := filepath.Join(root, testDirName, fixturesSubdir)

	filter := config.ResolveCollectionFilter(&cfg, config.CollectionTargetAuto(benchName))

	out := make(map[string][]string, len(allFixtureProfiles))
	for _, fp := range allFixtureProfiles {
		fixturePath := filepath.Join(fixturesPath, fp.fileName)
		if _, statErr := os.Stat(fixturePath); statErr != nil {
			t.Fatalf("fixture missing (%s); run `go generate ./tests/...`: %v", fixturePath, statErr)
		}

		names, parseErr := parser.GetAllFunctionNamesV2(fixturePath, filter)
		if parseErr != nil {
			t.Fatalf("GetAllFunctionNamesV2(%s): %v", fp.kind, parseErr)
		}
		out[fp.kind] = names
	}

	return out
}

// checkFilteredNamesAgainstSpec verifies the filtered function name list for
// one profile kind matches the spec — same semantics as validateFileWithConfig
// but against the in-memory list instead of the on-disk per-function dir.
//
// Behavior preserved from the original integration test:
//   - if a name is in spec and isFileExpected: bump currentAppearances
//   - if a name is in spec and !isFileExpected: t.Fatal (it should've been filtered)
//   - if a name is NOT in spec and !expectNonSpec: t.Fatal (unexpected leak)
//   - missing-but-expected names are tolerated (filter+sampling can omit them)
func checkFilteredNamesAgainstSpec(t *testing.T, names []string, spec map[fileFullName]*FieldsCheck, expectNonSpec bool) {
	t.Helper()

	if len(names) == 0 {
		t.Fatalf("filter produced empty function list — fixture corruption or overly aggressive filter?")
	}

	for _, name := range names {
		fileName := name + "." + workspace.TextExtension
		if spec != nil {
			validateFileWithConfig(t, fileName, expectNonSpec, spec)
		}
	}
}
