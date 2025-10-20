package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// KeepLeaves returns only the most precise paths
// of a input list of paths.
// For instance if the list is the following:
// "/a/b/c", "/a/b", "/b/d", "/b"
// It returns "/a/b/c", "/b/d"
// Note: the paths are sorted in the output
func KeepLeaves(files []string) []string {
	n := len(files)
	if n <= 1 {
		return files
	}
	// sort the files

	sort.Strings(files)

	base := files[n-1]
	baseDir := filepath.Dir(base)
	out := []string{base}

	// for i := n - 2; i >= 0; {
	for i := n - 2; i >= 0; i-- {
		for ; i >= 0 && strings.HasPrefix(base, files[i]) && filepath.Dir(files[i]) != baseDir; i-- {
		}
		if i < 0 {
			return out
		}
		out = append(out, files[i])
		base = files[i] // new base
		baseDir = filepath.Dir(base)
	}
	return out
}

// GetLines reads a file and returns all its lines
// You can pass callbacks to modify the content of the line.
// If you pass multiple callacks, the first is called on the line, then
// the second is called on the result of the first one and so one
// ex: callback2(callback1(callback0(line)))
func GetLines(file string, callbacks ...func(string) string) ([]string, error) {
	f, err := os.Open(file) // #nosec G304 -- GetLines is a helper function
	if err != nil {
		return nil, err
	}
	// Create new Scanner.
	scanner := bufio.NewScanner(f)
	result := make([]string, 0)
	// Use Scan.
	for scanner.Scan() {
		line := scanner.Text()
		// Append line to result.
		for _, callback := range callbacks {
			line = callback(line)
		}
		result = append(result, line)
	}
	return result, nil
}
