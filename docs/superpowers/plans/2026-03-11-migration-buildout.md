# KtulueKit-Migration Build-Out Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build KtulueKit-Migration from working scaffold to a shippable, fully-configured, polished desktop tool with folder picker, dry-run mode, and a locked manifest contract for KtulueKit-Cleanup.

**Architecture:** Go backend (Wails-bound App struct) + Svelte frontend communicating via Wails event bridge. Internal packages (config, runner, copier, mapper, reporter) handle all file operations independently of the GUI layer, making them unit-testable. The runner is the central orchestrator; the reporter owns all output (log + manifest JSON).

**Tech Stack:** Go 1.25, Wails v2.11, Svelte 4, Vite 5. Tests use Go's standard `testing` package. Run tests with `go test ./internal/...` from the project root.

---

## File Map

### Files to Modify
| File | What changes |
|------|-------------|
| `types.go` | Add `Strategy` to `ItemView`; add `SelectedPaths` to `ManifestEntry`; update `categoryOrder` |
| `app.go` | Add `ListFolder` method; update `StartMigration` signature (add `selectivePaths`, `dryRun`); call `WriteManifest` after run |
| `internal/config/config.go` | Update Item.Strategy comment to retire `"glob"`, add `"selective"` |
| `internal/runner/runner.go` | Add selective copy strategy; add dry-run mode; add `SelectedPaths` to `RunResultItem` |
| `internal/copier/copier.go` | Add exported `CopyPath` (handles file or dir) |
| `internal/reporter/reporter.go` | Add `NewNull()`; add `WriteManifest(path string) error` |
| `schema/ktuluekit-migration.schema.json` | Add `"selective"` to strategy enum; retire `"glob"`; add new category values |
| `ktuluekit-migration.json` | Add all new app entries (Dev Tools, Creative Suite, Utilities, Communication, Personal Files) |
| `frontend/src/App.svelte` | Add `selectivePaths` state; add `dryRun` state; update `handleStartMigration` |
| `frontend/src/screens/SelectionScreen.svelte` | Add dry-run toggle |
| `frontend/src/components/ItemRow.svelte` | Accept `strategy` prop; trigger picker on `selective` check |

### Files to Create
| File | Purpose |
|------|---------|
| `frontend/src/components/FolderPicker.svelte` | Modal folder browser for selective items |
| `internal/runner/runner_test.go` | Tests for selective strategy and dry-run |
| `internal/reporter/reporter_test.go` | Tests for `WriteManifest` and `NewNull` |
| `internal/copier/copier_test.go` | Tests for `CopyPath` |

---

## Chunk 1: Phase 1 — Dev Run & Wiring Verification

### Task 1: Verify the app compiles and opens

**Files:** none modified — verification only

- [ ] **Step 1: Run wails dev from project root**

```bash
cd "F:/GDriveClone/Claude_Code/KtulueKit-Migration"
wails dev
```

Expected: Window opens titled "KtulueKit Migration" showing SelectionScreen with categories loaded from config. Check browser console (F12 in dev window) for errors.

- [ ] **Step 2: Verify Wails JS bindings were generated**

Check that this directory exists and contains Go method stubs:
```
frontend/wailsjs/go/main/App.js
frontend/wailsjs/runtime/runtime.js
```

If missing, `wails dev` should have created them. If not, run:
```bash
wails generate module
```

- [ ] **Step 3: Verify GetConfig loads correctly**

In the running app, SelectionScreen should show these categories:
- Streaming (OBS Studio, Streamer.bot, Stream Deck)
- Browser & Identity (Brave Browser)
- Media Assets (GIFs & Media Assets)
- Shell & Terminal (PowerShell)
- Games (LurkBait Twitch Fishing)

Profile dropdown should show: Full Restore, Streaming Only, Browser & Identity

- [ ] **Step 4: Run a migration with no backup present**

Select any item, click Start Migration. Expected: ProgressScreen shows, each item gets "skipped" status (source not found), lands on SummaryScreen. No crash.

- [ ] **Step 5: Verify log file created**

Check that `logs/migration_<timestamp>.log` was created and contains skip entries.

- [ ] **Step 6: Commit verified working state**

Run `git status` first to see what was generated. If `wails dev` regenerated the Wails JS bindings (`frontend/wailsjs/`), commit those explicitly:

```bash
git add frontend/wailsjs/
git commit -m "chore: verify wails dev run — all screens functional, deps installed"
```

If `git status` is clean (nothing changed), no commit needed — Phase 1 is verification only.

---

## Chunk 2: Phase 2 — Config Expansion

### Task 2: Update schema for new categories and selective strategy

**Files:**
- Modify: `schema/ktuluekit-migration.schema.json`

- [ ] **Step 1: Read the current schema**

```bash
cat "F:/GDriveClone/Claude_Code/KtulueKit-Migration/schema/ktuluekit-migration.schema.json"
```

- [ ] **Step 2: Update strategy enum — retire "glob", add "selective"**

Find the `strategy` property in the schema (line 87). Replace:
```json
"enum": ["mirror", "file", "glob"]
```
With:
```json
"enum": ["mirror", "file", "selective"]
```
Also update the `description` field on strategy to remove mention of "glob".

- [ ] **Step 3: Add enum to category property in schema**

The `category` property currently uses `"examples"` not `"enum"` — there is no existing enum to update. Add one:

```json
"category": {
  "type": "string",
  "description": "Grouping category for the UI.",
  "enum": [
    "Streaming",
    "Dev Tools",
    "Creative Suite",
    "Utilities",
    "Browser & Identity",
    "Communication",
    "Media Assets",
    "Shell & Terminal",
    "Games",
    "Personal Files"
  ]
}
```

Remove the `"examples"` array from `category` when adding the `"enum"`.

- [ ] **Step 4: Update categoryOrder in types.go**

In `types.go`, replace the `categoryOrder` slice:
```go
var categoryOrder = []string{
	"Streaming",
	"Dev Tools",
	"Creative Suite",
	"Utilities",
	"Browser & Identity",
	"Communication",
	"Media Assets",
	"Shell & Terminal",
	"Games",
	"Personal Files",
}
```

