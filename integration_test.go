// Package main integration tests exercise the full migration pipeline
// against a synthetic backup directory — no real backup drive needed.
package main_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ktulue/KtulueKit-Migration/internal/config"
	"github.com/Ktulue/KtulueKit-Migration/internal/reporter"
	"github.com/Ktulue/KtulueKit-Migration/internal/runner"
)

// fakeBackup creates a realistic fake backup tree and returns its root path.
// Mirrors the structure of D:\KtulueBackup for the apps in the config.
func fakeBackup(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	dirs := []string{
		"obs-studio/basic/scenes",
		"obs-studio/basic/profiles",
		"obs-studio/plugin_config",
		"streamerbot/data",
		"streamerbot/settings",
		"streamdeck/ProfilesV2",
		"brave/User Data/Default",
		"media/stream-assets",
		"powershell",
		"lurkbait/CustomCatches",
		"git",
		"ssh",
		"vscode/User",
		"gimp",
		"sharex",
		"discord",
		"personal/Desktop",
		"personal/Documents/StreamScripts",
		"personal/Documents/Notes",
	}
	for _, d := range dirs {
		_ = os.MkdirAll(filepath.Join(root, d), 0755)
	}

	files := map[string]string{
		"obs-studio/basic/scenes/main.json":              `{"scene":"main"}`,
		"obs-studio/basic/profiles/default/basic.ini":   "[Profile]",
		"obs-studio/plugin_config/obs-browser/global.ini": "[Browser]",
		"streamerbot/data/actions.db":                    "SQLite data",
		"streamerbot/settings/settings.json":             `{"version":1}`,
		"streamdeck/ProfilesV2/default.sdProfile":       "profile",
		"brave/User Data/Default/Bookmarks":              `{"roots":{}}`,
		"media/stream-assets/alert.gif":                  "GIF89a",
		"powershell/Microsoft.PowerShell_profile.ps1":   "oh-my-posh init",
		"lurkbait/CustomCatches/fish1.png":               "PNG",
		"lurkbait/PlayerData.txt":                        "player",
		"lurkbait/CatchData.txt":                         "catches",
		"git/.gitconfig":                                 "[user]\n\tname=Ktulue",
		"ssh/id_ed25519":                                 "ssh-key",
		"ssh/config":                                     "Host github.com",
		"vscode/User/settings.json":                      `{"editor.fontSize":14}`,
		"gimp/gimprc":                                    "()",
		"sharex/ApplicationConfig.json":                  `{}`,
		"discord/settings.json":                          `{"SKIP_HOST_UPDATE":true}`,
		"personal/Desktop/shortcut.lnk":                 "link",
		"personal/Documents/StreamScripts/main.js":      "// script",
		"personal/Documents/Notes/ideas.txt":            "ideas",
	}
	for path, content := range files {
		full := filepath.Join(root, path)
		_ = os.MkdirAll(filepath.Dir(full), 0755)
		_ = os.WriteFile(full, []byte(content), 0644)
	}

	return root
}

// syntheticConfig builds a config using the fake backup root and temp target dirs
// so no real system paths are touched during copy operations.
func syntheticConfig(t *testing.T, backupRoot string) *config.Config {
	t.Helper()
	dst := t.TempDir()

	return &config.Config{
		Version:    "1.0",
		BackupRoot: backupRoot,
		Apps: []config.App{
			{Name: "OBS Studio", Category: "Streaming", Items: []config.Item{
				{Label: "scenes & profiles", Source: "obs-studio/basic", Target: filepath.Join(dst, "obs/basic")},
				{Label: "plugin configs", Source: "obs-studio/plugin_config", Target: filepath.Join(dst, "obs/plugin_config")},
			}},
			{Name: "Streamer.bot", Category: "Streaming", Items: []config.Item{
				{Label: "actions & commands", Source: "streamerbot/data", Target: filepath.Join(dst, "streamerbot/data")},
				{Label: "settings", Source: "streamerbot/settings", Target: filepath.Join(dst, "streamerbot/settings")},
			}},
			{Name: "Stream Deck", Category: "Streaming", Items: []config.Item{
				{Label: "profiles & icons", Source: "streamdeck/ProfilesV2", Target: filepath.Join(dst, "streamdeck")},
			}},
			{Name: "Brave Browser", Category: "Browser & Identity", Items: []config.Item{
				{Label: "user profile", Source: "brave/User Data/Default", Target: filepath.Join(dst, "brave/Default")},
			}},
			{Name: "GIFs & Media Assets", Category: "Media Assets", Items: []config.Item{
				{Label: "stream media", Source: "media/stream-assets", Target: filepath.Join(dst, "stream-assets")},
			}},
			{Name: "PowerShell", Category: "Shell & Terminal", Items: []config.Item{
				{Label: "profile (oh-my-posh)", Source: "powershell/Microsoft.PowerShell_profile.ps1", Target: filepath.Join(dst, "powershell/profile.ps1"), Strategy: "file"},
			}},
			{Name: "LurkBait Twitch Fishing", Category: "Games", Items: []config.Item{
				{Label: "save data & custom catches", Source: "lurkbait", Target: filepath.Join(dst, "lurkbait")},
			}},
			{Name: "Git", Category: "Dev Tools", Items: []config.Item{
				{Label: "global config", Source: "git/.gitconfig", Target: filepath.Join(dst, ".gitconfig"), Strategy: "file"},
			}},
			{Name: "SSH", Category: "Dev Tools", Items: []config.Item{
				{Label: "keys & config", Source: "ssh", Target: filepath.Join(dst, ".ssh")},
			}},
			{Name: "VS Code", Category: "Dev Tools", Items: []config.Item{
				{Label: "settings & keybindings", Source: "vscode/User", Target: filepath.Join(dst, "vscode/User")},
			}},
			{Name: "ShareX", Category: "Utilities", Items: []config.Item{
				{Label: "settings & screenshots", Source: "sharex", Target: filepath.Join(dst, "sharex")},
			}},
			{Name: "Discord", Category: "Communication", Items: []config.Item{
				{Label: "local settings", Source: "discord", Target: filepath.Join(dst, "discord")},
			}},
			{Name: "Personal Files", Category: "Personal Files", Items: []config.Item{
				{Label: "Desktop", Source: "personal/Desktop", Target: filepath.Join(dst, "Desktop"), Strategy: "mirror"},
				{Label: "Documents", Source: "personal/Documents", Target: filepath.Join(dst, "Documents"), Strategy: "selective"},
			}},
			// Source intentionally absent — should be skipped
			{Name: "Missing App", Category: "Utilities", Items: []config.Item{
				{Label: "settings", Source: "doesnotexist/settings", Target: filepath.Join(dst, "missing")},
			}},
		},
	}
}

