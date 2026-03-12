// Package copier handles the actual file and directory copy operations.
// It is purely additive — it never deletes anything from the source.
package copier

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Result captures the outcome of a single copy operation.
type Result struct {
	App         string
	Label       string
	SourcePath  string
	TargetPath  string
	Status      string // "copied", "skipped", "failed"
	BytesCopied int64
	Detail      string
}

// MirrorDir recursively copies a source directory to a target directory.
// Existing files in the target are overwritten. Nothing is ever deleted.
func MirrorDir(src, dst string) (int64, error) {
	var totalBytes int64

	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Compute the relative path from source root
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("computing relative path: %w", err)
		}
		targetPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		n, err := copyFile(path, targetPath)
		if err != nil {
			return fmt.Errorf("copying %s: %w", relPath, err)
		}
		totalBytes += n

		return nil
	})

	return totalBytes, err
}

// CopyFile copies a single file from src to dst, creating parent directories
// as needed. Returns bytes written.
func CopyFile(src, dst string) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return 0, fmt.Errorf("creating target directory: %w", err)
	}
	return copyFile(src, dst)
}

// CopyPath copies src to dst. If src is a directory, MirrorDir is used.
// If src is a file, CopyFile is used. Returns bytes copied.
func CopyPath(src, dst string) (int64, error) {
	info, err := os.Stat(src)
	if err != nil {
		return 0, fmt.Errorf("stat source: %w", err)
	}
	if info.IsDir() {
		return MirrorDir(src, dst)
	}
	return CopyFile(src, dst)
}

// copyFile is the internal implementation that copies a single file.
func copyFile(src, dst string) (int64, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, fmt.Errorf("opening source: %w", err)
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return 0, fmt.Errorf("creating parent dirs: %w", err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, fmt.Errorf("creating target: %w", err)
	}
	defer dstFile.Close()

	n, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return n, fmt.Errorf("writing target: %w", err)
	}

	// Preserve file permissions
	info, err := os.Stat(src)
	if err == nil {
		_ = os.Chmod(dst, info.Mode())
	}

	return n, nil
}
