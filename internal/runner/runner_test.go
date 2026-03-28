package runner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Ktulue/KtulueKit-Migration/internal/config"
	"github.com/Ktulue/KtulueKit-Migration/internal/reporter"
	"github.com/Ktulue/KtulueKit-Migration/internal/runner"
)

func makeTestConfig(t *testing.T, backupRoot string) *config.Config {
	t.Helper()
	return &config.Config{
		Version:    "1.0",
		BackupRoot: backupRoot,
		Apps: []config.App{
			{
				Name: "TestApp", Category: "Test",
				Items: []config.Item{
					{Label: "mirror item", Source: "mirror", Target: filepath.Join(t.TempDir(), "dst-mirror"), Strategy: "mirror"},
					{Label: "selective item", Source: "selective", Target: filepath.Join(t.TempDir(), "dst-selective"), Strategy: "selective"},
				},
			},
		},
	}
}

func TestRunner_SelectiveStrategy(t *testing.T) {
	tmp := t.TempDir()

	// Create source files
	selectiveDir := filepath.Join(tmp, "selective")
	_ = os.MkdirAll(selectiveDir, 0755)
	_ = os.WriteFile(filepath.Join(selectiveDir, "keep.txt"), []byte("keep"), 0644)
	_ = os.WriteFile(filepath.Join(selectiveDir, "skip.txt"), []byte("skip"), 0644)

	cfg := makeTestConfig(t, tmp)
	dstDir := cfg.Apps[0].Items[1].Target

	rep := reporter.NewNull()
	r := runner.New(cfg, rep)
	r.SetSelectedIDs([]string{"TestApp:selective item"})
	r.SetSelectivePaths(map[string][]string{
		"TestApp:selective item": {filepath.Join(selectiveDir, "keep.txt")},
	})

	r.Run()

	if _, err := os.Stat(filepath.Join(dstDir, "keep.txt")); err != nil {
		t.Error("keep.txt should have been copied")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "skip.txt")); err == nil {
		t.Error("skip.txt should NOT have been copied")
	}
}

func TestRunner_UsesDestPathMap(t *testing.T) {
	tmp := t.TempDir()

	// Create source files
	srcDir := filepath.Join(tmp, "mirror")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("hello"), 0644)

	// Custom destination from detection
	customDest := filepath.Join(t.TempDir(), "custom-dest")

	cfg := makeTestConfig(t, tmp)
	rep := reporter.NewNull()
	r := runner.New(cfg, rep)
	r.SetSelectedIDs([]string{"TestApp:mirror item"})
	r.SetDestPathMap(map[string]string{
		"TestApp:mirror item": customDest,
	})

	result := r.Run()

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Items))
	}
	if result.Items[0].Status != reporter.StatusCopied {
		t.Errorf("expected copied, got %s", result.Items[0].Status)
	}
	if result.Items[0].TargetPath != customDest {
		t.Errorf("TargetPath = %q, want %q", result.Items[0].TargetPath, customDest)
	}

	// Verify file landed in custom destination
	if _, err := os.Stat(filepath.Join(customDest, "file.txt")); err != nil {
		t.Error("file.txt should have been copied to custom destination")
	}
}

func TestRunner_DryRun_NoFilesWritten(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "mirror")
	_ = os.MkdirAll(srcDir, 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("data"), 0644)

	cfg := makeTestConfig(t, tmp)
	dstDir := cfg.Apps[0].Items[0].Target

	rep := reporter.NewNull()
	r := runner.New(cfg, rep)
	r.SetSelectedIDs([]string{"TestApp:mirror item"})
	r.SetDryRun(true)

	r.Run()

	if _, err := os.Stat(dstDir); err == nil {
		t.Error("dry-run should not create any destination files")
	}
}
