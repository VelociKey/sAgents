package natvs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindGitRoot looks for the .gitroot file starting from the current directory and moving up.
func FindGitRoot() (string, error) {
	curr, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(curr, ".gitroot")); err == nil {
			return curr, nil
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			return "", fmt.Errorf(".gitroot not found in ancestry")
		}
		curr = parent
	}
}

// ResolvePaths ensures a list of paths are relative to the discovered gitRoot.
// This is critical for Jules CLI context ingestion.
func ResolvePaths(gitRoot string, paths []string) []string {
	var resolved []string
	absRoot, _ := filepath.Abs(gitRoot)

	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			resolved = append(resolved, p)
			continue
		}

		if strings.HasPrefix(absPath, absRoot) {
			rel, err := filepath.Rel(absRoot, absPath)
			if err == nil {
				resolved = append(resolved, rel)
				continue
			}
		}
		resolved = append(resolved, p)
	}
	return resolved
}