func allIDs(cfg *config.Config) []string {
	var ids []string
	for _, app := range cfg.Apps {
		for _, item := range app.Items {
			ids = append(ids, app.Name+":"+item.Label)
		}
	}
	return ids
}

// ────────────────────────────────────────────────────────────────────────────
// Test 1: Dry-run — no files written anywhere
// ────────────────────────────────────────────────────────────────────────────

func TestIntegration_DryRun_NoFilesWritten(t *testing.T) {
	backup := fakeBackup(t)
	cfg := syntheticConfig(t, backup)

	rep := reporter.NewNull()
	r := runner.New(cfg, rep)
	r.SetSelectedIDs(allIDs(cfg))
	r.SetDryRun(true)
	result := r.Run()

	// Verify no target directories were created
	for _, item := range result.Items {
		if item.Status == reporter.StatusCopied {
			if _, err := os.Stat(item.TargetPath); err == nil {
				t.Errorf("dry-run: target path was created for %s — %s: %s", item.App, item.Label, item.TargetPath)
			}
		}
	}

	// Verify all items with present sources were estimated (not zero bytes for dirs)
	copied := 0
	skipped := 0
	for _, item := range result.Items {
		switch item.Status {
		case reporter.StatusCopied:
			copied++
		case reporter.StatusSkipped:
			skipped++
		}
	}

	t.Logf("Dry-run results: %d estimated, %d skipped (missing source)", copied, skipped)

	if copied == 0 {
		t.Error("expected at least some items to show estimated copies in dry-run")
	}
	if skipped == 0 {
		t.Error("expected at least one item skipped (Missing App has no source)")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Test 2: Real copy — files land in correct locations
// ────────────────────────────────────────────────────────────────────────────

func TestIntegration_RealCopy_FilesLandCorrectly(t *testing.T) {
	backup := fakeBackup(t)
	cfg := syntheticConfig(t, backup)

	logDir := t.TempDir()
	rep := reporter.New(logDir)
	r := runner.New(cfg, rep)

	// Select everything except selective items (those require picker)
	var ids []string
	for _, app := range cfg.Apps {
		for _, item := range app.Items {
			if item.Strategy != "selective" {
				ids = append(ids, app.Name+":"+item.Label)
			}
		}
	}
	r.SetSelectedIDs(ids)
	result := r.Run()

	// Spot-check key files were copied
	checks := map[string]string{
		"OBS scenes json":    filepath.Join(cfg.Apps[0].Items[0].Target, "scenes", "main.json"),
		"Streamer.bot db":    filepath.Join(cfg.Apps[1].Items[0].Target, "actions.db"),
		"LurkBait catches":   filepath.Join(cfg.Apps[6].Items[0].Target, "CustomCatches", "fish1.png"),
		"LurkBait PlayerData": filepath.Join(cfg.Apps[6].Items[0].Target, "PlayerData.txt"),
		"Git config":         cfg.Apps[7].Items[0].Target,
		"SSH key":            filepath.Join(cfg.Apps[8].Items[0].Target, "id_ed25519"),
		"Desktop shortcut":   filepath.Join(cfg.Apps[12].Items[0].Target, "shortcut.lnk"),
	}
	for name, path := range checks {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("missing after copy — %s: %s", name, path)
		}
	}

	// Log file should exist
	if rep.LogPath() == "" {
		t.Error("expected log file to be created")
	}
	if _, err := os.Stat(rep.LogPath()); err != nil {
		t.Errorf("log file not found at %s", rep.LogPath())
	}

	// Count outcomes
	copied, skipped, failed := 0, 0, 0
	for _, item := range result.Items {
		switch item.Status {
		case reporter.StatusCopied:
			copied++
		case reporter.StatusSkipped:
			skipped++
		case reporter.StatusFailed:
			failed++
		}
	}
	t.Logf("Copy results: %d copied, %d skipped, %d failed", copied, skipped, failed)

	if failed > 0 {
		t.Errorf("expected 0 failures, got %d", failed)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Test 3: Selective with picker selections — only chosen paths copy
// ────────────────────────────────────────────────────────────────────────────

func TestIntegration_Selective_OnlyChosenPathsCopied(t *testing.T) {
	backup := fakeBackup(t)
	cfg := syntheticConfig(t, backup)

	// Personal Files — Documents is the last app, second item (index 13, item 1)
	var docsCfg *config.Item
	var docsApp *config.App
	for i := range cfg.Apps {
		for j := range cfg.Apps[i].Items {
			if cfg.Apps[i].Items[j].Strategy == "selective" {
				docsApp = &cfg.Apps[i]
				docsCfg = &cfg.Apps[i].Items[j]
			}
		}
	}
	if docsCfg == nil {
		t.Fatal("no selective item found in config")
	}

	docsID := docsApp.Name + ":" + docsCfg.Label
	keepPath := filepath.Join(backup, "personal/Documents/StreamScripts")

	rep := reporter.NewNull()
	r := runner.New(cfg, rep)
	r.SetSelectedIDs([]string{docsID})
	r.SetSelectivePaths(map[string][]string{
		docsID: {keepPath},
	})
	r.Run()

	// StreamScripts should be copied
	if _, err := os.Stat(filepath.Join(docsCfg.Target, "StreamScripts", "main.js")); err != nil {
		t.Error("StreamScripts/main.js should have been copied")
	}
	// Notes should NOT be copied
	if _, err := os.Stat(filepath.Join(docsCfg.Target, "Notes")); err == nil {
		t.Error("Notes/ should NOT have been copied")
	}

	t.Logf("Selective copy: StreamScripts copied, Notes skipped ✓")
}

// ────────────────────────────────────────────────────────────────────────────
// Test 4: Selective with no picker selection — reported as skipped
// ────────────────────────────────────────────────────────────────────────────

func TestIntegration_Selective_NoPicker_IsSkipped(t *testing.T) {
	backup := fakeBackup(t)
	cfg := syntheticConfig(t, backup)

	var docsID string
	for _, app := range cfg.Apps {
		for _, item := range app.Items {
			if item.Strategy == "selective" {
				docsID = app.Name + ":" + item.Label
			}
		}
	}

	rep := reporter.NewNull()
	r := runner.New(cfg, rep)
	r.SetSelectedIDs([]string{docsID})
	// No SetSelectivePaths — simulates user checking item but never opening picker
	result := r.Run()

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Items))
	}
	if result.Items[0].Status != reporter.StatusSkipped {
		t.Errorf("expected skipped, got %s", result.Items[0].Status)
	}
	t.Logf("Selective with no picker: correctly reported as skipped ✓")
}

