package config

import (
	"os"
	"testing"
)

func TestLoadLabels_NonexistentFile(t *testing.T) {
	t.Setenv("WT_CONFIG_DIR", t.TempDir())

	labels, err := LoadLabels()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(labels) != 0 {
		t.Fatalf("expected empty map, got %v", labels)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	t.Setenv("WT_CONFIG_DIR", t.TempDir())

	want := Labels{
		"/home/user/project-a":         "fixing login bug",
		"/home/user/project-b-feature": "new API endpoint",
	}

	if err := SaveLabels(want); err != nil {
		t.Fatalf("save error: %v", err)
	}

	got, err := LoadLabels()
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d labels, got %d", len(want), len(got))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("label %q: expected %q, got %q", k, v, got[k])
		}
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("WT_CONFIG_DIR", dir+"/nested/config")

	if err := SaveLabels(Labels{"key": "val"}); err != nil {
		t.Fatalf("save error: %v", err)
	}

	// Verify the file exists.
	if _, err := os.Stat(dir + "/nested/config/labels.json"); err != nil {
		t.Fatalf("expected labels.json to exist: %v", err)
	}
}
