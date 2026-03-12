# Path Override & Pre-flight Check Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let users override the source backup root and destination root at runtime, with a pre-flight check that validates all selected item paths before enabling Start Migration.

**Architecture:** Backend gains `ApplyDestOverride` (mapper), `BrowseForFolder` (native OS dialog), `PreflightCheck`, and an extended `StartMigration`. Frontend gains two new components (`PathBar`, `PreflightPanel`) and updates to `SelectionScreen` and `App.svelte`.

**Tech Stack:** Go 1.21, Wails v2, Svelte 4, wailsjs auto-generated bindings (regenerated on build).

**Spec:** `docs/superpowers/specs/2026-03-12-path-override-preflight-design.md`

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `internal/mapper/override.go` | `ApplyDestOverride` pure function |
| Create | `internal/mapper/override_test.go` | Table-driven tests for all guard branches |
| Modify | `types.go` | Add `PreflightResult`, `PreflightItem`; extend `SummaryResult` |
| Modify | `app.go` | Add `BrowseForFolder`, `PreflightCheck`; modify `StartMigration` |
| Create | `frontend/src/components/PathBar.svelte` | Two-row source/dest path bar with status icons |
| Create | `frontend/src/components/PreflightPanel.svelte` | Collapsible pre-flight results + "Run anyway" |
| Modify | `frontend/src/screens/SelectionScreen.svelte` | Replace backup banner; add preflight state machine |
| Modify | `frontend/src/App.svelte` | Wire new props/methods; update Run Again flow |

> **Note on Browse:** The spec says "opens the existing FolderPicker component." FolderPicker is built for multi-file selection within a known source path — wrong tool for picking a root directory. We use `runtime.OpenDirectoryDialog` (native OS picker) via a new `BrowseForFolder` backend method instead. This is strictly better UX.

---

## Chunk 1: Backend

### Task 1: `ApplyDestOverride` — mapper function

**Files:**
- Create: `internal/mapper/override.go`
- Create: `internal/mapper/override_test.go`

- [ ] **Step 1.1: Write the failing tests**

Create `internal/mapper/override_test.go`:

```go
package mapper_test

import (
	"testing"

	"github.com/Ktulue/KtulueKit-Migration/internal/mapper"
)

func TestApplyDestOverride(t *testing.T) {
	cases := []struct {
		name   string
		target string
		dest   string
		want   string
	}{
		// Guard 1: empty destRoot → unchanged
		{
			name:   "empty dest returns target unchanged",
			target: `C:\Users\Foo\AppData\Roaming\App`,
			dest:   "",
			want:   `C:\Users\Foo\AppData\Roaming\App`,
		},
		// Guard 2: target has no drive prefix → unchanged
		{
			name:   "relative target returned unchanged",
			target: `relative\path`,
			dest:   `D:\`,
			want:   `relative\path`,
		},
		{
			name:   "target too short to have drive prefix → unchanged",
			target: `C:`,
			dest:   `D:\`,
			want:   `C:`,
		},
		// Guard 3: destRoot is drive root (len==3) → drive swap
		{
			name:   "drive swap on full path",
			target: `C:\Users\Foo\AppData\Roaming\App`,
			dest:   `D:\`,
			want:   `D:\Users\Foo\AppData\Roaming\App`,
		},
		{
			name:   "drive swap on root only",
			target: `C:\`,
			dest:   `D:\`,
			want:   `D:\`,
		},
		// Guard 4: destRoot is longer path → prefix substitution
		{
			name:   "prefix substitution",
			target: `C:\Users\Foo\AppData\Roaming\App`,
			dest:   `D:\Restored\`,
			want:   `D:\Restored\Users\Foo\AppData\Roaming\App`,
		},
		{
			name:   "prefix substitution strips only X:\\",
			target: `C:\SomeApp`,
			dest:   `E:\Backup\New\`,
			want:   `E:\Backup\New\SomeApp`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapper.ApplyDestOverride(tc.target, tc.dest)
			if got != tc.want {
				t.Errorf("ApplyDestOverride(%q, %q) = %q; want %q", tc.target, tc.dest, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 1.2: Run to confirm tests fail**

```bash
cd "F:/GDriveClone/Claude_Code/KtulueKit-Migration"
go test ./internal/mapper/... -run TestApplyDestOverride -v
```

Expected: compile error — `ApplyDestOverride` undefined.

- [ ] **Step 1.3: Implement `ApplyDestOverride`**

Create `internal/mapper/override.go`:

```go
package mapper

