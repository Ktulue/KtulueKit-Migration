# Source Discovery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Auto-scan a cloned Windows drive to discover where each config item's app data lives, display results inline, and pass discovered paths through to the migration engine.

**Architecture:** New `internal/discovery` package scans `<drive>\Users\*` profiles, resolves each config item's `target` env vars against the cloned drive's user folder structure, and returns per-item results. A new `sourcePathMap` parameter threads discovered paths through `PreflightCheck` and `StartMigration` into the runner. Frontend adds a Scan button to PathBar and shows discovery status on each ItemRow.

**Tech Stack:** Go 1.21+, Svelte 3 (Wails v2 frontend), Wails v2 bindings

**Spec:** `docs/superpowers/specs/2026-03-22-source-discovery-design.md`

---

## File Structure

### New files
- `internal/discovery/discovery.go` — Scan function and types
- `internal/discovery/discovery_test.go` — Tests for discovery logic

### Modified files
- `app.go:106` — Add `sourcePathMap` param to `StartMigration`
- `app.go:267` — Add `sourcePathMap` param to `PreflightCheck`
- `app.go` — Add `ScanDrive` binding method
- `internal/runner/runner.go:48-56` — Add `sourcePathMap` field to Runner
- `internal/runner/runner.go:133` — Check sourcePathMap before BuildSourcePath
- `frontend/src/components/PathBar.svelte` — Add Scan button, emit scan event
- `frontend/src/components/ItemRow.svelte` — Add discovery status indicator, fix checkbox visibility
- `frontend/src/screens/SelectionScreen.svelte` — Wire discovery state, pass sourcePathMap to preflight/start
- `frontend/src/App.svelte:7,45` — Import ScanDrive, pass sourcePathMap through

---

### Task 1: Discovery package — types and Scan function

**Files:**
- Create: `internal/discovery/discovery.go`
- Create: `internal/discovery/discovery_test.go`

- [ ] **Step 1: Write the test file with core test cases**