- [ ] **Step 5: Update strategy comment in config.go**

In `internal/config/config.go`, update the `Strategy` field comment:
```go
Strategy    string `json:"strategy,omitempty"` // "mirror" (default) | "file" | "selective"
```

- [ ] **Step 6: Commit schema and category changes**

```bash
git add schema/ktuluekit-migration.schema.json types.go internal/config/config.go
git commit -m "feat: add selective strategy and new categories to schema"
```

---

### Task 3: Add all new apps to ktuluekit-migration.json

**Files:**
- Modify: `ktuluekit-migration.json`

- [ ] **Step 1: Add Dev Tools apps**

Add after the existing apps array entries:

```json
{
  "name": "Git",
  "category": "Dev Tools",
  "items": [
    {
      "label": "global config",
      "source": "git/.gitconfig",
      "target": "%USERPROFILE%/.gitconfig",
      "description": "Global git identity, aliases, and default branch settings"
    }
  ]
},
{
  "name": "SSH",
  "category": "Dev Tools",
  "items": [
    {
      "label": "keys & config",
      "source": "ssh",
      "target": "%USERPROFILE%/.ssh",
      "description": "SSH keys and config for GitHub and remote connections",
      "notes": "Restore file permissions after copy: id_rsa should be 600"
    }
  ]
},
{
  "name": "VS Code",
  "category": "Dev Tools",
  "items": [
    {
      "label": "settings & keybindings",
      "source": "vscode/User",
      "target": "%APPDATA%/Code/User",
      "description": "VS Code settings.json, keybindings.json, and snippets"
    }
  ]
},
{
  "name": "Windows Terminal",
  "category": "Dev Tools",
  "items": [
    {
      "label": "settings",
      "source": "terminal/settings.json",
      "target": "%LOCALAPPDATA%/Packages/Microsoft.WindowsTerminal_8wekyb3d8bbwe/LocalState/settings.json",
      "description": "Terminal profiles, color schemes, and keybindings",
      "notes": "Applies only when Terminal is installed via Microsoft Store or WinGet"
    }
  ]
}
```

- [ ] **Step 2: Add Creative Suite apps**

```json
{
  "name": "GIMP",
  "category": "Creative Suite",
  "items": [
    {
      "label": "brushes & settings",
      "source": "gimp",
      "target": "%APPDATA%/GIMP",
      "description": "GIMP user settings, brushes, and presets"
    }
  ]
},
{
  "name": "Krita",
  "category": "Creative Suite",
  "items": [
    {
      "label": "brushes & settings",
      "source": "krita",
      "target": "%APPDATA%/krita",
      "description": "Krita user settings and brush presets"
    }
  ]
},
{
  "name": "Audacity",
  "category": "Creative Suite",
  "items": [
    {
      "label": "settings",
      "source": "audacity",
      "target": "%APPDATA%/audacity",
      "description": "Audacity preferences and EQ presets"
    }
  ]
},
{
  "name": "Blender",
  "category": "Creative Suite",
  "items": [
    {
      "label": "preferences & addons",
      "source": "blender",
      "target": "%APPDATA%/Blender Foundation",
      "description": "Blender user preferences, themes, and installed addons"
    }
  ]
},
{
  "name": "Aseprite",
  "category": "Creative Suite",
  "items": [
    {
      "label": "settings",
      "source": "aseprite",
      "target": "%APPDATA%/Aseprite",
      "description": "Aseprite user settings and palettes"
    }
  ]
},
{
  "name": "Kdenlive",
  "category": "Creative Suite",
  "items": [
    {
      "label": "settings",
      "source": "kdenlive",
      "target": "%APPDATA%/kdenlive",
      "description": "Kdenlive user settings and render profiles"
    }
  ]
},
{
  "name": "DaVinci Resolve",
  "category": "Creative Suite",
  "items": [
    {
      "label": "project library",
      "source": "resolve",
      "target": "%USERPROFILE%/Documents/DaVinci Resolve",
      "description": "DaVinci Resolve project database and settings"
    }
  ]
}
```

- [ ] **Step 3: Add Utilities apps**

```json
{
  "name": "ShareX",
  "category": "Utilities",
  "items": [
    {
      "label": "settings & screenshots",
      "source": "sharex",
      "target": "%APPDATA%/ShareX",
      "description": "ShareX application settings and screenshot library"
    }
  ]
},
{
  "name": "PowerToys",
  "category": "Utilities",
  "items": [
    {
      "label": "settings",
      "source": "powertoys",
      "target": "%LOCALAPPDATA%/Microsoft/PowerToys",
      "description": "PowerToys module settings and keybindings"
    }
  ]
},
{
  "name": "Claude Desktop",
  "category": "Utilities",
  "items": [
    {
      "label": "MCP config",
      "source": "claude-desktop",
      "target": "%APPDATA%/Claude",
      "description": "Claude Desktop MCP server configuration"
    }
  ]
}
```

- [ ] **Step 4: Add Communication apps**

```json
{
  "name": "Discord",
  "category": "Communication",
  "items": [
    {
      "label": "local settings",
      "source": "discord",
      "target": "%APPDATA%/discord",
      "description": "Discord local user settings and keybindings"
    }
  ]
},
{
  "name": "Spotify",
  "category": "Communication",
  "items": [
    {
      "label": "local files",
      "source": "spotify",
      "target": "%LOCALAPPDATA%/Spotify",
      "description": "Spotify local cache and downloaded files"
    }
  ]
}
```

- [ ] **Step 5: Add Personal Files**