// ────────────────────────────────────────────────────────────────────────────
// Test 5: WriteManifest — output is valid JSON with correct contract shape
// ────────────────────────────────────────────────────────────────────────────

func TestIntegration_WriteManifest_ValidContract(t *testing.T) {
	backup := fakeBackup(t)
	cfg := syntheticConfig(t, backup)

	logDir := t.TempDir()
	rep := reporter.New(logDir)
	r := runner.New(cfg, rep)

	// Run OBS only (mirror, source present)
	r.SetSelectedIDs([]string{"OBS Studio:scenes & profiles"})
	r.Run()

	manifestPath := filepath.Join(logDir, "manifest.json")
	if err := rep.WriteManifest(manifestPath); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	var m struct {
		Version string `json:"version"`
		RunAt   string `json:"runAt"`
		Items   []struct {
			App           string   `json:"app"`
			Label         string   `json:"label"`
			SourcePath    string   `json:"sourcePath"`
			TargetPath    string   `json:"targetPath"`
			Status        string   `json:"status"`
			BytesCopied   int64    `json:"bytesCopied"`
			SelectedPaths []string `json:"selectedPaths"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("manifest is not valid JSON: %v\nContent: %s", err, data)
	}

	if m.Version != "1.0" {
		t.Errorf("expected version 1.0, got %q", m.Version)
	}
	if m.RunAt == "" {
		t.Error("runAt should not be empty")
	}
	if len(m.Items) == 0 {
		t.Error("expected at least one manifest item")
	}
	item := m.Items[0]
	if item.App == "" || item.Label == "" || item.SourcePath == "" || item.TargetPath == "" {
		t.Errorf("manifest item missing required fields: %+v", item)
	}
	if item.SelectedPaths == nil {
		t.Error("selectedPaths should be [] not null")
	}

	t.Logf("Manifest contract ✓  version=%s  runAt=%s  items=%d  bytesCopied=%d",
		m.Version, m.RunAt, len(m.Items), item.BytesCopied)
}
