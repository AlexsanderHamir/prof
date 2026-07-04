package collect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal/workspace"
)

func writeModuleRoot(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module layouttest\n\ngo 1.24\n"), workspace.PermFile); err != nil {
		t.Fatal(err)
	}
}

func TestCreateBenchDirectories(t *testing.T) {
	t.Parallel()
	tagDir := filepath.Join(t.TempDir(), "tag")
	if err := os.Mkdir(tagDir, workspace.PermDir); err != nil {
		t.Fatal(err)
	}

	benchmarks := []string{"BenchmarkFoo", "BenchmarkBar"}
	if err := createBenchDirectories(tagDir, benchmarks, false); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		path  string
		isDir bool
	}{
		{"profiles root", filepath.Join(tagDir, workspace.ProfilesDir), true},
		{"measurements root", filepath.Join(tagDir, workspace.MeasurementsDir), true},
		{"hotspots root", filepath.Join(tagDir, workspace.HotspotsDir), true},
		{"notes", filepath.Join(tagDir, workspace.TagNotesFileName), false},
		{"profiles bench foo", filepath.Join(tagDir, workspace.ProfilesDir, "BenchmarkFoo"), true},
		{"measurements bench bar", filepath.Join(tagDir, workspace.MeasurementsDir, "BenchmarkBar"), true},
		{"hotspots bench bar", filepath.Join(tagDir, workspace.HotspotsDir, "BenchmarkBar"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			st, err := os.Stat(tc.path)
			if err != nil {
				t.Fatalf("stat %s: %v", tc.path, err)
			}
			if tc.isDir && !st.IsDir() {
				t.Fatalf("%s should be a directory", tc.path)
			}
			if !tc.isDir && st.IsDir() {
				t.Fatalf("%s should be a file", tc.path)
			}
		})
	}
}

func TestCreateSourceLinesDirectories(t *testing.T) {
	t.Parallel()
	tagDir := t.TempDir()
	profiles := []string{"cpu", "memory"}
	benchmarks := []string{"BenchmarkFoo"}

	if err := createSourceLinesDirectories(tagDir, profiles, benchmarks, false); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(tagDir, workspace.SourceLinesDir, "cpu", "BenchmarkFoo")
	if st, err := os.Stat(want); err != nil || !st.IsDir() {
		t.Fatalf("expected directory %s: err=%v isDir=%v", want, err, st != nil && st.IsDir())
	}
}

func TestSetupDirectories_createsTagLayout(t *testing.T) {
	modRoot := t.TempDir()
	writeModuleRoot(t, modRoot)
	t.Chdir(modRoot)

	tag := "run1"
	benchmarks := []string{"BenchmarkFoo"}
	profiles := []string{"cpu", "memory"}

	if err := setupDirectories(tag, benchmarks, profiles, false); err != nil {
		t.Fatal(err)
	}

	layout, err := workspace.TagLayoutFromCWD(tag)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		path string
		dir  bool
	}{
		{
			"profiles bench dir",
			filepath.Join(layout.Root, workspace.ProfilesDir, "BenchmarkFoo"),
			true,
		},
		{
			"measurements bench dir",
			filepath.Join(layout.Root, workspace.MeasurementsDir, "BenchmarkFoo"),
			true,
		},
		{
			"hotspots bench dir",
			filepath.Join(layout.Root, workspace.HotspotsDir, "BenchmarkFoo"),
			true,
		},
		{
			"source lines dir",
			layout.SourceLinesDir("cpu", "BenchmarkFoo"),
			true,
		},
		{
			"memory source lines dir",
			layout.SourceLinesDir("memory", "BenchmarkFoo"),
			true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st, statErr := os.Stat(tc.path)
			if statErr != nil {
				t.Fatalf("path missing %s: %v", tc.path, statErr)
			}
			if tc.dir && !st.IsDir() {
				t.Fatalf("%s should be a directory", tc.path)
			}
		})
	}
}