```json
{
  "name": "Personal Files",
  "category": "Personal Files",
  "items": [
    {
      "label": "Desktop",
      "source": "personal/Desktop",
      "target": "%USERPROFILE%/Desktop",
      "description": "All files and folders stored on the Desktop",
      "strategy": "mirror"
    },
    {
      "label": "Documents",
      "source": "personal/Documents",
      "target": "%USERPROFILE%/Documents",
      "description": "Documents folder — select specific items via folder picker",
      "strategy": "selective"
    },
    {
      "label": "Videos",
      "source": "personal/Videos",
      "target": "%USERPROFILE%/Videos",
      "description": "Videos folder — select specific items via folder picker",
      "strategy": "selective"
    },
    {
      "label": "Pictures",
      "source": "personal/Pictures",
      "target": "%USERPROFILE%/Pictures",
      "description": "Pictures folder — select specific items via folder picker",
      "strategy": "selective"
    },
    {
      "label": "Music",
      "source": "personal/Music",
      "target": "%USERPROFILE%/Music",
      "description": "Music folder — select specific items via folder picker",
      "strategy": "selective"
    }
  ]
}
```

- [ ] **Step 5b: Add notes to OBS Studio and Streamer.bot entries**

Per the spec, both apps must note that they should be closed before backup. Find the existing OBS Studio and Streamer.bot entries in `ktuluekit-migration.json` and add or update their `notes` field:

OBS Studio — scenes & profiles:
```json
"notes": "Close OBS before backing up to avoid file locks. Scenes stored as JSON in %APPDATA%/obs-studio/basic/scenes/"
```

Streamer.bot — actions & commands:
```json
"notes": "Close Streamer.bot before backing up to avoid SQLite file locks on .db files"
```

- [ ] **Step 6: Update profiles**

Add new IDs to "Full Restore" profile. The complete list of new IDs to append:

```json
"Git:global config",
"SSH:keys & config",
"VS Code:settings & keybindings",
"Windows Terminal:settings",
"GIMP:brushes & settings",
"Krita:brushes & settings",
"Audacity:settings",
"Blender:preferences & addons",
"Aseprite:settings",
"Kdenlive:settings",
"DaVinci Resolve:project library",
"ShareX:settings & screenshots",
"PowerToys:settings",
"Claude Desktop:MCP config",
"Discord:local settings",
"Spotify:local files",
"Personal Files:Desktop",
"Personal Files:Documents",
"Personal Files:Videos",
"Personal Files:Pictures",
"Personal Files:Music"
```

Add a new "Dev Tools" profile:
```json
{
  "name": "Dev Tools",
  "ids": [
    "Git:global config",
    "SSH:keys & config",
    "VS Code:settings & keybindings",
    "Windows Terminal:settings"
  ]
}
```

- [ ] **Step 7: Verify config loads in running app**

With `wails dev` running, reload the app. Verify all new categories appear in SelectionScreen.

- [ ] **Step 8: Commit config expansion**

> **Note:** Task 3 depends on Task 2 having been committed first. Do not start Task 3 until the Task 2 commit (schema + categoryOrder + config.go) is complete.

```bash
git add ktuluekit-migration.json
git commit -m "feat: expand config with Dev Tools, Creative Suite, Utilities, Communication, Personal Files"
```

---

## Chunk 3: Phase 3 — Folder Picker Go Backend

### Task 4: Add CopyPath to copier

**Files:**
- Modify: `internal/copier/copier.go`
- Create: `internal/copier/copier_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/copier/copier_test.go`:
```go
package copier_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Ktulue/KtulueKit-Migration/internal/copier"
)

func TestCopyPath_File(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	dst := filepath.Join(tmp, "dst", "src.txt")
	if err := os.WriteFile(src, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	n, err := copier.CopyPath(src, dst)
	if err != nil {
		t.Fatalf("CopyPath file: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes, got %d", n)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "hello" {
		t.Errorf("expected 'hello', got %q", string(got))
	}
}

func TestCopyPath_Dir(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "srcdir")
	_ = os.MkdirAll(srcDir, 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("aaa"), 0644) // 3 bytes
	_ = os.WriteFile(filepath.Join(srcDir, "b.txt"), []byte("bb"), 0644)  // 2 bytes
	dstDir := filepath.Join(tmp, "dstdir")

	n, err := copier.CopyPath(srcDir, dstDir)
	if err != nil {
		t.Fatalf("CopyPath dir: %v", err)
	}
	if n != 5 { // 3 + 2 bytes
		t.Errorf("expected 5 bytes, got %d", n)
	}
	if _, err := os.Stat(filepath.Join(dstDir, "a.txt")); err != nil {
		t.Error("a.txt not found in dst")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "b.txt")); err != nil {
		t.Error("b.txt not found in dst")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

```bash
cd "F:/GDriveClone/Claude_Code/KtulueKit-Migration"
go test ./internal/copier/... -v -run TestCopyPath
```
Expected: compile error — `copier.CopyPath undefined`

- [ ] **Step 3: Add CopyPath to copier.go**

Add to `internal/copier/copier.go` after the `CopyFile` function:
```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/copier/... -v -run TestCopyPath
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/copier/
git commit -m "feat: add CopyPath to copier — handles file or directory"
```

---

### Task 5: Add NullReporter and WriteManifest to reporter

**Files:**
- Modify: `internal/reporter/reporter.go`
- Create: `internal/reporter/reporter_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/reporter/reporter_test.go`:
```go
package reporter_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ktulue/KtulueKit-Migration/internal/reporter"
)

func TestNewNull_NoFileCreated(t *testing.T) {
	tmp := t.TempDir()
	r := reporter.NewNull()
	// All methods must not panic and must not create files
	r.Add(reporter.Result{App: "Test", Label: "item", Status: reporter.StatusCopied})
	r.Summary()
	r.Close()

	entries, _ := os.ReadDir(tmp)
	if len(entries) != 0 {
		t.Error("NewNull should not create any files")
	}
	if r.LogPath() != "" {
		t.Errorf("expected empty log path, got %q", r.LogPath())
	}
}

