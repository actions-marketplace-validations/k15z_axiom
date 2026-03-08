package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func writeYAML(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscover_Ordering(t *testing.T) {
	dir := t.TempDir()

	// Files named so lexicographic order is: a.yml, b.yml, c.yaml
	writeYAML(t, dir, "c.yaml", `
second_in_c:
  condition: "c second"
first_in_c:
  condition: "c first"
`)
	writeYAML(t, dir, "a.yml", `
only_in_a:
  condition: "a only"
`)
	writeYAML(t, dir, "b.yml", `
beta:
  condition: "b beta"
alpha:
  condition: "b alpha"
`)

	tests, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	// Expected order: a.yml tests, then b.yml tests (YAML key order), then c.yaml tests
	wantNames := []string{"only_in_a", "beta", "alpha", "second_in_c", "first_in_c"}
	if len(tests) != len(wantNames) {
		t.Fatalf("got %d tests, want %d", len(tests), len(wantNames))
	}
	for i, want := range wantNames {
		if tests[i].Name != want {
			t.Errorf("tests[%d].Name = %q, want %q", i, tests[i].Name, want)
		}
	}
}

func TestDiscover_HiddenDirSkipped(t *testing.T) {
	dir := t.TempDir()

	writeYAML(t, dir, "visible.yml", `
vis:
  condition: "visible"
`)
	writeYAML(t, dir, ".hidden/secret.yml", `
secret:
  condition: "hidden"
`)

	tests, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}
	if len(tests) != 1 || tests[0].Name != "vis" {
		t.Errorf("expected only 'vis' test, got %v", tests)
	}
}

func TestDiscover_MissingCondition(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "bad.yml", `
no_condition:
  on: ["*.go"]
`)

	_, err := Discover(dir)
	if err == nil {
		t.Fatal("expected error for missing condition")
	}
}

func TestDiscover_NonexistentDir(t *testing.T) {
	_, err := Discover("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestDiscover_DuplicateName(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "a.yml", `
dup:
  condition: "first definition"
`)
	writeYAML(t, dir, "b.yml", `
dup:
  condition: "second definition"
`)

	_, err := Discover(dir)
	if err == nil {
		t.Fatal("expected error for duplicate test name")
	}
}

func TestDiscover_OnField(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "with_on.yml", `
tracked:
  on: ["src/*.go", "**/*.md"]
  condition: "check tracked files"
`)

	tests, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}
	if len(tests) != 1 {
		t.Fatalf("got %d tests, want 1", len(tests))
	}
	if len(tests[0].On) != 2 {
		t.Errorf("tests[0].On = %v, want 2 globs", tests[0].On)
	}
}

func TestDiscover_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	tests, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}
	if len(tests) != 0 {
		t.Errorf("expected 0 tests from empty dir, got %d", len(tests))
	}
}

func TestDiscover_EmptyYAMLFile(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "empty.yml", "")

	tests, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}
	if len(tests) != 0 {
		t.Errorf("expected 0 tests from empty YAML, got %d", len(tests))
	}
}

func TestDiscover_CommentOnlyYAML(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "comments.yml", "# just a comment\n# nothing here\n")

	tests, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}
	if len(tests) != 0 {
		t.Errorf("expected 0 tests from comment-only YAML, got %d", len(tests))
	}
}

func TestDiscover_NestedSubdirectories(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "top.yml", `
top_test:
  condition: "top level"
`)
	writeYAML(t, dir, "sub/nested.yml", `
nested_test:
  condition: "nested"
`)
	writeYAML(t, dir, "sub/deep/deeper.yml", `
deep_test:
  condition: "deeply nested"
`)

	tests, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}
	if len(tests) != 3 {
		t.Fatalf("expected 3 tests, got %d", len(tests))
	}
	// Verify SourceFile tracks relative paths including subdirs
	sourceFiles := map[string]bool{}
	for _, tt := range tests {
		sourceFiles[tt.SourceFile] = true
	}
	for _, want := range []string{"top.yml", filepath.Join("sub", "nested.yml"), filepath.Join("sub", "deep", "deeper.yml")} {
		if !sourceFiles[want] {
			t.Errorf("expected SourceFile %q, got %v", want, sourceFiles)
		}
	}
}

func TestDiscover_InvalidYAMLSyntax(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "bad.yml", ":\n  :\n    invalid yaml {{{\n")

	_, err := Discover(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML syntax")
	}
}

func TestDiscover_DuplicateNameSameFile(t *testing.T) {
	dir := t.TempDir()
	// YAML spec: duplicate keys in same mapping -- last one wins in parsing,
	// but we use yaml.Node which preserves both keys
	writeYAML(t, dir, "dups.yml", `
test_a:
  condition: "first"
test_b:
  condition: "second"
test_a:
  condition: "duplicate"
`)

	_, err := Discover(dir)
	if err == nil {
		t.Fatal("expected error for duplicate test name within same file")
	}
}

func TestDiscover_NonYAMLFilesIgnored(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "tests.yml", `
real_test:
  condition: "real"
`)
	// Write non-YAML files that should be ignored
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not yaml"), 0o644)
	os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"key":"val"}`), 0o644)

	tests, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}
	if len(tests) != 1 {
		t.Errorf("expected 1 test (non-YAML files ignored), got %d", len(tests))
	}
}

