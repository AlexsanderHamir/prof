package testpaths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModuleRoot_FromNestedDir(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n\ngo 1.24\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)

	got, err := ModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(root)
	if err != nil {
		want = root
	}
	gotResolved, err := filepath.EvalSymlinks(got)
	if err != nil {
		gotResolved = got
	}
	if filepath.Clean(gotResolved) != filepath.Clean(want) {
		t.Fatalf("ModuleRoot() = %q want %q", gotResolved, want)
	}
}

func TestTestsAssetsDir_JoinsCorrectly(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n\ngo 1.24\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	dir, err := TestsAssetsDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "tests", "assets")
	if filepath.Clean(dir) != filepath.Clean(want) {
		t.Fatalf("TestsAssetsDir() = %q want %q", dir, want)
	}

	cpu, err := Asset("cpu.out")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(cpu) != "cpu.out" || filepath.Dir(cpu) != want {
		t.Fatalf("Asset() = %q want .../tests/assets/cpu.out", cpu)
	}
}