```go
// internal/discovery/discovery_test.go
package discovery

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ktulue/KtulueKit-Migration/internal/config"
)

// helper to create a fake cloned drive structure
func setupFakeDrive(t *testing.T) (string, *config.Config) {
	t.Helper()
	root := t.TempDir()

	// Create a user profile with AppData
	user := filepath.Join(root, "Users", "TestUser")
	dirs := []string{
		filepath.Join(user, "AppData", "Roaming", "obs-studio", "basic"),
		filepath.Join(user, "AppData", "Local", "BraveSoftware", "Brave-Browser", "User Data", "Default"),
		filepath.Join(user, ".ssh"),
		filepath.Join(user, "Documents"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{
		Version:    "1.0",
		BackupRoot: "D:\\Backup",
		Apps: []config.App{
			{
				Name:     "OBS Studio",
				Category: "Streaming",
				Items: []config.Item{
					{Label: "scenes & profiles", Source: "obs-studio/basic", Target: "%APPDATA%/obs-studio/basic"},
				},
			},
			{
				Name:     "Brave Browser",
				Category: "Browser & Identity",
				Items: []config.Item{
					{Label: "user profile", Source: "brave/User Data/Default", Target: "%LOCALAPPDATA%/BraveSoftware/Brave-Browser/User Data/Default"},
				},
			},
			{
				Name:     "SSH",
				Category: "Dev Tools",
				Items: []config.Item{
					{Label: "keys & config", Source: "ssh", Target: "%USERPROFILE%/.ssh"},
				},
			},
			{
				Name:     "Missing App",
				Category: "Other",
				Items: []config.Item{
					{Label: "not here", Source: "missing", Target: "%APPDATA%/DoesNotExist"},
				},
			},
		},
	}

	return root, cfg
}

func TestScan_FindsItemsInProfile(t *testing.T) {
	root, cfg := setupFakeDrive(t)
	result, err := Scan(context.Background(), root, cfg)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if result.TotalCount != 4 {
		t.Errorf("TotalCount = %d, want 4", result.TotalCount)
	}
	if result.FoundCount != 3 {
		t.Errorf("FoundCount = %d, want 3", result.FoundCount)
	}

	// Check specific items
	itemMap := make(map[string]DiscoveredItem)
	for _, item := range result.Items {
		itemMap[item.ID] = item
	}

	obs := itemMap["OBS Studio:scenes & profiles"]
	if !obs.Found {
		t.Error("OBS Studio should be found")
	}
	if obs.SourcePath == "" {
		t.Error("OBS Studio SourcePath should not be empty")
	}

	missing := itemMap["Missing App:not here"]
	if missing.Found {
		t.Error("Missing App should not be found")
	}
}

func TestScan_NoUsersDir(t *testing.T) {
	root := t.TempDir() // empty dir, no Users folder
	cfg := &config.Config{
		Version:    "1.0",
		BackupRoot: "D:\\Backup",
		Apps: []config.App{
			{
				Name:     "Test",
				Category: "Test",
				Items: []config.Item{
					{Label: "item", Source: "test", Target: "%APPDATA%/test"},
				},
			},
		},
	}

	result, err := Scan(context.Background(), root, cfg)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if result.FoundCount != 0 {
		t.Errorf("FoundCount = %d, want 0", result.FoundCount)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", result.TotalCount)
	}
}

func TestScan_PicksBestProfile(t *testing.T) {
	root := t.TempDir()

	// Profile A: has 1 item
	userA := filepath.Join(root, "Users", "ProfileA")
	os.MkdirAll(filepath.Join(userA, "AppData", "Roaming", "obs-studio", "basic"), 0755)

	// Profile B: has 2 items
	userB := filepath.Join(root, "Users", "ProfileB")
	os.MkdirAll(filepath.Join(userB, "AppData", "Roaming", "obs-studio", "basic"), 0755)
	os.MkdirAll(filepath.Join(userB, ".ssh"), 0755)

	cfg := &config.Config{
		Version:    "1.0",
		BackupRoot: "D:\\Backup",
		Apps: []config.App{
			{Name: "OBS Studio", Category: "Streaming", Items: []config.Item{
				{Label: "scenes", Source: "obs", Target: "%APPDATA%/obs-studio/basic"},
			}},
			{Name: "SSH", Category: "Dev", Items: []config.Item{
				{Label: "keys", Source: "ssh", Target: "%USERPROFILE%/.ssh"},
			}},
		},
	}

	result, err := Scan(context.Background(), root, cfg)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}

	// Profile B should win (2 matches vs 1)
	if result.FoundCount != 2 {
		t.Errorf("FoundCount = %d, want 2 (ProfileB should win)", result.FoundCount)
	}
}

func TestScan_FiltersSystemProfiles(t *testing.T) {
	root := t.TempDir()

	// Create system profiles that should be skipped
	for _, name := range []string{"Default", "Public", "All Users", "Default User"} {
		os.MkdirAll(filepath.Join(root, "Users", name, "AppData", "Roaming", "obs-studio", "basic"), 0755)
	}

	cfg := &config.Config{
		Version:    "1.0",
		BackupRoot: "D:\\Backup",
		Apps: []config.App{
			{Name: "OBS", Category: "S", Items: []config.Item{
				{Label: "x", Source: "obs", Target: "%APPDATA%/obs-studio/basic"},
			}},
		},
	}

	result, err := Scan(context.Background(), root, cfg)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if result.FoundCount != 0 {
		t.Errorf("FoundCount = %d, want 0 (system profiles should be skipped)", result.FoundCount)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd F:/GDriveClone/Claude_Code/KtulueKit-Migration && go test ./internal/discovery/ -v`
Expected: FAIL — package does not exist yet

- [ ] **Step 3: Write the discovery package**