func TestDiscover_PerTestOverrides(t *testing.T) {
	t.Run("all overrides", func(t *testing.T) {
		dir := t.TempDir()
		writeYAML(t, dir, "overrides.yml", `
full_override:
  condition: "check something"
  model: "claude-sonnet-4-20250514"
  timeout: 120
  max_iterations: 5
  tags: ["ci", "fast"]
`)
		tests, err := Discover(dir)
		if err != nil {
			t.Fatalf("Discover() error: %v", err)
		}
		if len(tests) != 1 {
			t.Fatalf("got %d tests, want 1", len(tests))
		}
		tt := tests[0]
		if tt.Model != "claude-sonnet-4-20250514" {
			t.Errorf("Model = %q, want %q", tt.Model, "claude-sonnet-4-20250514")
		}
		if tt.Timeout != 120 {
			t.Errorf("Timeout = %d, want 120", tt.Timeout)
		}
		if tt.MaxIterations != 5 {
			t.Errorf("MaxIterations = %d, want 5", tt.MaxIterations)
		}
		if len(tt.Tags) != 2 || tt.Tags[0] != "ci" || tt.Tags[1] != "fast" {
			t.Errorf("Tags = %v, want [ci fast]", tt.Tags)
		}
	})

	t.Run("partial overrides", func(t *testing.T) {
		dir := t.TempDir()
		writeYAML(t, dir, "partial.yml", `
partial:
  condition: "check partial"
  timeout: 60
`)
		tests, err := Discover(dir)
		if err != nil {
			t.Fatalf("Discover() error: %v", err)
		}
		tt := tests[0]
		if tt.Model != "" {
			t.Errorf("Model = %q, want empty", tt.Model)
		}
		if tt.Timeout != 60 {
			t.Errorf("Timeout = %d, want 60", tt.Timeout)
		}
		if tt.MaxIterations != 0 {
			t.Errorf("MaxIterations = %d, want 0", tt.MaxIterations)
		}
	})

	t.Run("no overrides", func(t *testing.T) {
		dir := t.TempDir()
		writeYAML(t, dir, "plain.yml", `
plain:
  condition: "just a condition"
`)
		tests, err := Discover(dir)
		if err != nil {
			t.Fatalf("Discover() error: %v", err)
		}
		tt := tests[0]
		if tt.Model != "" || tt.Timeout != 0 || tt.MaxIterations != 0 || len(tt.Tags) != 0 {
			t.Errorf("expected zero values, got model=%q timeout=%d maxIter=%d tags=%v",
				tt.Model, tt.Timeout, tt.MaxIterations, tt.Tags)
		}
	})
}
