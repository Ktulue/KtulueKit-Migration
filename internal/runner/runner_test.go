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
