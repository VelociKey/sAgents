package natvs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestFindGitRoot(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "natvs-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	gitRootFile := filepath.Join(tempDir, ".gitroot")
	if err := os.WriteFile(gitRootFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	deepDir := filepath.Join(tempDir, "a", "b", "c")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatal(err)
	}

	oldWd, _ := os.Getwd()
	if err := os.Chdir(deepDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	root, err := FindGitRoot()
	if err != nil {
		t.Errorf("FindGitRoot failed: %v", err)
	}

	expected, _ := filepath.Abs(tempDir)
	actual, _ := filepath.Abs(root)
	if actual != expected {
		t.Errorf("Expected root %s, got %s", expected, actual)
	}
}

func TestResolvePaths(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "resolve-test-*")
	defer os.RemoveAll(tempDir)

	// Test with N-Silos to demonstrate dynamic nature
	numSilos := 5
	var paths []string
	var expected []string

	for i := 1; i <= numSilos; i++ {
		name := fmt.Sprintf("Silo-%d", i)
		fullPath := filepath.Join(tempDir, name)
		os.Mkdir(fullPath, 0755)
		
		paths = append(paths, fullPath)
		expected = append(expected, name)
	}

	resolved := ResolvePaths(tempDir, paths)

	if len(resolved) != len(paths) {
		t.Fatalf("Expected %d resolved paths, got %d", len(paths), len(resolved))
	}

	for i, v := range resolved {
		v = filepath.ToSlash(v)
		if v != expected[i] {
			t.Errorf("Expected path %s, got %s", expected[i], v)
		}
	}
}
