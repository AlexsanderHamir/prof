package collect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/config"
	"github.com/AlexsanderHamir/prof/internal/testpaths"
)

const (
	filterBenchName     = "BenchmarkStringProcessor"
	filterFixtureCPU    = filterBenchName + "_cpu.out"
	filterFixtureMemory = filterBenchName + "_memory.out"
)

func TestWriteGroupedPackageProfile_fixture(t *testing.T) {
	t.Parallel()
	cpuPath := testpaths.MustAsset(t, "fixtures", filterFixtureCPU)
	outPath := filepath.Join(t.TempDir(), "grouped.txt")

	if err := writeGroupedPackageProfile(cpuPath, outPath, config.FunctionFilter{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty grouped profile output")
	}
}

func TestWriteGroupedPackageProfile_respectsFilter(t *testing.T) {
	t.Parallel()
	cpuPath := testpaths.MustAsset(t, "fixtures", filterFixtureCPU)
	outPath := filepath.Join(t.TempDir(), "grouped_filtered.txt")

	cfg := &config.Config{
		Collection: config.Collection{
			Benchmarks: map[string]config.FunctionFilter{
				filterBenchName: {IgnoreFunctions: []string{filterBenchName}},
			},
		},
	}
	filter := config.ResolveCollectionFilter(cfg, config.CollectionTargetAuto(filterBenchName))

	if err := writeGroupedPackageProfile(cpuPath, outPath, filter); err != nil {
		t.Fatal(err)
	}
	text := string(mustReadFile(t, outPath))
	if strings.Contains(text, filterBenchName) {
		t.Fatalf("grouped output should omit ignored benchmark name %q", filterBenchName)
	}
}

func TestResolveCollectionFilter_groupedPipeline(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		Collection: config.Collection{
			Benchmarks: map[string]config.FunctionFilter{
				filterBenchName: {IncludePrefixes: []string{"test-environment"}},
			},
		},
	}
	filter := config.ResolveCollectionFilter(cfg, config.CollectionTargetAuto(filterBenchName))
	memPath := testpaths.MustAsset(t, "fixtures", filterFixtureMemory)
	outPath := filepath.Join(t.TempDir(), "grouped_include.txt")

	if err := writeGroupedPackageProfile(memPath, outPath, filter); err != nil {
		t.Fatal(err)
	}
	if len(mustReadFile(t, outPath)) == 0 {
		t.Fatal("expected grouped output with include filter")
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
