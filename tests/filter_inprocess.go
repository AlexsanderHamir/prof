package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/engine/collector"
	"github.com/AlexsanderHamir/prof/internal"
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

// runFilterInProcess replays the per-function extraction stage of `prof auto`
// against the committed fixtures: for each profile kind, it calls
// parser.GetAllFunctionNamesV2 with the configured FunctionFilter, then writes
// per-function .txt files via collector.GetFunctionsOutput. The returned map
// is keyed by profile kind and points at the directory containing those .txt
// files (mirroring `bench/<tag>/<profile>_functions/<bench>/`).
//
// This is the fast path: no `go test -bench`, no prof subprocess, no
// per-scenario synthetic Go module. The committed pprof binary makes filter
// behavior deterministic, so a single run is enough — no count=10 needed.
func runFilterInProcess(t *testing.T, cfg internal.Config) map[string]string {
	t.Helper()

	root, rootErr := getProjectRoot()
	if rootErr != nil {
		t.Fatalf("getProjectRoot: %v", rootErr)
	}
	fixturesPath := filepath.Join(root, testDirName, fixturesSubdir)

	filter := cfg.FunctionFilter[benchName]

	outDirs := make(map[string]string, len(allFixtureProfiles))
	tagDir := filepath.Join(t.TempDir(), internal.MainDirOutput, tag)

	for _, fp := range allFixtureProfiles {
		fixturePath := filepath.Join(fixturesPath, fp.fileName)
		if _, statErr := os.Stat(fixturePath); statErr != nil {
			t.Fatalf("fixture missing (%s); run `go generate ./tests/...`: %v", fixturePath, statErr)
		}

		funcDir := filepath.Join(tagDir, fp.kind+internal.FunctionsDirSuffix, benchName)
		if mkErr := os.MkdirAll(funcDir, internal.PermDir); mkErr != nil {
			t.Fatalf("mkdir %s: %v", funcDir, mkErr)
		}

		functions, parseErr := parser.GetAllFunctionNamesV2(fixturePath, filter)
		if parseErr != nil {
			t.Fatalf("GetAllFunctionNamesV2(%s): %v", fp.kind, parseErr)
		}

		if writeErr := collector.GetFunctionsOutput(functions, fixturePath, funcDir); writeErr != nil {
			t.Fatalf("GetFunctionsOutput(%s): %v", fp.kind, writeErr)
		}

		outDirs[fp.kind] = funcDir
	}

	return outDirs
}

// checkFunctionDirAgainstSpec asserts every per-function .txt in dir is either
// expected (per specifiedFiles) or unfiltered noise (per expectNonSpecified).
// Behavior matches validateFileWithConfig: missing-but-expected files are not
// errors (sampling jitter could omit them), but unexpected files always fail.
func checkFunctionDirAgainstSpec(t *testing.T, dir string, specifiedFiles map[fileFullName]*FieldsCheck, expectNonSpecified bool) {
	t.Helper()

	files, dirNames := getFileOrDirs(t, dir, dir)
	if len(dirNames) > 0 {
		t.Fatalf("%s contains unexpected directories: %v", dir, dirNames)
	}
	if len(files) == 0 {
		t.Fatalf("%s contains no files", dir)
	}

	for _, file := range files {
		name := file.Name()
		if specifiedFiles != nil {
			validateFileWithConfig(t, name, expectNonSpecified, specifiedFiles)
		}
		checkFileNotEmpty(t, filepath.Join(dir, name), name)
	}
}
