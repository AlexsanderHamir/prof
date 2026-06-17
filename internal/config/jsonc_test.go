package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStripJSONComments_removesLineComments(t *testing.T) {
	in := []byte(`{"a": 1 // comment
}`)
	out := stripJSONComments(in)
	if strings.Contains(string(out), "comment") {
		t.Fatalf("comment not stripped: %q", out)
	}
	var v map[string]int
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatal(err)
	}
	if v["a"] != 1 {
		t.Fatalf("got %v", v)
	}
}

func TestLoadFromFile_withComments(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/foo\n\ngo 1.24\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	content := DefaultTemplate("example.com/foo")
	if err := os.WriteFile(filepath.Join(root, "prof.json"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Collection.Defaults.IncludePrefixes) != 1 || loaded.Collection.Defaults.IncludePrefixes[0] != "example.com/foo" {
		t.Fatalf("include_prefixes: %+v", loaded.Collection.Defaults.IncludePrefixes)
	}
	if loaded.Track.Defaults.MaxRegressionPercent != 15 {
		t.Fatalf("track defaults: %+v", loaded.Track.Defaults)
	}
}

func TestCreateDefaultFile_writesCommentedTemplate(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/bar\n\ngo 1.24\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	if err := CreateDefaultFile(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "prof.json"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "// include_prefixes:") {
		t.Fatal("expected include_prefixes comment")
	}
	if !strings.Contains(text, "example.com/bar") {
		t.Fatal("expected module path in template")
	}
	if _, err = Load(); err != nil {
		t.Fatalf("created file should load: %v", err)
	}
}

func TestDefaultTemplate_loadsAfterCommentStrip(t *testing.T) {
	tmpl := DefaultTemplate("github.com/org/app")
	var c Config
	if err := json.Unmarshal(stripJSONComments([]byte(tmpl)), &c); err != nil {
		t.Fatal(err)
	}
	if len(c.Collection.Defaults.IncludePrefixes) != 1 {
		t.Fatalf("got %+v", c.Collection.Defaults.IncludePrefixes)
	}
}
