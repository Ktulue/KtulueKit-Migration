// Package mapper resolves environment variables and validates paths
// for source/target mapping during migration.
package mapper

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// envVarPattern matches Windows-style environment variables like %APPDATA%.
var envVarPattern = regexp.MustCompile(`%([^%]+)%`)

// ResolvePath expands environment variables in a path string and returns
// the fully resolved absolute path.
//
// Supports Windows-style %VAR% syntax (e.g., %APPDATA%, %USERPROFILE%)
// as well as Go-native os.ExpandEnv for $VAR and ${VAR} syntax.
func ResolvePath(raw string) string {
	// First pass: expand %VAR% style
	resolved := envVarPattern.ReplaceAllStringFunc(raw, func(match string) string {
		varName := strings.Trim(match, "%")
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match // leave unexpanded if not found
	})

	// Second pass: expand $VAR style
	resolved = os.ExpandEnv(resolved)

	// Normalize path separators
	resolved = filepath.FromSlash(resolved)

	return resolved
}

// BuildSourcePath constructs the full source path from the backup root and
// the item's relative source path.
func BuildSourcePath(backupRoot, source string) string {
	resolved := ResolvePath(source)
	if filepath.IsAbs(resolved) {
		return resolved
	}
	return filepath.Join(ResolvePath(backupRoot), resolved)
}

// BuildTargetPath resolves the target path, expanding any environment variables.
func BuildTargetPath(target string) string {
	return ResolvePath(target)
}

// ValidateSourceExists checks that the source path exists and returns
// an error with context if it doesn't.
func ValidateSourceExists(sourcePath, appName, label string) error {
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source not found for %s — %s: %s", appName, label, sourcePath)
	}
	return nil
}