// ApplyDestOverride rewrites resolvedTarget according to destRoot.
//
// Guard order:
//  1. Empty destRoot → return resolvedTarget unchanged.
//  2. resolvedTarget does not match [A-Za-z]:\ → return unchanged.
//  3. len(destRoot)==3 (e.g. "D:\") → drive-letter swap.
//  4. Otherwise → strip X:\ prefix and prepend destRoot.
func ApplyDestOverride(resolvedTarget, destRoot string) string {
	if destRoot == "" {
		return resolvedTarget
	}
	// Guard 2: require X:\ prefix on target
	if len(resolvedTarget) < 3 || resolvedTarget[1] != ':' || resolvedTarget[2] != '\\' {
		return resolvedTarget
	}
	if len(destRoot) == 3 {
		// Drive swap: replace first character (drive letter)
		return string(destRoot[0]) + resolvedTarget[1:]
	}
	// Prefix substitution: strip "X:\" (first 3 chars) and prepend destRoot
	return destRoot + resolvedTarget[3:]
}
```

- [ ] **Step 1.4: Run tests — expect all pass**

```bash
go test ./internal/mapper/... -run TestApplyDestOverride -v
```

Expected: `PASS` for all 7 cases.

- [ ] **Step 1.5: Commit**

```bash
git add internal/mapper/override.go internal/mapper/override_test.go
git commit -m "feat: add ApplyDestOverride to mapper package"
```

---

### Task 2: Update `types.go`

**Files:**
- Modify: `types.go`

- [ ] **Step 2.1: Add `PreflightResult`, `PreflightItem`, extend `SummaryResult`**

Open `types.go`. After the `SummaryResult` block, add:

```go
// PreflightResult is returned by the PreflightCheck backend method.
type PreflightResult struct {
	SourceRootOK    bool            `json:"sourceRootOK"`
	DestRootOK      bool            `json:"destRootOK"`
	HasItemWarnings bool            `json:"hasItemWarnings"`
	Items           []PreflightItem `json:"items"`
	ReadyCount      int             `json:"readyCount"`
	TotalCount      int             `json:"totalCount"`
}

