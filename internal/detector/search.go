package detector

import (
	"os"
	"path/filepath"
	"strings"
)

// SearchForApp scans common directories across all available drive letters
// for a folder matching target. If executable is non-empty, it verifies the
// folder contains that file. Returns all matching paths.
func SearchForApp(searchPaths []string, target string, executable string) []string {
	roots := listDriveRoots()
	var absolutePaths []string
	for _, root := range roots {
		for _, sp := range searchPaths {
			absolutePaths = append(absolutePaths, filepath.Join(root, filepath.FromSlash(sp)))
		}
	}
	return SearchForAppInRoots(absolutePaths, target, executable)
}

// SearchForAppInRoots scans the given absolute directories for a folder
// matching target. If executable is non-empty, it verifies the folder
// contains that file. Returns all matching paths.
func SearchForAppInRoots(roots []string, target string, executable string) []string {
	var candidates []string
	targetLower := strings.ToLower(target)

	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			if strings.ToLower(e.Name()) != targetLower {
				continue
			}
			candidate := filepath.Join(root, e.Name())
			if executable != "" {
				exePath := filepath.Join(candidate, executable)
				if _, err := os.Stat(exePath); err != nil {
					continue
				}
			}
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

// listDriveRoots returns available Windows drive roots (C:\, D:\, etc.).
func listDriveRoots() []string {
	var roots []string
	for letter := 'A'; letter <= 'Z'; letter++ {
		root := string(letter) + ":\\"
		if _, err := os.Stat(root); err == nil {
			roots = append(roots, root)
		}
	}
	return roots
}
