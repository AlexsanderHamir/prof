package tests

import "github.com/AlexsanderHamir/prof/internal"

// Test environment, run, and fixture constants. Every literal that the
// integration suite cares about lives here; if you find yourself adding a
// new string in a test or helper, hoist it into this file first.
const (
	envDirNameStatic = "Enviroment" // legacy spelling preserved on disk
	sharedEnvLabel   = "shared"
	templateFile     = "config_template.json"
	testDirName      = "tests"
	tag              = "test_tag"

	moduleName     = "test-environment"
	utilsPkgPrefix = "utils."

	cpuProfile   = "cpu"
	memProfile   = "memory"
	blockProfile = "block"

	benchName     = "BenchmarkStringProcessor"
	funcProcess   = "ProcessStrings"
	funcGenerate  = "GenerateStrings"
	funcAddString = "AddString"

	// Run counts. count=10 was needed only because each filter scenario
	// re-sampled CPU; with shared fixtures, smokeCount is plenty for
	// wiring smoke tests and validationCount is enough for error-path checks.
	count           = "10"
	smokeCount      = "1"
	validationCount = "1"

	fixturesSubdir = "assets/fixtures"
	fixtureCPUFile = benchName + "_" + cpuProfile + ".out"
	fixtureMemFile = benchName + "_" + memProfile + ".out"
)

// expectedFunctionFiles names every per-function .txt file the committed
// pprof fixtures (or a real prof auto run) are expected to surface.
// Test scenarios reference these instead of repeating string literals.
var expectedFunctionFiles = []string{
	benchName + "." + internal.TextExtension,
	funcProcess + "." + internal.TextExtension,
	funcGenerate + "." + internal.TextExtension,
	funcAddString + "." + internal.TextExtension,
}

// filterIncludePrefixes mirrors the IncludePrefixes the original TestConfig
// scenarios used: anything starting with the synthetic module name or the
// utils package prefix.
var filterIncludePrefixes = []string{moduleName, utilsPkgPrefix}

// filterIgnoreFunctions mirrors the IgnoreFunctions the original TestConfig
// scenarios used: short names after the last '.' that should be filtered out.
var filterIgnoreFunctions = []string{benchName, funcProcess, funcAddString}
