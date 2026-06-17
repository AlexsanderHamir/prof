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
	if err := createBenchDirectories(tagDir, benchmarks); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		path string
	}{
		{"bin root", filepath.Join(tagDir, workspace.ProfileBinDir)},
		{"text root", filepath.Join(tagDir, workspace.ProfileTextDir)},
		{"description", filepath.Join(tagDir, workspace.BenchDescriptionFileName)},
		{"bin bench foo", filepath.Join(tagDir, workspace.ProfileBinDir, "BenchmarkFoo")},
		{"text bench bar", filepath.Join(tagDir, workspace.ProfileTextDir, "BenchmarkBar")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			st, err := os.Stat(tc.path)
			if err != nil {
				t.Fatalf("stat %s: %v", tc.path, err)
			}
			if !st.IsDir() && tc.name != "description" {
				t.Fatalf("%s should be a directory", tc.path)
			}
		})
	}
}

func TestCreateProfileFunctionDirectories(t *testing.T) {
	t.Parallel()
	tagDir := t.TempDir()
	profiles := []string{"cpu", "memory"}
	benchmarks := []string{"BenchmarkFoo"}

	if err := createProfileFunctionDirectories(tagDir, profiles, benchmarks); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(tagDir, "cpu"+workspace.FunctionsDirSuffix, "BenchmarkFoo")
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

	if err := setupDirectories(tag, benchmarks, profiles); err != nil {
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
			"bin bench dir",
			filepath.Join(layout.Root, workspace.ProfileBinDir, "BenchmarkFoo"),
			true,
		},
		{
			"text bench dir",
			filepath.Join(layout.Root, workspace.ProfileTextDir, "BenchmarkFoo"),
			true,
		},
		{
			"functions dir",
			layout.FunctionsDir("cpu", "BenchmarkFoo"),
			true,
		},
		{
			"memory functions dir",
			layout.FunctionsDir("memory", "BenchmarkFoo"),
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