// PreflightItem records the check result for a single selected item.
type PreflightItem struct {
	ID    string `json:"id"`    // app.Name + ":" + item.Label
	Label string `json:"label"` // app.Name + " \u2014 " + item.Label
	Path  string `json:"path"`  // resolved source path actually checked
	Found bool   `json:"found"`
}
```

In `SummaryResult`, add two fields after `ManifestPath`:

```go
SourceRootOverride string          `json:"sourceRootOverride,omitempty"`
DestRootOverride   string          `json:"destRootOverride,omitempty"`
```

- [ ] **Step 2.2: Verify it compiles**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 2.3: Commit**

```bash
git add types.go
git commit -m "feat: add PreflightResult types and SummaryResult override fields"
```

---

### Task 3: `BrowseForFolder` and `PreflightCheck` in `app.go`

**Files:**
- Modify: `app.go`

- [ ] **Step 3.1: Add `BrowseForFolder` method**

In `app.go`, after the `ValidateBackupRoot` method, add:

```go
// BrowseForFolder opens a native OS directory picker dialog and returns
// the selected folder path, or an empty string if cancelled.
func (a *App) BrowseForFolder(startPath string) (string, error) {
	selected, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		DefaultDirectory: startPath,
		Title:            "Select Folder",
	})
	if err != nil {
		return "", err
	}
	return selected, nil
}
```

- [ ] **Step 3.2: Add `PreflightCheck` method**

After `BrowseForFolder`, add:

```go
// PreflightCheck validates the source root, destination root, and all selected
// item source paths before allowing the user to start migration.
func (a *App) PreflightCheck(selectedIDs []string, sourceRoot string, destRoot string) (PreflightResult, error) {
	cfg, err := config.Load(a.configPath)
	if err != nil {
		return PreflightResult{}, fmt.Errorf("loading config: %w", err)
	}

	result := PreflightResult{}

	// Check 1: source root
	if info, err := os.Stat(sourceRoot); err == nil && info.IsDir() {
		result.SourceRootOK = true
	}

	// Check 2: destination root.
	// DestRootOK = true means migration can proceed:
	//   - blank destRoot → no override, always OK
	//   - destRoot exists as dir → OK (already present)
	//   - destRoot doesn't exist but parent does → OK (will be created at run time; drive is mounted)
	// DestRootOK = false means hard block: the drive/parent is not accessible.
	if destRoot == "" {
		result.DestRootOK = true
	} else {
		if info, err := os.Stat(destRoot); err == nil && info.IsDir() {
			result.DestRootOK = true
		} else {
			parent := filepath.Dir(strings.TrimRight(destRoot, `\/`))
			if info, err := os.Stat(parent); err == nil && info.IsDir() {
				result.DestRootOK = true
			}
		}
	}

	// Build selected ID set
	selectedSet := make(map[string]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		selectedSet[id] = true
	}

	// Check 3: per-item source paths
	for _, app := range cfg.Apps {
		for _, item := range app.Items {
			id := app.Name + ":" + item.Label
			if !selectedSet[id] {
				continue
			}
			result.TotalCount++

			sourcePath := mapper.BuildSourcePath(sourceRoot, item.Source)
			_, statErr := os.Stat(sourcePath)
			found := statErr == nil

			label := app.Name + " \u2014 " + item.Label
			if !found {
				result.HasItemWarnings = true
			} else {
				result.ReadyCount++
			}

			result.Items = append(result.Items, PreflightItem{
				ID:    id,
				Label: label,
				Path:  sourcePath,
				Found: found,
			})
		}
	}

	return result, nil
}
```

The `strings` package is needed for `strings.TrimRight(destRoot, `\/`)` in the parent-directory check of `PreflightCheck` (Task 3). Add `"strings"` to the import block — it will be used there.

- [ ] **Step 3.3: Verify it compiles**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 3.4: Commit**

```bash
git add app.go
git commit -m "feat: add BrowseForFolder and PreflightCheck to App"
```

---

### Task 4: Modify `StartMigration` signature

**Files:**
- Modify: `app.go`

- [ ] **Step 4.1: Update `StartMigration` signature**

Change the signature from:
```go
func (a *App) StartMigration(selectedIDs []string, selectivePaths map[string][]string, dryRun bool) error {
```
to:
```go
func (a *App) StartMigration(selectedIDs []string, selectivePaths map[string][]string, dryRun bool, sourceRootOverride string, destRootOverride string) error {
```

- [ ] **Step 4.2: Capture overrides before the goroutine**

Immediately after the mutex unlock (before `go func() {`), add:

```go
// Capture overrides into locals before the goroutine to avoid closure issues.
srcOverride := sourceRootOverride
dstOverride := destRootOverride
```

Then replace all uses of `sourceRootOverride` and `destRootOverride` inside the goroutine with `srcOverride` and `dstOverride`.

- [ ] **Step 4.3: Apply source override in the runner**

> **Why this works:** `config.Load()` validates that `backup_root` is non-empty in the JSON — this is always true for a valid config file (the field is required at authoring time). The override replaces the value at runtime *after* loading. There is no scenario where a valid config has an empty `backup_root` — the file would fail to load regardless of any override.

Inside the goroutine, after `cfg` is loaded, declare a local copy with the override applied. **Must be a named variable before taking its address** (not an inline literal):

```go
// Shallow copy is safe here: BackupRoot is a string (value type);
// Apps slice header is copied but the underlying array is only read, never mutated.
cfgCopy := *cfg
if srcOverride != "" {
    cfgCopy.BackupRoot = srcOverride
}
r := runner.New(&cfgCopy, rep)
```

- [ ] **Step 4.4: Apply destination override in `runner.go`**

Open `internal/runner/runner.go`. Add the field to `Runner`:

```go
destRootOverride string
```

Add the setter:

```go
// SetDestRootOverride sets a runtime destination root override applied to all target paths.
func (r *Runner) SetDestRootOverride(override string) {
	r.destRootOverride = override
}
```

In `Run()`, find the existing target resolution line:
```go
targetPath := mapper.BuildTargetPath(w.item.Target)
```
Replace it with:
```go
// resolvedTarget is the fully env-var-expanded absolute path from the config.
// If env-var expansion succeeded, it will match X:\ on Windows.
// If it doesn't (e.g. unexpanded %UNKNOWN_VAR%), we treat it as invalid.
resolvedTarget := mapper.BuildTargetPath(w.item.Target)

// Guard: if a dest override is active and the resolved target has no drive prefix,
// log as failed and skip — do not copy to a garbage path.
if r.destRootOverride != "" {
    if len(resolvedTarget) < 3 || resolvedTarget[1] != ':' || resolvedTarget[2] != '\\' {
        r.reportItemFull(w.app.Name, w.item.Label, sourcePath, resolvedTarget, reporter.StatusFailed, 0, "target path has no drive prefix", nil)
        result.Items = append(result.Items, RunResultItem{
            App: w.app.Name, Label: w.item.Label,
            SourcePath: sourcePath, TargetPath: resolvedTarget,
            Status: reporter.StatusFailed,
        })
        r.emitProgress(ProgressEvent{
            Index: i + 1, Total: total,
            App: w.app.Name, Label: w.item.Label,
            Status: "failed", Detail: "target path has no drive prefix",
            Elapsed: time.Since(start).Round(time.Second).String(),
        })
        continue
    }
}
targetPath := mapper.ApplyDestOverride(resolvedTarget, r.destRootOverride)
```

> **Note:** `targetPath` is used in the existing selective-strategy branch as `filepath.Join(targetPath, filepath.Base(p))`. Because `ApplyDestOverride` is applied to `targetPath` before the selective branch, the `filepath.Join` call correctly inherits the overridden root. No further changes needed for selective items.

Back in `app.go` goroutine, after `r.SetDryRun(dryRun)`, add:

```go
r.SetDestRootOverride(dstOverride)
```

- [ ] **Step 4.5: Emit override fields in `SummaryResult`**

In the `StartMigration` goroutine, find the `runtime.EventsEmit(a.ctx, "complete", SummaryResult{...})` call and add the two new fields:

```go
SourceRootOverride: srcOverride,
DestRootOverride:   dstOverride,
```

- [ ] **Step 4.6: Verify it compiles**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 4.7: Run all Go tests**

```bash
go test ./...
```

Expected: all tests pass (runner tests may need updating if they call `StartMigration` directly — check `integration_test.go` and update any call sites to pass two empty strings as the new final args).

- [ ] **Step 4.8: Commit**

```bash
git add app.go internal/runner/runner.go
git commit -m "feat: extend StartMigration with source/dest root overrides"
```

---

---

## Inter-chunk: Regenerate Wails bindings

> Before writing any frontend code, the Wails JS bindings must be regenerated from the updated Go methods added in Chunk 1. The frontend imports `BrowseForFolder`, `PreflightCheck`, and the new `StartMigration` arity from `frontend/wailsjs/go/main/App.js` — this file is auto-generated and must be up-to-date.

- [ ] **Run `wails build` (or `wails dev`) to regenerate bindings**

```bash
cd "F:/GDriveClone/Claude_Code/KtulueKit-Migration"
wails build
```

Expected: build succeeds; `frontend/wailsjs/go/main/App.js` now exports `BrowseForFolder`, `PreflightCheck`, and `StartMigration` with 5 parameters.

---

## Chunk 2: Frontend

### Task 5: `PathBar.svelte` component

**Files:**
- Create: `frontend/src/components/PathBar.svelte`

- [ ] **Step 5.1: Create `PathBar.svelte`**

```svelte
<script>
  import { BrowseForFolder } from '../../wailsjs/go/main/App'

  export let sourceRoot = ''
  export let destRoot = ''

  // icon states: 'blank' | 'unchecked' | 'ok' | 'error'
  let sourceIcon = 'unchecked'
  let destIcon = 'blank'

  // normalise D: → D:\
  function normaliseDrive(path) {
    if (/^[A-Za-z]:$/.test(path)) return path + '\\'
    return path
  }

  import { createEventDispatcher } from 'svelte'
  const dispatch = createEventDispatcher()

  function emitChange() {
    dispatch('change', { sourceRoot, destRoot })
  }

  function handleSourceBlur() {
    sourceRoot = normaliseDrive(sourceRoot)
    sourceIcon = 'unchecked'
    emitChange()
  }

  function handleDestBlur() {
    destRoot = normaliseDrive(destRoot)
    destIcon = destRoot === '' ? 'blank' : 'unchecked'
    emitChange()
  }

  async function browseSource() {
    try {
      const chosen = await BrowseForFolder(sourceRoot)
      if (chosen) {
        sourceRoot = chosen
        sourceIcon = 'unchecked'
        emitChange()
      }
    } catch (e) {
      console.error('Browse failed:', e)
    }
  }

  async function browseDest() {
    try {
      const chosen = await BrowseForFolder(destRoot || '')
      if (chosen) {
        destRoot = chosen
        destIcon = 'unchecked'
        emitChange()
      }
    } catch (e) {
      console.error('Browse failed:', e)
    }
  }

</script>

<div class="path-bar">
  <div class="path-row">
    <span class="path-label">Source:</span>
    <input
      class="path-input"
      type="text"
      bind:value={sourceRoot}
      on:blur={handleSourceBlur}
      placeholder="Backup root (e.g. D:\Backup\W10)"
      spellcheck="false"
    />
    <button class="browse-btn" on:click={browseSource}>Browse</button>
    <span class="icon icon-{sourceIcon}" aria-label={sourceIcon}>
      {#if sourceIcon === 'ok'}✓{:else if sourceIcon === 'error'}⚠{:else if sourceIcon === 'unchecked'}—{/if}
    </span>
  </div>

  <div class="path-row">
    <span class="path-label">Destination:</span>
    <input
      class="path-input"
      type="text"
      bind:value={destRoot}
      on:blur={handleDestBlur}
      placeholder="Override destination root (optional, e.g. D:\)"
      spellcheck="false"
    />
    <button class="browse-btn" on:click={browseDest}>Browse</button>
    <span class="icon icon-{destIcon}" aria-label={destIcon}>
      {#if destIcon === 'ok'}✓{:else if destIcon === 'error'}⚠{:else if destIcon === 'unchecked'}—{/if}
    </span>
  </div>
</div>

<style>
  .path-bar {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 7px 20px;
    background: #141414;
    border-bottom: 1px solid #333;
  }

  .path-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .path-label {
    width: 80px;
    font-size: 12px;
    color: #888;
    flex-shrink: 0;
  }

  .path-input {
    flex: 1;
    background: #2a2a2a;
    color: #e0e0e0;
    border: 1px solid #444;
    border-radius: 4px;
    padding: 4px 8px;
    font-size: 12px;
    font-family: 'Cascadia Code', 'Consolas', monospace;
    min-width: 0;
  }

  .path-input:focus {
    outline: none;
    border-color: #555;
  }

  .browse-btn {
    background: transparent;
    color: #999;
    border: 1px solid #444;
    border-radius: 4px;
    padding: 3px 10px;
    font-size: 12px;
    cursor: pointer;
    flex-shrink: 0;
  }

  .browse-btn:hover { color: #e0e0e0; border-color: #666; }

  .icon {
    width: 18px;
    text-align: center;
    font-size: 13px;
    flex-shrink: 0;
  }

  .icon-ok    { color: #2ea043; }
  .icon-error { color: #d4a017; }
  .icon-unchecked { color: #555; }
  .icon-blank { color: transparent; }
</style>
```

- [ ] **Step 5.2: Verify Wails bindings are regenerated**

After the Wails build process regenerates bindings (happens on `wails dev` or `wails build`), `BrowseForFolder` will be available in `frontend/wailsjs/go/main/App.js`. Since we haven't run a build yet, a temporary import error is expected — it will resolve in Task 8's build step.

- [ ] **Step 5.3: Commit**

```bash
git add frontend/src/components/PathBar.svelte
git commit -m "feat: add PathBar component with source/dest root inputs"
```

---

### Task 6: `PreflightPanel.svelte` component

**Files:**
- Create: `frontend/src/components/PreflightPanel.svelte`

- [ ] **Step 6.1: Create `PreflightPanel.svelte`**

```svelte
<script>
  import { createEventDispatcher } from 'svelte'

  export let result = null  // PreflightResult | null

  const dispatch = createEventDispatcher()

  let expanded = true
  let runAnyway = false

  $: if (result) { runAnyway = false }  // reset on new result

  $: showRunAnyway = result &&
    result.sourceRootOK &&
    result.destRootOK &&
    result.hasItemWarnings

  function handleRunAnyway() {
    dispatch('runAnyway', runAnyway)
  }
</script>

{#if result}
  <div class="preflight-panel">
    <div class="summary-row">
      <span class="summary-text">
        {#if !result.sourceRootOK}
          <span class="err">⚠ Source root not found</span>
        {:else if !result.destRootOK}
          <span class="err">⚠ Destination root not found</span>
        {:else}
          Pre-flight: <strong>{result.readyCount}/{result.totalCount}</strong> ready
          {#if result.hasItemWarnings}
            <span class="warn"> — {result.totalCount - result.readyCount} source path{result.totalCount - result.readyCount !== 1 ? 's' : ''} not found</span>
          {/if}
        {/if}
      </span>

      {#if result.items && result.items.length > 0}
        <button class="toggle-btn" on:click={() => expanded = !expanded}>
          {expanded ? '▲' : '▼'}
        </button>
      {/if}

      {#if showRunAnyway}
        <label class="run-anyway-label">
          <input type="checkbox" bind:checked={runAnyway} on:change={handleRunAnyway} />
          Run anyway
        </label>
      {/if}
    </div>

    {#if expanded && result.items && result.items.length > 0}
      <ul class="item-list">
        {#each result.items.filter(i => !i.found) as item}
          <li class="item-row item-missing">
            <span class="item-icon">↳</span>
            <span class="item-label">{item.label}</span>
            <span class="item-path">[not found at {item.path}]</span>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}

<style>
  .preflight-panel {
    background: #181818;
    border-bottom: 1px solid #333;
    padding: 6px 20px;
    font-size: 12px;
  }

  .summary-row {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .summary-text { color: #999; flex: 1; }
  .err  { color: #e55; }
  .warn { color: #d4a017; }

  .toggle-btn {
    background: transparent;
    border: none;
    color: #555;
    cursor: pointer;
    font-size: 10px;
    padding: 0 4px;
  }

  .run-anyway-label {
    display: flex;
    align-items: center;
    gap: 5px;
    color: #d4a017;
    cursor: pointer;
    font-size: 12px;
  }

  .run-anyway-label input { accent-color: #d4a017; }

  .item-list {
    list-style: none;
    margin: 4px 0 0;
    padding: 0 0 0 16px;
  }

  .item-row {
    display: flex;
    gap: 6px;
    padding: 2px 0;
    color: #d4a017;
  }

  .item-icon { color: #555; }

  .item-label { flex-shrink: 0; }

  .item-path {
    color: #666;
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: 11px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
```

- [ ] **Step 6.2: Commit**

```bash
git add frontend/src/components/PreflightPanel.svelte
git commit -m "feat: add PreflightPanel component"
```

---

### Task 7: Update `SelectionScreen.svelte`

**Files:**
- Modify: `frontend/src/screens/SelectionScreen.svelte`

- [ ] **Step 7.1: Add imports and new props**

At the top of the `<script>` block, add imports:

```js
import PathBar from '../components/PathBar.svelte'
import PreflightPanel from '../components/PreflightPanel.svelte'
import { PreflightCheck } from '../../wailsjs/go/main/App'
```

Add new props after the existing ones:

```js
export let initialSourceRoot = ''
export let initialDestRoot = ''
```

- [ ] **Step 7.2: Remove old backup-root state; add path override and preflight state**

Remove:
```js
// (these are now gone — backup status is replaced by PathBar)
```

Keep `selected`, `profileValue`, `dryRun`. Add:

```js
let sourceRoot = initialSourceRoot || (configView && configView.backupRoot) || ''
let destRoot = initialDestRoot

let preflightResult = null
let preflightDone = false
let runAnyway = false

function resetPreflight() {
  preflightResult = null
  preflightDone = false
  runAnyway = false
}

// Reset preflight whenever paths or selections change
$: sourceRoot, destRoot, selected, resetPreflight()
```

- [ ] **Step 7.3: Remove old backup-banner handlers; add preflight handler**

Remove: `onRefreshBackup` prop (it was for the old banner).

Add:

```js
async function handlePreflight() {
  try {
    preflightResult = await PreflightCheck([...selected], sourceRoot, destRoot)
    preflightDone = true
  } catch (e) {
    console.error('Preflight failed:', e)
  }
}

function handleRunAnyway(e) {
  runAnyway = e.detail
}
```

- [ ] **Step 7.4: Update `handleStart` to pass source/dest roots**

Change:
```js
function handleStart() {
  onStart([...selected], {}, dryRun)
}
```
To:
```js
function handleStart() {
  onStart([...selected], {}, dryRun, sourceRoot, destRoot)
}
```

- [ ] **Step 7.5: Compute `startEnabled`**

Add reactive:

```js
$: startEnabled = selectedCount > 0 &&
  preflightDone &&
  preflightResult &&
  preflightResult.sourceRootOK &&
  preflightResult.destRootOK &&
  (!preflightResult.hasItemWarnings || runAnyway)
```

- [ ] **Step 7.6: Update the template**

Replace the backup banner block (`{#if backupRootValid === false}...{/if}`) with:

```svelte
<PathBar
  {sourceRoot}
  {destRoot}
  on:change={(e) => { sourceRoot = e.detail.sourceRoot; destRoot = e.detail.destRoot }}
/>

<PreflightPanel result={preflightResult} on:runAnyway={handleRunAnyway} />
```

In the footer, replace the start button and its surrounding markup with:

```svelte
<footer>
  <span class="count">{selectedCount} item{selectedCount !== 1 ? 's' : ''} selected</span>
  <div class="footer-actions">
    <button
      class="preflight-btn"
      disabled={selectedCount === 0}
      on:click={handlePreflight}
    >
      Pre-flight Check
    </button>
    <button
      class="start-btn"
      disabled={!startEnabled}
      on:click={handleStart}
    >
      Start Migration
    </button>
  </div>
</footer>
```

- [ ] **Step 7.7: Add `preflight-btn` and `footer-actions` styles**

In the `<style>` block, add:

```css
.footer-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.preflight-btn {
  background: transparent;
  color: #d4a017;
  border: 1px solid #d4a017;
  border-radius: 6px;
  padding: 8px 16px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}

.preflight-btn:hover:not(:disabled) {
  background: rgba(212, 160, 23, 0.1);
}

.preflight-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
```

Remove `onRefreshBackup` from the props list and remove any remaining references to `backupRootValid` and `onRefreshBackup`.

- [ ] **Step 7.8: Commit**

```bash
git add frontend/src/screens/SelectionScreen.svelte
git commit -m "feat: replace backup banner with PathBar and preflight flow in SelectionScreen"
```

---

### Task 8: Update `App.svelte` and wire everything together

**Files:**
- Modify: `frontend/src/App.svelte`

- [ ] **Step 8.1: Update import**

`App.svelte` does not need `ValidateBackupRoot` (removed with the backup banner), `PreflightCheck` (imported directly by `SelectionScreen.svelte`), or `BrowseForFolder` (imported directly by `PathBar.svelte`). Just remove `ValidateBackupRoot`:

Change:
```js
import { GetConfig, StartMigration, GetSourcePath, ValidateBackupRoot } from '../wailsjs/go/main/App'
```
To:
```js
import { GetConfig, StartMigration, GetSourcePath } from '../wailsjs/go/main/App'
```

- [ ] **Step 8.2: Remove `backupRootValid` state; add Run Again override state**

Remove:
```js
let backupRootValid = null
```

Add:
```js
let pendingSourceRoot = ''
let pendingDestRoot = ''
```

- [ ] **Step 8.3: Remove `onMount` backup validation and `handleRefreshBackup`**

In `onMount`, remove:
```js
backupRootValid = await ValidateBackupRoot()
```

Remove the entire `handleRefreshBackup` function.

- [ ] **Step 8.4: Update `handleStartMigration`**

Change:
```js
async function handleStartMigration(selectedIDs, userSelectivePaths, isDryRun) {
  dryRun = isDryRun
  progressEvents = []
  screen = 'progress'
  try {
    await StartMigration(selectedIDs, { ...selectivePaths, ...userSelectivePaths }, isDryRun)
  } catch (err) {
    summaryResult = { failed: [err.toString()], copied: [], skipped: [], manifest: [] }
    screen = 'summary'
  }
}
```
To:
```js
async function handleStartMigration(selectedIDs, userSelectivePaths, isDryRun, sourceRoot, destRoot) {
  dryRun = isDryRun
  progressEvents = []
  screen = 'progress'
  try {
    await StartMigration(selectedIDs, { ...selectivePaths, ...userSelectivePaths }, isDryRun, sourceRoot || '', destRoot || '')
  } catch (err) {
    summaryResult = { failed: [err.toString()], copied: [], skipped: [], manifest: [] }
    screen = 'summary'
  }
}
```

- [ ] **Step 8.5: Update `handleRunAgain`**

Change:
```js
function handleRunAgain() {
  summaryResult = null
  progressEvents = []
  screen = 'selection'
}
```
To:
```js
function handleRunAgain() {
  pendingSourceRoot = summaryResult?.sourceRootOverride ?? ''
  pendingDestRoot = summaryResult?.destRootOverride ?? ''
  summaryResult = null
  progressEvents = []
  screen = 'selection'
}
```

- [ ] **Step 8.6: Update the `<SelectionScreen>` template binding**

Change:
```svelte
<SelectionScreen
  {configView}
  {backupRootValid}
  onStart={handleStartMigration}
  onOpenPicker={handleOpenPicker}
  onProfileChange={handleProfileChange}
  onRefreshBackup={handleRefreshBackup}
/>
```
To:
```svelte
<SelectionScreen
  {configView}
  initialSourceRoot={pendingSourceRoot}
  initialDestRoot={pendingDestRoot}
  onStart={handleStartMigration}
  onOpenPicker={handleOpenPicker}
  onProfileChange={handleProfileChange}
/>
```

- [ ] **Step 8.7: Commit**

```bash
git add frontend/src/App.svelte
git commit -m "feat: wire path override and Run Again flow in App.svelte"
```

---

### Task 9: Build and smoke test

- [ ] **Step 9.1: Run Go tests**

```bash
go test ./...
```

Expected: all pass.

- [ ] **Step 9.2: Build with Wails to regenerate bindings**

```bash
wails build
```

This regenerates `frontend/wailsjs/go/main/App.js` with `BrowseForFolder`, `PreflightCheck`, and the updated `StartMigration` signatures.

Expected: build succeeds, `.exe` produced in `build/bin/`.

- [ ] **Step 9.3: Manual smoke test checklist**

Launch the built `.exe` and verify:

- [ ] Path bar appears with Source pre-populated from config `backup_root`
- [ ] Destination field starts empty
- [ ] Browse button on Source opens native OS folder picker; selected path populates the field
- [ ] Browse button on Destination does the same
- [ ] Typing `D:` then blurring normalises to `D:\`
- [ ] "Pre-flight Check" button is disabled until at least one item is selected
- [ ] After selecting items and clicking Pre-flight, a result appears below the path bar
- [ ] If all paths found: "Start Migration" enables; no "Run anyway"
- [ ] If some paths missing but roots OK: "Run anyway" checkbox appears; Start only enables when checked
- [ ] If source root not found: Start stays disabled; no "Run anyway"
- [ ] Changing a selection after a passing check resets the panel and disables Start
- [ ] Completing a migration and clicking "Run Again" returns to selection with source/dest pre-populated
- [ ] Dry run mode still works end-to-end

- [ ] **Step 9.4: Final commit (if any fixes were needed during smoke test)**

```bash
git add -A
git commit -m "fix: smoke test corrections for path override feature"
```
