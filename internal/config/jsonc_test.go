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

func TestLoadFromFile_withCommentsInExample(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/foo\n\ngo 1.24\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	content := ExampleTemplate("example.com/foo")
	if err := os.WriteFile(filepath.Join(root, ExampleFilename), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	var c Config
	if err := json.Unmarshal(stripJSONComments([]byte(content)), &c); err != nil {
		t.Fatal(err)
	}
	if len(c.Collection.Defaults.IncludePrefixes) != 1 {
		t.Fatalf("include_prefixes: %+v", c.Collection.Defaults.IncludePrefixes)
	}
}

func TestCreateDefaultFile_writesValidJSONAndExample(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/bar\n\ngo 1.24\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	if err := CreateDefaultFile(); err != nil {
		t.Fatal(err)
	}

	profData, err := os.ReadFile(filepath.Join(root, Filename))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(profData), "//") {
		t.Fatal("prof.json must be valid JSON without comments")
	}
	if _, err = Load(); err != nil {
		t.Fatalf("prof.json should load: %v", err)
	}
	if strings.Contains(string(profData), `"collection"`) || strings.Contains(string(profData), `"track"`) {
		t.Fatalf("prof.json should be minimal (version only), got: %s", profData)
	}

	exampleData, err := os.ReadFile(filepath.Join(root, ExampleFilename))
	if err != nil {
		t.Fatal(err)
	}
	example := string(exampleData)
	if !strings.Contains(example, "// include_prefixes:") {
		t.Fatal("expected include_prefixes comment in example file")
	}
	if !strings.Contains(example, "example.com/bar") {
		t.Fatal("expected module path in example file")
	}
	for _, anchor := range []string{
		docSiteBase + "/configure/#collection",
		docSiteBase + "/configure/#collection-benchmarks",
		docSiteBase + "/configure/#collection-manual-profiles",
		docSiteBase + "/configure/#track",
		docSiteBase + "/configure/#track-benchmarks",
		docSiteBase + "/collect/#artifact-layout-under-benchtag",
		docSiteBase + "/compare/#regression-gate",
		docSiteBase + "/ci/#json-in-profjson",
	} {
		if !strings.Contains(example, anchor) {
			t.Fatalf("expected doc link %q in example file", anchor)
		}
	}
}

func TestExampleTemplate_loadsAfterCommentStrip(t *testing.T) {
	tmpl := ExampleTemplate("github.com/org/app")
	var c Config
	if err := json.Unmarshal(stripJSONComments([]byte(tmpl)), &c); err != nil {
		t.Fatal(err)
	}
	if len(c.Collection.Defaults.IncludePrefixes) != 1 {
		t.Fatalf("got %+v", c.Collection.Defaults.IncludePrefixes)
	}
	if c.Version != CurrentVersion {
		t.Fatalf("version %d", c.Version)
	}
	if len(c.Track.Defaults.IgnorePrefixes) != 3 {
		t.Fatalf("track defaults: %+v", c.Track.Defaults)
	}
}

func TestExampleTemplate_containsDocLinks(t *testing.T) {
	tmpl := ExampleTemplate("example.com/mod")
	for _, anchor := range []string{
		"/configure/#collection-benchmarks",
		"/collect/#prof-manual",
		"/compare/#regression-gate",
	} {
		if !strings.Contains(tmpl, anchor) {
			t.Fatalf("expected %q in template", anchor)
		}
	}
}