func TestWriteManifest(t *testing.T) {
	tmp := t.TempDir()
	r := reporter.New(tmp)
	r.Add(reporter.Result{
		App: "OBS Studio", Label: "scenes",
		SourcePath: "D:/backup/obs", TargetPath: "C:/AppData/obs",
		Status: reporter.StatusCopied, BytesCopied: 1024,
	})
	r.Add(reporter.Result{
		App: "Personal Files", Label: "Documents",
		SourcePath: "D:/backup/docs", TargetPath: "C:/Users/docs",
		Status: reporter.StatusCopied, BytesCopied: 512,
		SelectedPaths: []string{"D:/backup/docs/Notes.txt"},
	})
	r.Close()

	manifestPath := filepath.Join(tmp, "manifest.json")
	if err := r.WriteManifest(manifestPath); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	var m struct {
		Version string `json:"version"`
		Items   []struct {
			App           string   `json:"app"`
			SelectedPaths []string `json:"selectedPaths"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m.Version != "1.0" {
		t.Errorf("expected version 1.0, got %q", m.Version)
	}
	if len(m.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(m.Items))
	}
	if len(m.Items[1].SelectedPaths) != 1 {
		t.Error("expected selectedPaths on Documents item")
	}
}
```

- [ ] **Step 2: Run to verify they fail**

```bash
go test ./internal/reporter/... -v
```
Expected: compile errors — `NewNull`, `WriteManifest`, `SelectedPaths` undefined

- [ ] **Step 3: Add timestamp field to Reporter, SelectedPaths to Result**

In `internal/reporter/reporter.go`:

Add `timestamp time.Time` field to the `Reporter` struct. Update `New()` to capture timestamp before building the log filename and assign it to `r.timestamp`. Update `NewNull()` to set `timestamp: time.Now()`. This ensures `WriteManifest`'s `runAt` field matches the log file's timestamp exactly.

Add `SelectedPaths` to the `Result` struct:
```go
type Result struct {
	App           string
	Label         string
	SourcePath    string
	TargetPath    string
	Status        string
	BytesCopied   int64
	Detail        string
	SelectedPaths []string
}
```

- [ ] **Step 4: Add NewNull and WriteManifest**

Add to `internal/reporter/reporter.go`:
```go
// NewNull returns a Reporter that discards all output. Used for dry-run mode.
func NewNull() *Reporter {
	return &Reporter{timestamp: time.Now()}
}

// manifestItem is the JSON structure for a single manifest entry.
type manifestItem struct {
	App           string   `json:"app"`
	Label         string   `json:"label"`
	SourcePath    string   `json:"sourcePath"`
	TargetPath    string   `json:"targetPath"`
	Status        string   `json:"status"`
	BytesCopied   int64    `json:"bytesCopied"`
	SelectedPaths []string `json:"selectedPaths"`
}

// manifest is the top-level JSON structure written for KtulueKit-Cleanup.
type manifest struct {
	Version string         `json:"version"`
	RunAt   string         `json:"runAt"`
	Items   []manifestItem `json:"items"`
}

// WriteManifest serializes all results to a JSON file at the given path.
// This is the contract consumed by KtulueKit-Cleanup.
func (r *Reporter) WriteManifest(path string) error {
	items := make([]manifestItem, 0, len(r.results))
	for _, res := range r.results {
		sp := res.SelectedPaths
		if sp == nil {
			sp = []string{}
		}
		items = append(items, manifestItem{
			App:           res.App,
			Label:         res.Label,
			SourcePath:    res.SourcePath,
			TargetPath:    res.TargetPath,
			Status:        res.Status,
			BytesCopied:   res.BytesCopied,
			SelectedPaths: sp,
		})
	}

	data, err := json.MarshalIndent(manifest{
		Version: "1.0",
		RunAt:   r.timestamp.UTC().Format(time.RFC3339), // same timestamp as log filename
		Items:   items,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling manifest: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
```

Add `"encoding/json"` to the imports in reporter.go.

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/reporter/... -v
```
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/reporter/
git commit -m "feat: add NewNull reporter and WriteManifest for dry-run and Cleanup contract"
```

---

### Task 6: Add selective strategy and dry-run to runner

**Files:**
- Modify: `internal/runner/runner.go`
- Create: `internal/runner/runner_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/runner/runner_test.go`:
```go
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
```

- [ ] **Step 2: Run to verify they fail**

```bash
go test ./internal/runner/... -v
```
Expected: compile errors — `SetSelectivePaths`, `SetDryRun` undefined

- [ ] **Step 3: Add SetSelectivePaths, SetDryRun, SelectedPaths to runner**

In `internal/runner/runner.go`:

Add fields to `Runner` struct:
```go
type Runner struct {
	cfg            *config.Config
	rep            *reporter.Reporter
	selectedIDs    map[string]bool
	selectivePaths map[string][]string
	dryRun         bool
	onProgress     func(ProgressEvent)
}
```

Add setter methods:
```go
// SetSelectivePaths sets per-item path selections for selective strategy items.
func (r *Runner) SetSelectivePaths(paths map[string][]string) {
	r.selectivePaths = paths
}

// SetDryRun enables dry-run mode — paths are resolved but no files are written.
func (r *Runner) SetDryRun(dryRun bool) {
	r.dryRun = dryRun
}
```

Add `SelectedPaths` to `RunResultItem`:
```go
type RunResultItem struct {
	App           string
	Label         string
	SourcePath    string
	TargetPath    string
	Status        string
	BytesCopied   int64
	SelectedPaths []string
}
```

- [ ] **Step 4: Update the Run loop for selective strategy and dry-run**

In the `Run()` method, replace the copy block (after path validation) with:

```go
// Determine copy strategy
strategy := w.item.Strategy
if strategy == "" {
	strategy = "mirror"
}

var bytesCopied int64
var copyErr error

if r.dryRun {
	// Dry-run: estimate size without copying
	bytesCopied, copyErr = estimateSize(sourcePath)
} else if strategy == "selective" {
	paths := r.selectivePaths[w.id]
	for _, p := range paths {
		n, err := copier.CopyPath(p, filepath.Join(targetPath, filepath.Base(p)))
		bytesCopied += n
		if err != nil {
			copyErr = err
			break
		}
	}
} else {
	// mirror or file strategy
	info, _ := os.Stat(sourcePath)
	if info.IsDir() {
		bytesCopied, copyErr = copier.MirrorDir(sourcePath, targetPath)
	} else {
		bytesCopied, copyErr = copier.CopyFile(sourcePath, targetPath)
	}
}
```

Add `SelectedPaths` to the success `RunResultItem` append for selective items:
```go
result.Items = append(result.Items, RunResultItem{
	App: w.app.Name, Label: w.item.Label,
	SourcePath: sourcePath, TargetPath: targetPath,
	Status: reporter.StatusCopied, BytesCopied: bytesCopied,
	SelectedPaths: r.selectivePaths[w.id],
})
```

Add the `estimateSize` helper:
```go
func estimateSize(path string) (int64, error) {
	var total int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	return total, err
}
```

Add `"io/fs"` and `"path/filepath"` to runner imports if not already present.

- [ ] **Step 5: Update reporter.Add calls to pass SelectedPaths**

In the `reportItem` helper, update `reporter.Result` to include SelectedPaths:
```go
func (r *Runner) reportItem(app, label, source, target, status string, bytes int64, detail string, selectedPaths []string) {
	r.rep.Add(reporter.Result{
		App: app, Label: label,
		SourcePath: source, TargetPath: target,
		Status: status, BytesCopied: bytes,
		Detail: detail, SelectedPaths: selectedPaths,
	})
}
```

Update all `reportItem` call sites to pass `nil` for selectedPaths on non-selective items, and `r.selectivePaths[w.id]` for selective items.

- [ ] **Step 6: Run tests to verify they pass**

```bash
go test ./internal/runner/... -v
go test ./internal/... -v
```
Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add internal/runner/ internal/reporter/ internal/copier/
git commit -m "feat: add selective copy strategy, dry-run mode, and SelectedPaths to runner"
```

---

### Task 7: Add ListFolder to App and update StartMigration signature

**Files:**
- Modify: `app.go`
- Modify: `types.go`

- [ ] **Step 1: Add Strategy to ItemView in types.go**

In `types.go`, update `ItemView`:
```go
type ItemView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
	Strategy    string `json:"strategy"`
}
```

Add `FolderEntry` type:
```go
// FolderEntry represents a single file or directory inside a listed folder.
type FolderEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}
```

Update `ManifestEntry` to include `SelectedPaths`:
```go
type ManifestEntry struct {
	App           string   `json:"app"`
	Label         string   `json:"label"`
	SourcePath    string   `json:"sourcePath"`
	TargetPath    string   `json:"targetPath"`
	Status        string   `json:"status"`
	BytesCopied   int64    `json:"bytesCopied"`
	SelectedPaths []string `json:"selectedPaths"`
}
```

- [ ] **Step 2: Update GetConfig to populate Strategy on ItemView**

In `app.go`, inside the `GetConfig` loop, update the `ItemView` construction:
```go
iv := ItemView{
	ID:          app.Name + ":" + item.Label,
	Name:        app.Name + " — " + item.Label,
	Description: item.Description,
	Notes:       item.Notes,
	Strategy:    item.Strategy,
}
```

- [ ] **Step 3: Add ListFolder method to App**

Add to `app.go`:
```go
// ListFolder returns the immediate contents of the directory at path.
// Used by the frontend FolderPicker component.
func (a *App) ListFolder(path string) ([]FolderEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("reading folder: %w", err)
	}
	result := make([]FolderEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		var size int64
		if !e.IsDir() {
			size = info.Size()
		}
		result = append(result, FolderEntry{
			Name:  e.Name(),
			Path:  filepath.Join(path, e.Name()),
			IsDir: e.IsDir(),
			Size:  size,
		})
	}
	return result, nil
}
```

Add `"os"` and `"path/filepath"` to app.go imports if not present. `"os"` is needed for `os.ReadDir` in `ListFolder`.

- [ ] **Step 4: Update StartMigration signature**

Replace the existing `StartMigration` signature:
```go
func (a *App) StartMigration(selectedIDs []string, selectivePaths map[string][]string, dryRun bool) error {
```

Update the runner setup inside the goroutine:
```go
r := runner.New(cfg, rep)
r.SetSelectedIDs(selectedIDs)
r.SetSelectivePaths(selectivePaths)
r.SetDryRun(dryRun)
r.SetOnProgress(func(evt runner.ProgressEvent) {
	runtime.EventsEmit(a.ctx, "progress", evt)
})
```

- [ ] **Step 5: Update buildManifest to include SelectedPaths**

```go
func buildManifest(result runner.RunResult) []ManifestEntry {
	var entries []ManifestEntry
	for _, r := range result.Items {
		sp := r.SelectedPaths
		if sp == nil {
			sp = []string{}
		}
		entries = append(entries, ManifestEntry{
			App:           r.App,
			Label:         r.Label,
			SourcePath:    r.SourcePath,
			TargetPath:    r.TargetPath,
			Status:        r.Status,
			BytesCopied:   r.BytesCopied,
			SelectedPaths: sp,
		})
	}
	return entries
}
```

- [ ] **Step 6: Add WriteManifest call after run in StartMigration goroutine**

After `result := r.Run()`, add — but **only when not in dry-run mode**:
```go
// Write manifest JSON for KtulueKit-Cleanup (skipped in dry-run)
var manifestPath string
if !dryRun {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	manifestPath = filepath.Join("logs", fmt.Sprintf("manifest_%s.json", timestamp))
	if err := rep.WriteManifest(manifestPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write manifest: %v\n", err)
		manifestPath = ""
	}
}
```

Update the `SummaryResult` emission to include `ManifestPath`:
```go
runtime.EventsEmit(a.ctx, "complete", SummaryResult{
	...
	ManifestPath: manifestPath,
})
```

Add `ManifestPath string \`json:"manifestPath"\`` to `SummaryResult` in `types.go`.

Add necessary imports (`"time"`, `"os"`, `"fmt"`) to app.go.

- [ ] **Step 7: Use NullReporter for dry-run in StartMigration**

```go
var rep *reporter.Reporter
if dryRun {
	rep = reporter.NewNull()
} else {
	rep = reporter.New("logs")
}
```

- [ ] **Step 8: Restart wails dev to regenerate bindings**

Stop and restart `wails dev`. Verify `frontend/wailsjs/go/main/App.js` now includes `ListFolder` and the updated `StartMigration`.

- [ ] **Step 9: Compile check**

```bash
go build ./...
```
Expected: no errors

- [ ] **Step 10: Commit**

```bash
git add app.go types.go
git commit -m "feat: add ListFolder, update StartMigration for selective/dry-run, write manifest JSON"
```

---

## Chunk 4: Phase 3 Svelte Frontend + Phase 4 UI Polish

### Task 8: Create FolderPicker component

**Files:**
- Create: `frontend/src/components/FolderPicker.svelte`

- [ ] **Step 1: Create FolderPicker.svelte**

```svelte
<script>
  import { ListFolder } from '../../wailsjs/go/main/App'

  export let sourcePath = ''
  export let itemId = ''
  export let onConfirm = (id, paths) => {}
  export let onCancel = () => {}

  let entries = []
  let selected = new Set()
  let loading = true
  let error = null

  $: if (sourcePath) {
    loading = true
    error = null
    ListFolder(sourcePath)
      .then(e => { entries = e; loading = false })
      .catch(err => { error = err.toString(); loading = false })
  }

  $: allChecked = entries.length > 0 && entries.every(e => selected.has(e.path))

  function toggleAll() {
    if (allChecked) {
      selected = new Set()
    } else {
      selected = new Set(entries.map(e => e.path))
    }
  }

  function toggle(path) {
    selected = new Set(selected)
    if (selected.has(path)) selected.delete(path)
    else selected.add(path)
  }

  function confirm() {
    onConfirm(itemId, [...selected])
  }

  function formatSize(bytes) {
    if (bytes === 0) return '—'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1048576) return `${(bytes/1024).toFixed(1)} KB`
    return `${(bytes/1048576).toFixed(1)} MB`
  }
</script>

<div class="overlay">
  <div class="modal">
    <div class="modal-header">
      <h3>Select items to migrate</h3>
      <span class="path">{sourcePath}</span>
    </div>

    <div class="modal-toolbar">
      <button class="select-all-btn" on:click={toggleAll} disabled={loading}>
        {allChecked ? 'Deselect all' : 'Select all'}
      </button>
      <span class="count">{selected.size} selected</span>
    </div>

    <div class="modal-body">
      {#if loading}
        <p class="state-msg">Loading...</p>
      {:else if error}
        <p class="state-msg error">{error}</p>
      {:else if entries.length === 0}
        <p class="state-msg">Folder is empty</p>
      {:else}
        {#each entries as entry}
          <label class="entry-row">
            <input
              type="checkbox"
              checked={selected.has(entry.path)}
              on:change={() => toggle(entry.path)}
            />
            <span class="entry-icon">{entry.isDir ? '📁' : '📄'}</span>
            <span class="entry-name">{entry.name}</span>
            <span class="entry-size">{formatSize(entry.size)}</span>
          </label>
        {/each}
      {/if}
    </div>

    <div class="modal-footer">
      <button class="cancel-btn" on:click={onCancel}>Cancel</button>
      <button
        class="confirm-btn"
        disabled={selected.size === 0}
        on:click={confirm}
      >
        Confirm ({selected.size})
      </button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0,0,0,0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }
  .modal {
    background: #1e1e1e;
    border: 1px solid #333;
    border-radius: 6px;
    width: 560px;
    max-height: 70vh;
    display: flex;
    flex-direction: column;
  }
  .modal-header {
    padding: 16px 20px 12px;
    border-bottom: 1px solid #333;
  }
  .modal-header h3 { margin: 0 0 4px; font-size: 15px; }
  .path {
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: 11px;
    color: #666;
    word-break: break-all;
  }
  .modal-toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px 20px;
    background: #181818;
    border-bottom: 1px solid #2a2a2a;
  }
  .select-all-btn {
    background: transparent;
    color: #999;
    border: 1px solid #444;
    border-radius: 4px;
    padding: 3px 10px;
    font-size: 12px;
    cursor: pointer;
  }
  .select-all-btn:hover:not(:disabled) { color: #e0e0e0; border-color: #666; }
  .select-all-btn:disabled { opacity: 0.4; cursor: not-allowed; }
  .count { font-size: 12px; color: #666; }
  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: 4px 0;
  }
  .entry-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 20px;
    cursor: pointer;
    font-size: 13px;
  }
  .entry-row:hover { background: #2a2a2a; }
  input[type="checkbox"] { accent-color: #2ea043; flex-shrink: 0; }
  .entry-icon { font-size: 14px; }
  .entry-name { flex: 1; color: #ddd; }
  .entry-size { font-size: 11px; color: #666; font-family: 'Cascadia Code', 'Consolas', monospace; }
  .state-msg { padding: 20px; text-align: center; color: #666; }
  .state-msg.error { color: #e55; }
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 12px 20px;
    border-top: 1px solid #333;
    background: #181818;
  }
  .cancel-btn {
    background: transparent;
    color: #999;
    border: 1px solid #444;
    border-radius: 4px;
    padding: 8px 20px;
    cursor: pointer;
  }
  .cancel-btn:hover { color: #e0e0e0; border-color: #666; }
  .confirm-btn {
    background: #2ea043;
    color: #fff;
    border: none;
    border-radius: 4px;
    padding: 8px 20px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
  }
  .confirm-btn:disabled { background: #333; color: #666; cursor: not-allowed; }
  .confirm-btn:not(:disabled):hover { background: #3ab553; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/FolderPicker.svelte
git commit -m "feat: add FolderPicker modal component"
```

---

### Task 9: Wire FolderPicker into SelectionScreen and App

**Files:**
- Modify: `frontend/src/components/ItemRow.svelte`
- Modify: `frontend/src/screens/SelectionScreen.svelte`
- Modify: `frontend/src/App.svelte`

- [ ] **Step 1: Update ItemRow to accept strategy and picker trigger**

Replace `ItemRow.svelte` content:
```svelte
<script>
  export let item
  export let checked
  export let onChange
  export let onOpenPicker = () => {}

  let showTooltip = false
  $: tooltipText = item.description || item.notes || ''
  $: isSelective = item.strategy === 'selective'
</script>

<div class="item-row">
  <label>
    <input
      type="checkbox"
      {checked}
      on:change={(e) => {
        if (isSelective && e.target.checked) {
          e.preventDefault()
          onOpenPicker(item)
        } else {
          onChange()
        }
      }}
    />
    <span>{item.name}</span>
    {#if isSelective && checked}
      <button class="picker-btn" on:click|stopPropagation={() => onOpenPicker(item)}>
        Edit selection
      </button>
    {/if}
  </label>
  {#if tooltipText}
    <div
      class="tooltip-trigger"
      on:mouseenter={() => showTooltip = true}
      on:mouseleave={() => showTooltip = false}
    >
      ?
      {#if showTooltip}
        <div class="tooltip">{tooltipText}</div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .item-row {
    display: flex;
    align-items: center;
    padding: 6px 0;
    gap: 8px;
  }
  label {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    flex: 1;
    font-size: 13px;
  }
  input[type="checkbox"] { accent-color: #2ea043; }
  .picker-btn {
    background: transparent;
    color: #2ea043;
    border: 1px solid #2ea043;
    border-radius: 3px;
    padding: 1px 8px;
    font-size: 11px;
    cursor: pointer;
  }
  .picker-btn:hover { background: rgba(46,160,67,0.1); }
  .tooltip-trigger {
    position: relative;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: #444;
    color: #999;
    font-size: 11px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: help;
    flex-shrink: 0;
  }
  .tooltip {
    position: absolute;
    bottom: calc(100% + 8px);
    right: 0;
    background: #333;
    color: #ddd;
    padding: 8px 12px;
    border-radius: 4px;
    font-size: 12px;
    white-space: normal;
    width: 250px;
    z-index: 10;
    box-shadow: 0 2px 8px rgba(0,0,0,0.4);
  }
</style>
```

- [ ] **Step 2: Update SelectionScreen to pass picker callback and dry-run toggle**

In `SelectionScreen.svelte`, add dry-run toggle to header and pass `onOpenPicker` down:
```svelte
<script>
  import CategoryAccordion from '../components/CategoryAccordion.svelte'

  export let configView
  export let onStart = (ids, selectivePaths, dryRun) => {}
  export let onOpenPicker = (item) => {}

  let selected = new Set()
  let profileValue = ''
  let dryRun = false

  // ... existing loadProfile, toggleItem, handleStart unchanged
  function handleStart() {
    onStart([...selected], {}, dryRun)
  }
  $: selectedCount = selected.size
</script>
```

Add dry-run toggle to the header:
```svelte
<header>
  <div class="header-left">
    <h1>KtulueKit <span class="accent">Migration</span></h1>
  </div>
  <div class="header-right">
    <label class="dry-run-label">
      <input type="checkbox" bind:checked={dryRun} />
      Dry run
    </label>
    <select on:change={loadProfile} bind:value={profileValue}>
      ...
    </select>
  </div>
</header>
```

Add dry-run label style:
```css
.dry-run-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #999;
  cursor: pointer;
}
.dry-run-label input { accent-color: #d4a017; }
```

Pass `onOpenPicker` into `CategoryAccordion`:
```svelte
<CategoryAccordion {category} {selected} onToggle={handleToggle} {onOpenPicker} />
```

- [ ] **Step 3: Thread onOpenPicker through CategoryAccordion to ItemRow, and guard "Select all" for selective items**

In `CategoryAccordion.svelte`, add:
```svelte
export let onOpenPicker = () => {}
```
Pass to ItemRow:
```svelte
<ItemRow {item} checked={...} onChange={...} {onOpenPicker} />
```

**Also update `toggleAll` in `CategoryAccordion.svelte`** to skip selective items — they must go through the picker, not be bulk-added to `selected`:
```svelte
function toggleAll(e) {
  e.stopPropagation()
  if (allChecked) {
    category.items.forEach(item => selected.delete(item.id))
  } else {
    category.items.forEach(item => {
      if (item.strategy === 'selective') return  // selective items must use the picker
      selected.add(item.id)
    })
  }
  onToggle()
}
```

This ensures a checked selective item in `selected` always has a corresponding entry in `selectivePaths`.

- [ ] **Step 4: Update App.svelte to manage picker state and selectivePaths**

> **Important:** Steps 4 and 5 must be implemented together before committing. Step 4 sets up the picker state; Step 5 adds `GetSourcePath` which provides the correct `sourcePath`. Do not commit Step 4 code alone — `sourcePath` is resolved in Step 5.

```svelte
<script>
  // ... existing imports
  import FolderPicker from './components/FolderPicker.svelte'
  import { GetSourcePath } from '../wailsjs/go/main/App'

  // ... existing state
  let selectivePaths = {}
  let pickerItem = null      // null = closed
  let pickerSourcePath = ''  // resolved backup source path for the current picker item

  async function handleOpenPicker(item) {
    try {
      pickerSourcePath = await GetSourcePath(item.id)
      pickerItem = item
    } catch (err) {
      console.error('Could not resolve source path for picker:', err)
    }
  }

  function handlePickerConfirm(itemId, paths) {
    selectivePaths = { ...selectivePaths, [itemId]: paths }
    pickerItem = null
    pickerSourcePath = ''
  }

  function handlePickerCancel() {
    pickerItem = null
    pickerSourcePath = ''
  }

  async function handleStartMigration(selectedIDs, userSelectivePaths, dryRun) {
    progressEvents = []
    screen = 'progress'
    try {
      await StartMigration(selectedIDs, { ...selectivePaths, ...userSelectivePaths }, dryRun)
    } catch (err) {
      summaryResult = { failed: [err.toString()], copied: [], skipped: [], manifest: [] }
      screen = 'summary'
    }
  }
</script>

<!-- In template, add FolderPicker overlay and pass onOpenPicker to SelectionScreen -->
{#if pickerItem}
  <FolderPicker
    sourcePath={pickerSourcePath}
    itemId={pickerItem.id}
    onConfirm={handlePickerConfirm}
    onCancel={handlePickerCancel}
  />
{/if}
```

- [ ] **Step 5: Add GetSourcePath to app.go**

```go
// GetSourcePath resolves the backup source path for a given item ID.
// Used by the frontend FolderPicker to list folder contents.
func (a *App) GetSourcePath(itemID string) (string, error) {
	cfg, err := config.Load(a.configPath)
	if err != nil {
		return "", err
	}
	for _, app := range cfg.Apps {
		for _, item := range app.Items {
			if app.Name+":"+item.Label == itemID {
				return mapper.BuildSourcePath(cfg.BackupRoot, item.Source), nil
			}
		}
	}
	return "", fmt.Errorf("item not found: %s", itemID)
}
```

Update `App.svelte` to call `GetSourcePath(item.id)` before opening picker, then pass the result as `sourcePath` to `FolderPicker`.

- [ ] **Step 6: Restart wails dev and verify**

- FolderPicker opens when checking a selective item
- Confirm stores paths, Cancel unchecks item
- Dry-run toggle shows in header
- Non-selective items behave as before

- [ ] **Step 7: Commit**

```bash
git add frontend/src/
git commit -m "feat: wire FolderPicker into SelectionScreen, add dry-run toggle"
```

---

### Task 10: Phase 4 — UI Polish Pass

**Files:**
- Modify: all screen and component `.svelte` files

- [ ] **Step 1: Audit spacing against W11 (px values, 12px/20px padding)**

W11 uses `padding: 12px 20px` for headers/footers and `padding: 12px 20px` for content areas. Update Migration screens to match:
- `header` / `footer`: `padding: 12px 20px`
- Button font sizes: `14px` or `15px`
- Body text: `13px`
- Secondary text: `12px`

- [ ] **Step 2: Update SelectionScreen, ProgressScreen, SummaryScreen**

Go through each screen, replace rem values with px equivalents matching W11 conventions. The dark color palette stays the same (`#111`, `#1a1a1a`, `#2ea043`).

- [ ] **Step 3: Add dry-run banner to ProgressScreen**

In `ProgressScreen.svelte`, accept a `dryRun` prop and show a banner:
```svelte
{#if dryRun}
  <div class="dry-run-banner">Dry run — no files will be copied</div>
{/if}
```
```css
.dry-run-banner {
  background: #2a2000;
  color: #d4a017;
  padding: 6px 20px;
  font-size: 12px;
  text-align: center;
  border-bottom: 1px solid #3a3000;
}
```

- [ ] **Step 4: Update SummaryScreen to show manifest path**

Add manifest path display below log path in the footer, using same style as log path row.

- [ ] **Step 5: Verify visually in wails dev**

Check all three screens, FolderPicker modal, dry-run banner. Everything should feel tight and consistent.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/
git commit -m "style: UI polish pass — align spacing and font sizes with KtulueKit-W11"
```

---

## Chunk 5: Phase 5 & 6 — Testing and Build

### Task 11: Update TODO.md to reflect completed phases

**Files:**
- Modify: `TODO.md`

- [ ] **Step 1: Update TODO.md**

Mark Phase 1, 2, 3, 4 items complete. Leave Phase 5 (testing) and Phase 6 (build) open.

- [ ] **Step 2: Commit**

```bash
git add TODO.md
git commit -m "docs: update TODO — phases 1-4 complete"
```

---

### Task 12: Phase 5 — End-to-end test checklist

This phase is manual. Work through each item in sequence with the backup drive mounted.

- [ ] Verify `%USERPROFILE%` resolves to E: drive profile path
- [ ] Run Full Restore profile — verify all items resolve paths correctly
- [ ] Test a selective item (Documents) — open picker, select one subfolder, confirm only that copies
- [ ] Test dry-run — enable toggle, run Full Restore, confirm no files written, no log file created, and no manifest JSON written (check `logs/` is empty after run)
- [ ] Test skipped behavior — remove one backup source, verify summary shows skipped with detail
- [ ] Verify log file written to `logs/migration_<timestamp>.log`
- [ ] Verify manifest JSON written to `logs/manifest_<timestamp>.json`
- [ ] Open `logs/manifest_<timestamp>.json` and confirm JSON is valid and `selectedPaths` populated for selective items
- [ ] Verify manifest path shown in SummaryScreen
- [ ] **Manual:** Launch LurkBait after restore, confirm custom catch images relink
- [ ] Fix any issues found, commit fixes

---

### Task 13: Phase 6 — Build and ship

**Files:** none new

- [ ] **Step 1: Build the exe**

```bash
cd "F:/GDriveClone/Claude_Code/KtulueKit-Migration"
wails build
```

Output: `build/bin/ktuluekit-migration.exe`

- [ ] **Step 2: Copy config alongside exe**

Confirm `ktuluekit-migration.json` is in the same directory as the exe for distribution.

- [ ] **Step 3: Smoke test the built exe**

Run `build/bin/ktuluekit-migration.exe` directly (not via `wails dev`). Verify it opens, loads config, and completes a run. No Go/Wails/Node should be required on the machine.

- [ ] **Step 4: Update TODO.md — mark all done**

- [ ] **Step 5: Final commit**

```bash
git add TODO.md docs/ build/
git commit -m "feat: KtulueKit-Migration v1.0 — complete build"
```

> **Note:** Do not stage `build/bin/ktuluekit-migration.exe` unless the `.gitignore` already excludes `build/`. Verify with `git status` first.

---

## Run All Tests

At any point, run the full test suite:

```bash
cd "F:/GDriveClone/Claude_Code/KtulueKit-Migration"
go test ./internal/... -v
```

Expected output: all tests PASS across copier, reporter, and runner packages.
