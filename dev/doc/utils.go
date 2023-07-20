package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func unique[T comparable](slice []T) []T {
	// Create a map to store unique elements
	seen := make(map[T]bool)
	result := []T{}

	// Loop through the slice, adding elements to the map if they haven't been seen before
	for _, val := range slice {
		if _, ok := seen[val]; !ok {
			seen[val] = true
			result = append(result, val)
		}
	}
	return result
}

// detectImportPath return the import path of a given directory (assumed as a module)
// by looking at the closest go.mod file (that contains "module <NAME>") and prepending
// the path from root to that folder: <NAME>/path/to/dir
func detectImportPath(dir string, base string, levels int) (string, error) {
	d := filepath.Clean(dir)

	if levels < 0 {
		return "", fmt.Errorf("directory depth exceeded")
	}
	entries, err := os.ReadDir(d)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Name() == "go.mod" {
			file, err := os.Open(path.Join(dir, "go.mod")) //#nosec G304 -- Only the 'module ...' line will be parsed
			if err != nil {
				return "", err
			}
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "module") {
					words := strings.SplitN(line, " ", 2)
					return filepath.Join(words[1], base), nil
				}
			}
		}
	}
	// no entry found
	b := filepath.Base(d)
	return detectImportPath(filepath.Dir(d), filepath.Join(b, base), levels-1)
}