```go
// internal/discovery/discovery.go
package discovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Ktulue/KtulueKit-Migration/internal/config"
)

// DiscoveredItem records whether a single config item was found on the scanned drive.
type DiscoveredItem struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	SourcePath string   `json:"sourcePath"`
	Found      bool     `json:"found"`
	Partial    []string `json:"partial"`
}

// Result holds the outcome of a drive scan.
type Result struct {
	Items      []DiscoveredItem `json:"items"`
	FoundCount int              `json:"foundCount"`
	TotalCount int              `json:"totalCount"`
}

// systemProfiles are Windows profile directories that should be skipped.
var systemProfiles = map[string]bool{
	"default":      true,
	"public":       true,
	"all users":    true,
	"default user": true,
}

// envVarPattern matches Windows-style %VAR% environment variables.
var envVarPattern = regexp.MustCompile(`%([^%]+)%`)

// Scan discovers config items on a cloned drive by resolving each item's
// target path against user profiles found on the drive.
func Scan(ctx context.Context, drivePath string, cfg *config.Config) (*Result, error) {
	// Collect all items with IDs
	type itemInfo struct {
		id     string
		label  string
		target string
	}
	var allItems []itemInfo
	for _, app := range cfg.Apps {
		for _, item := range app.Items {
			allItems = append(allItems, itemInfo{
				id:     app.Name + ":" + item.Label,
				label:  app.Name + " — " + item.Label,
				target: item.Target,
			})
		}
	}

	totalCount := len(allItems)

	// Try to find user profiles
	usersDir := filepath.Join(drivePath, "Users")
	profiles, err := listRealProfiles(usersDir)
	if err != nil || len(profiles) == 0 {
		// No Users dir or no real profiles — return all not-found
		items := make([]DiscoveredItem, totalCount)
		for i, info := range allItems {
			items[i] = DiscoveredItem{
				ID:      info.id,
				Label:   info.label,
				Found:   false,
				Partial: []string{},
			}
		}
		return &Result{Items: items, FoundCount: 0, TotalCount: totalCount}, nil
	}

	// Score each profile by how many items it can resolve
	type profileScore struct {
		name  string
		path  string
		count int
		items []DiscoveredItem
	}

	var best profileScore
	for _, prof := range profiles {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		profPath := filepath.Join(usersDir, prof)
		envMap := buildEnvMap(drivePath, profPath)
		var items []DiscoveredItem
		found := 0

		for _, info := range allItems {
			resolved := resolveTargetWithMap(info.target, envMap)
			exists := pathExists(resolved)
			if exists {
				found++
			}
			items = append(items, DiscoveredItem{
				ID:         info.id,
				Label:      info.label,
				SourcePath: resolved,
				Found:      exists,
				Partial:    []string{},
			})
		}

		if found > best.count {
			best = profileScore{name: prof, path: profPath, count: found, items: items}
		}
	}

	// Clear SourcePath for not-found items (don't leak guessed paths)
	for i := range best.items {
		if !best.items[i].Found {
			best.items[i].SourcePath = ""
		}
	}

	return &Result{
		Items:      best.items,
		FoundCount: best.count,
		TotalCount: totalCount,
	}, nil
}

// listRealProfiles returns subdirectory names under usersDir, filtering out
// system profiles. Returns nil if usersDir doesn't exist.
func listRealProfiles(usersDir string) ([]string, error) {
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading Users directory: %w", err)
	}

	var profiles []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if systemProfiles[strings.ToLower(e.Name())] {
			continue
		}
		profiles = append(profiles, e.Name())
	}
	return profiles, nil
}

// buildEnvMap creates a mapping of Windows env var names to paths on the cloned drive.
func buildEnvMap(drivePath, profilePath string) map[string]string {
	return map[string]string{
		"APPDATA":       filepath.Join(profilePath, "AppData", "Roaming"),
		"LOCALAPPDATA":  filepath.Join(profilePath, "AppData", "Local"),
		"USERPROFILE":   profilePath,
	}
}

// resolveTargetWithMap expands %VAR% placeholders using the provided env map
// and normalizes path separators.
func resolveTargetWithMap(target string, envMap map[string]string) string {
	resolved := envVarPattern.ReplaceAllStringFunc(target, func(match string) string {
		varName := strings.Trim(match, "%")
		if val, ok := envMap[strings.ToUpper(varName)]; ok {
			return val
		}
		return match
	})
	return filepath.FromSlash(resolved)
}

// pathExists checks if a path exists on disk.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd F:/GDriveClone/Claude_Code/KtulueKit-Migration && go test ./internal/discovery/ -v`
Expected: All 4 tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/discovery/discovery.go internal/discovery/discovery_test.go
git commit -m "feat: add discovery package for scanning cloned drives"
```

---

### Task 2: Add ScanDrive Wails binding

**Files:**
- Modify: `app.go` — Add ScanDrive method

- [ ] **Step 1: Add the ScanDrive method to app.go**

Add after the `BrowseForFolder` method (after line 263):

```go
// ScanDrive scans a drive path for app data matching the config items.
// Used by the frontend to auto-discover source paths on a cloned drive.
func (a *App) ScanDrive(drivePath string) (*discovery.Result, error) {
	cfg, err := config.Load(a.configPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return discovery.Scan(a.ctx, drivePath, cfg)
}
```

Add `"github.com/Ktulue/KtulueKit-Migration/internal/discovery"` to the imports.

- [ ] **Step 2: Verify it compiles**

Run: `cd F:/GDriveClone/Claude_Code/KtulueKit-Migration && go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add app.go
git commit -m "feat: add ScanDrive Wails binding"
```

---

### Task 3: Add sourcePathMap to PreflightCheck and StartMigration

**Files:**
- Modify: `app.go:106` — StartMigration signature + sourcePathMap plumbing
- Modify: `app.go:267` — PreflightCheck signature + sourcePathMap plumbing
- Modify: `internal/runner/runner.go:48-56` — Add sourcePathMap to Runner
- Modify: `internal/runner/runner.go:133` — Check sourcePathMap before BuildSourcePath

- [ ] **Step 1: Add sourcePathMap field and setter to Runner**

In `internal/runner/runner.go`, add to the Runner struct (after line 55):

```go
	sourcePathMap    map[string]string
```

Add a setter method after `SetDestRootOverride` (after line 92):

```go
// SetSourcePathMap sets per-item source path overrides from discovery.
func (r *Runner) SetSourcePathMap(m map[string]string) {
	r.sourcePathMap = m
}
```

- [ ] **Step 2: Update Runner.Run to check sourcePathMap**

In `internal/runner/runner.go`, replace line 133:

```go
		sourcePath := mapper.BuildSourcePath(r.cfg.BackupRoot, w.item.Source)
```

With:

```go
		// Check for discovered source path override first
		sourcePath := ""
		if r.sourcePathMap != nil {
			sourcePath = r.sourcePathMap[w.id]
		}
		if sourcePath == "" {
			sourcePath = mapper.BuildSourcePath(r.cfg.BackupRoot, w.item.Source)
		}
```

- [ ] **Step 3: Update PreflightCheck signature**

In `app.go`, change line 267 from:

```go
func (a *App) PreflightCheck(selectedIDs []string, sourceRoot string, destRoot string) (PreflightResult, error) {
```

To:

```go
func (a *App) PreflightCheck(selectedIDs []string, sourceRoot string, destRoot string, sourcePathMap map[string]string) (PreflightResult, error) {
```

And update the per-item source check at line 314 from:

```go
			sourcePath := mapper.BuildSourcePath(sourceRoot, item.Source)
```

To:

```go
			sourcePath := ""
			if sourcePathMap != nil {
				sourcePath = sourcePathMap[id]
			}
			if sourcePath == "" {
				sourcePath = mapper.BuildSourcePath(sourceRoot, item.Source)
			}
```

- [ ] **Step 4: Update StartMigration signature**

In `app.go`, change line 106 from:

```go
func (a *App) StartMigration(selectedIDs []string, selectivePaths map[string][]string, dryRun bool, sourceRootOverride string, destRootOverride string) error {
```

To:

```go
func (a *App) StartMigration(selectedIDs []string, selectivePaths map[string][]string, dryRun bool, sourceRootOverride string, destRootOverride string, sourcePathMap map[string]string) error {
```

Also capture sourcePathMap in a local before the goroutine (after line 121):

```go
	srcPaths := sourcePathMap
```

And add after line 155 (`r.SetDestRootOverride(dstOverride)`) inside the goroutine, using the captured local:

```go
		r.SetSourcePathMap(srcPaths)
```

- [ ] **Step 5: Run existing tests to verify nothing broke**

Run: `cd F:/GDriveClone/Claude_Code/KtulueKit-Migration && go test ./... -v`
Expected: All tests PASS (existing tests pass nil/empty for new param)

- [ ] **Step 6: Commit**

```bash
git add app.go internal/runner/runner.go
git commit -m "feat: thread sourcePathMap through preflight and migration"
```

---

### Task 4: Frontend — Scan button in PathBar

**Files:**
- Modify: `frontend/src/components/PathBar.svelte`

- [ ] **Step 1: Add Scan button and event to PathBar**

Add a `scanning` prop and dispatch a `scan` event. Add a "Scan" button next to the Source Browse button.

In the `<script>` section, add:

```js
  export let scanning = false

  function handleScan() {
    sourceRoot = normaliseDrive(sourceRoot)
    emitChange()
    dispatch('scan', { sourcePath: sourceRoot })
  }
```

In the template, after the Source Browse button (line 75), add the Scan button:

```svelte
    <button class="scan-btn" on:click={handleScan} disabled={!sourceRoot || scanning}>
      {scanning ? 'Scanning...' : 'Scan'}
    </button>
```

In the `<style>` section, add:

```css
  .scan-btn {
    background: transparent;
    color: var(--color-success);
    border: 1px solid var(--color-success);
    border-radius: var(--radius);
    padding: var(--spacing-xs) var(--spacing-lg);
    font-size: var(--font-size-sm);
    cursor: pointer;
    flex-shrink: 0;
    transition: color 100ms ease, border-color 100ms ease, background 100ms ease;
  }

  .scan-btn:hover:not(:disabled) {
    background: rgba(46, 160, 67, 0.1);
  }

  .scan-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
```

- [ ] **Step 2: Verify it renders**

Run: `wails dev` and confirm the Scan button appears next to Browse on the Source row.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/PathBar.svelte
git commit -m "feat: add Scan button to PathBar"
```

---

### Task 5: Frontend — Wire discovery state in SelectionScreen

**Files:**
- Modify: `frontend/src/screens/SelectionScreen.svelte`
- Modify: `frontend/src/App.svelte`

- [ ] **Step 1: Add discovery state and scan handler to SelectionScreen**

In `SelectionScreen.svelte`, add imports and state:

```js
  import { ScanDrive, BrowseForFolder } from '../../wailsjs/go/main/App'

  let discoveryMap = {}  // itemID → DiscoveredItem
  let scanning = false
  let scanDone = false

  async function handleAssist(item) {
    try {
      const chosen = await BrowseForFolder(sourceRoot)
      if (chosen) {
        discoveryMap = {
          ...discoveryMap,
          [item.id]: { ...discoveryMap[item.id], found: true, sourcePath: chosen }
        }
        selected.add(item.id)
        selected = new Set(selected)
      }
    } catch (err) {
      console.error('Assist browse failed:', err)
    }
  }

  async function handleScan(e) {
    const { sourcePath } = e.detail
    if (!sourcePath) return
    scanning = true
    scanDone = false
    discoveryMap = {}
    try {
      const result = await ScanDrive(sourcePath)
      const map = {}
      for (const item of result.items) {
        map[item.id] = item
      }
      discoveryMap = map
      scanDone = true

      // Auto-select found items
      for (const item of result.items) {
        if (item.found) {
          selected.add(item.id)
        }
      }
      selected = new Set(selected)
    } catch (err) {
      console.error('Scan failed:', err)
    } finally {
      scanning = false
    }
  }
```

- [ ] **Step 2: Pass scanning prop and scan event to PathBar**

Update the PathBar usage:

```svelte
  <PathBar
    {sourceRoot}
    {destRoot}
    {scanning}
    on:change={(e) => { sourceRoot = e.detail.sourceRoot; destRoot = e.detail.destRoot }}
    on:scan={handleScan}
  />
```

- [ ] **Step 3: Pass discoveryMap to CategoryAccordion**

Update the CategoryAccordion usage:

```svelte
      <CategoryAccordion {category} {selected} {discoveryMap} onToggle={handleToggle} onOpenPicker={handleOpenPickerWrapped} onAssist={handleAssist} />
```

- [ ] **Step 4: Build sourcePathMap and pass to preflight and start**

Update `handlePreflight`:

```js
  function buildSourcePathMap() {
    const map = {}
    for (const [id, item] of Object.entries(discoveryMap)) {
      if (item.found && item.sourcePath) {
        map[id] = item.sourcePath
      }
    }
    return map
  }

  async function handlePreflight() {
    try {
      preflightResult = await PreflightCheck([...selected], sourceRoot, destRoot, buildSourcePathMap())
      preflightDone = true
    } catch (e) {
      preflightDone = false
      console.error('Preflight failed:', e)
    }
  }
```

Update `handleStart`:

```js
  function handleStart() {
    onStart([...selected], {}, dryRun, sourceRoot, destRoot, buildSourcePathMap())
  }
```

- [ ] **Step 5: Update App.svelte to pass sourcePathMap through to StartMigration**

In `App.svelte`, update `handleStartMigration` signature and call:

```js
  async function handleStartMigration(selectedIDs, userSelectivePaths, isDryRun, sourceRoot, destRoot, sourcePathMap) {
    dryRun = isDryRun
    progressEvents = []
    screen = 'progress'
    try {
      await StartMigration(selectedIDs, { ...selectivePaths, ...userSelectivePaths }, isDryRun, sourceRoot || '', destRoot || '', sourcePathMap || {})
    } catch (err) {
      summaryResult = { failed: [err.toString()], copied: [], skipped: [], manifest: [] }
      screen = 'summary'
    }
  }
```

- [ ] **Step 6: Verify it compiles and scan works**

Run: `wails dev`, set source to a drive, click Scan, verify results appear in console.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/screens/SelectionScreen.svelte frontend/src/App.svelte
git commit -m "feat: wire discovery state through selection screen and App"
```

---

### Task 6: Frontend — Discovery status on ItemRow + checkbox fix

**Files:**
- Modify: `frontend/src/components/ItemRow.svelte`
- Modify: `frontend/src/components/CategoryAccordion.svelte`

- [ ] **Step 1: Update CategoryAccordion to pass discovery info and fix toggleAll**

In `CategoryAccordion.svelte`, add the props:

```js
  export let discoveryMap = {}
  export let onAssist = () => {}
```

Update `toggleAll` to skip not-found items (replace the existing function):

```js
  function toggleAll(e) {
    e.stopPropagation()
    if (allChecked) {
      category.items.forEach(item => selected.delete(item.id))
    } else {
      category.items.forEach(item => {
        if (item.strategy === 'selective') return  // must use picker
        const disc = discoveryMap[item.id]
        if (disc && !disc.found) return  // skip not-found items
        selected.add(item.id)
      })
    }
    onToggle()
  }
```

And pass discoveryMap to ItemRow:

```svelte
        <ItemRow
          {item}
          checked={selected.has(item.id)}
          discoveryStatus={discoveryMap[item.id] || null}
          {onAssist}
          onChange={() => {
```

- [ ] **Step 2: Update ItemRow with discovery status and checkbox fix**

In `ItemRow.svelte`, add the props:

```js
  export let discoveryStatus = null
  export let onAssist = () => {}
```

Add computed values:

```js
  $: discovered = discoveryStatus !== null
  $: discoveredFound = discoveryStatus?.found ?? false
  $: discoveredNotFound = discovered && !discoveredFound
```

Replace the `<input type="checkbox">` with a custom styled checkbox and add discovery indicator. Replace the entire template section:

```svelte
<div class="item-row" class:dimmed={discoveredNotFound}>
  <label>
    <span
      class="checkbox"
      class:checked={checked}
      class:disabled={discoveredNotFound}
      on:click|preventDefault={() => {
        if (discoveredNotFound) return
        if (isSelective && !checked) {
          onOpenPicker(item)
        } else {
          onChange()
        }
      }}
      role="checkbox"
      aria-checked={checked}
      tabindex="0"
      on:keydown={(e) => { if (e.key === ' ' || e.key === 'Enter') { e.preventDefault(); onChange() }}}
    >
      {#if checked}
        <svg width="12" height="12" viewBox="0 0 12 12">
          <polyline points="2,6 5,9 10,3" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      {/if}
    </span>
    <span class="item-name">{item.name}</span>
    {#if discoveredFound}
      <span class="discovery-badge found" title={discoveryStatus.sourcePath}>found</span>
    {:else if discoveredNotFound}
      <span class="discovery-badge not-found">not found</span>
      <button class="assist-btn" on:click|stopPropagation={() => onAssist(item)}>Locate</button>
    {/if}
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
```

Replace the entire `<style>` section:

```css
<style>
  .item-row {
    display: flex;
    align-items: center;
    padding: var(--spacing-sm) 0;
    gap: var(--spacing-md);
  }
  .item-row.dimmed {
    opacity: 0.45;
  }
  label {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
    cursor: pointer;
    flex: 1;
    font-size: var(--font-size-sm);
  }
  .checkbox {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    border: 2px solid var(--color-text-secondary);
    border-radius: 3px;
    flex-shrink: 0;
    cursor: pointer;
    transition: background 100ms ease, border-color 100ms ease;
    color: #fff;
  }
  .checkbox.checked {
    background: var(--color-accent);
    border-color: var(--color-accent);
  }
  .checkbox.disabled {
    border-color: var(--color-border-input);
    cursor: not-allowed;
  }
  .discovery-badge {
    font-size: var(--font-size-xs);
    padding: 1px 6px;
    border-radius: var(--radius);
    font-weight: 600;
  }
  .discovery-badge.found {
    color: var(--color-success);
    background: rgba(46, 160, 67, 0.12);
  }
  .discovery-badge.not-found {
    color: var(--color-text-secondary);
    background: rgba(136, 136, 136, 0.12);
  }
  .assist-btn {
    background: transparent;
    color: var(--color-warning);
    border: 1px solid var(--color-warning);
    border-radius: var(--radius);
    padding: 1px var(--spacing-md);
    font-size: var(--font-size-xs);
    cursor: pointer;
    transition: color 100ms ease, border-color 100ms ease, background 100ms ease;
  }
  .assist-btn:hover { background: rgba(230, 168, 23, 0.1); }
  .picker-btn {
    background: transparent;
    color: var(--color-accent);
    border: 1px solid var(--color-accent);
    border-radius: var(--radius);
    padding: 1px var(--spacing-md);
    font-size: var(--font-size-xs);
    cursor: pointer;
    transition: color 100ms ease, border-color 100ms ease, background 100ms ease;
  }
  .picker-btn:hover { background: rgba(14, 127, 212, 0.1); }
  .tooltip-trigger {
    position: relative;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--color-bg-hover);
    color: var(--color-text-secondary);
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
    background: var(--color-bg-hover);
    color: var(--color-text-primary);
    padding: var(--spacing-md) var(--spacing-lg);
    border-radius: var(--radius);
    font-size: var(--font-size-sm);
    white-space: normal;
    width: 250px;
    z-index: 10;
    box-shadow: 0 2px 8px rgba(0,0,0,0.4);
  }
</style>
```

- [ ] **Step 3: Verify UI renders correctly**

Run: `wails dev` and confirm:
- Checkboxes are clearly visible (bordered when unchecked, filled blue when checked)
- After scanning, found items show green "found" badge
- Not-found items are dimmed with "not found" badge

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ItemRow.svelte frontend/src/components/CategoryAccordion.svelte
git commit -m "feat: discovery status on items + fix checkbox visibility"
```

---

### Task 7: Generate Wails bindings and verify full flow

**Files:**
- Modified by `wails generate`: `frontend/wailsjs/go/main/App.js` and `.d.ts`

- [ ] **Step 1: Generate updated Wails JS bindings**

Run: `cd F:/GDriveClone/Claude_Code/KtulueKit-Migration && wails generate module`

This regenerates the JS/TS bindings to include the new `ScanDrive` function and updated `PreflightCheck`/`StartMigration` signatures.

- [ ] **Step 2: Verify full build**

Run: `cd F:/GDriveClone/Claude_Code/KtulueKit-Migration && wails build`
Expected: Build succeeds, exe at `build/bin/`

- [ ] **Step 3: Manual test**

1. Launch the built exe
2. Set Source to the cloned drive (e.g., `E:\`)
3. Click Scan — verify items populate with found/not-found status
4. Found items should be pre-checked
5. Click Pre-flight Check — verify it uses discovered paths
6. Run a dry run to verify full pipeline

- [ ] **Step 4: Run all Go tests**

Run: `cd F:/GDriveClone/Claude_Code/KtulueKit-Migration && go test ./... -v`
Expected: All tests pass

- [ ] **Step 5: Commit**

```bash
git add frontend/wailsjs/
git commit -m "chore: regenerate Wails bindings for discovery"
```

---

### Task 8: Update docs

**Files:**
- Modify: `docs/how-to-use.md`

- [ ] **Step 1: Update how-to-use.md with discovery instructions**

Add a section about the Scan feature after the existing source path documentation. Describe:
- Setting the Source to a drive root
- Clicking Scan to auto-discover items
- How found/not-found items are displayed
- That found items are pre-checked

- [ ] **Step 2: Commit**

```bash
git add docs/how-to-use.md
git commit -m "docs: add source discovery instructions to how-to-use"
```
