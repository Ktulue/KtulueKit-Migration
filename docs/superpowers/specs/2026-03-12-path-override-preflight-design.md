# Path Override & Pre-flight Check — Design Spec

**Date:** 2026-03-12
**Status:** Approved

---

## Problem

Drive letters and names may differ between Windows 10 and Windows 11 setups. The current `backup_root` and item target paths are hardcoded in the JSON config, forcing users to edit the file when drive letters change. Users need a way to override source and destination roots at runtime without modifying the config.

---

## Solution Overview

Add a **path bar** to the SelectionScreen with inline source/destination root overrides, plus a **pre-flight check** that validates all selected items before enabling "Start Migration." Both the override paths and pre-flight results are logged in the run manifest.

---

## UI Layout

The existing backup banner is replaced by a two-row **path bar** between the header and item list:

```
[ Source:      D:\Backup\W10              ] [Browse] [✓]
[ Destination: C:\                        ] [Browse] [✓]
```

- **Source** pre-populates from `backup_root` in the config JSON
- **Destination** pre-populates empty (blank = no override)
- **Browse** opens the existing FolderPicker component
- **Status icon** per row — four distinct states:
  - `✓` green — path confirmed to exist and is a directory
  - `⚠` yellow — path checked and not found
  - `—` grey — not yet checked (user has typed but not confirmed)
  - *(no icon)* — destination field is intentionally blank

---

## Destination Root Logic

Applies only to Windows single-letter drive paths (`X:\`). UNC paths are not supported.

**Normalisation:** if the user types `D:` (without trailing `\`) in either field, it is normalised to `D:\` on blur/confirm before any processing. Applies to both source and destination fields.

**Detection rule** (applied to `destRoot` after normalisation):
- `len(destRoot) == 3` (e.g. `D:\`) → **drive swap**: replace first character of resolved target with the new drive letter
- Otherwise → **prefix substitution**: strip first 3 characters (`X:\`) from resolved target and prepend `destRoot`
- blank → **no override**

| User sets | Resolved target example | Result |
|-----------|------------------------|--------|
| `D:\` | `C:\Users\Foo\AppData\Roaming\App` | `D:\Users\Foo\AppData\Roaming\App` |
| `D:\Restored\` | `C:\Users\Foo\AppData\Roaming\App` | `D:\Restored\Users\Foo\AppData\Roaming\App` |
| *(blank)* | `C:\Users\Foo\AppData\Roaming\App` | unchanged |

**Edge case — resolved target has no drive prefix:** if `resolvedTarget` does not begin with the pattern `X:\` (letter + `:\`), `ApplyDestOverride` returns it unchanged and logs a warning. This handles the case where env-var expansion fails and leaves a relative path (which is a config error, not a runtime override error).

**Coverage:** `ApplyDestOverride` is applied to all resolved target paths uniformly, including those resolved from env vars (e.g. `%APPDATA%`). This is intentional.

---

## Pre-flight Check

A **"Pre-flight Check"** button sits in the footer beside "Start Migration."

### What it checks

1. **Source root** — `sourceRoot` exists and `os.Stat` returns a directory → `SourceRootOK`
2. **Destination root** — if `destRoot` is non-empty:
   - `destRoot` itself exists as a directory → `DestRootOK = true`
   - `destRoot` does not exist but its parent does → `DestRootOK = true` (directory will be created at run time)
   - Neither exists → `DestRootOK = false`
   - `destRoot` is blank → `DestRootOK = true` always
3. **Per-item sources** — for each selected item, resolve source path. If `item.Source` is relative, call `BuildSourcePath(sourceRoot, item.Source)`. If `item.Source` resolves to an absolute path (i.e. `filepath.IsAbs` after env-var expansion), the source root override has no effect and the item's absolute path is used as-is; the pre-flight check checks that absolute path and the result panel shows the actual path checked so the user can see what was used.

No side effects (no `MkdirAll` during pre-flight).

### `PreflightResult` type

```go
type PreflightResult struct {
    SourceRootOK    bool            `json:"sourceRootOK"`
    DestRootOK      bool            `json:"destRootOK"`
    HasItemWarnings bool            `json:"hasItemWarnings"` // true if any item.Found == false
    Items           []PreflightItem `json:"items"`
    ReadyCount      int             `json:"readyCount"`
    TotalCount      int             `json:"totalCount"`
}

type PreflightItem struct {
    ID    string `json:"id"`
    Label string `json:"label"`  // "AppName — ItemLabel"
    Path  string `json:"path"`   // resolved source path that was actually checked
    Found bool   `json:"found"`
}
```

### Button and state machine

| Condition | "Start Migration" | "Run anyway" shown |
|-----------|-------------------|--------------------|
| No check run | Disabled | No |
| `SourceRootOK && DestRootOK && !HasItemWarnings` | Enabled | No |
| `SourceRootOK && DestRootOK && HasItemWarnings` | Disabled until acknowledged | Yes |
| `SourceRootOK && DestRootOK && HasItemWarnings && runAnyway` | Enabled | Yes (checked) |
| `!SourceRootOK \|\| !DestRootOK` | Disabled | No |

**Reset rule:** pre-flight state (`preflightResult`, `preflightDone`, `runAnyway`) is cleared and "Start Migration" is disabled whenever:
- Either path field is blurred/confirmed (including after Browse)
- Any item selection changes (individual toggle, profile load, or selective-picker confirmation)

Reset does NOT trigger on every keystroke — only on field blur/confirm or selection mutation.

**Reactive implementation in `SelectionScreen.svelte`:** use a Svelte reactive statement `$: resetPreflight(sourceRoot, destRoot, selected)` where `resetPreflight` sets the three state variables to their cleared values. This covers all selection-change paths (individual toggle via `handleToggle`, profile load via `loadProfile`, and selective-picker confirm via `handleOpenPickerWrapped`) because all of them mutate `selected`.

---

## Backend Changes

### New Go method: `PreflightCheck`

```go
func (a *App) PreflightCheck(selectedIDs []string, sourceRoot string, destRoot string) (PreflightResult, error)
```

`PreflightCheck` loads config internally (same as `GetConfig` and `StartMigration` do). `PreflightItem.Label` is assembled as `app.Name + " — " + item.Label` — matching the format used in `GetConfig`'s `ItemView.Name` so panel labels are consistent with the selection list.

### Modified: `StartMigration`

New signature (empty string = no override):

```go
func (a *App) StartMigration(
    selectedIDs []string,
    selectivePaths map[string][]string,
    dryRun bool,
    sourceRootOverride string,
    destRootOverride string,
) error
```

**Source override application:** wherever `StartMigration` calls `mapper.BuildSourcePath(cfg.BackupRoot, item.Source)`, it substitutes `sourceRootOverride` for `cfg.BackupRoot` when `sourceRootOverride` is non-empty. For items whose `item.Source` resolves to an absolute path, the override is silently ignored (same behaviour as pre-flight check).

**Destination override application:** wherever `StartMigration` calls `mapper.BuildTargetPath(item.Target)`, it wraps the result: `mapper.ApplyDestOverride(mapper.BuildTargetPath(item.Target), destRootOverride)`.

**Manifest fields:** `StartMigration` writes `sourceRootOverride` and `destRootOverride` into the `SummaryResult` emitted on the Wails `"complete"` event.

**Frontend call site:** `App.svelte`'s `handleStartMigration` function must be updated to accept `sourceRoot` and `destRoot` parameters and pass them as the 4th and 5th arguments to the regenerated Wails binding for `StartMigration`.

### New mapper function: `ApplyDestOverride`

```go
// ApplyDestOverride rewrites resolvedTarget according to destRoot.
// Empty destRoot returns resolvedTarget unchanged.
// len(destRoot)==3 (e.g. "D:\") swaps the drive letter (first character).
// Otherwise strips the "X:\" prefix from resolvedTarget and prepends destRoot.
// If resolvedTarget does not begin with "X:\" pattern, returns it unchanged.
func ApplyDestOverride(resolvedTarget, destRoot string) string
```

**Guard ordering inside `ApplyDestOverride` (must be implemented in this order):**
1. If `destRoot` is empty → return `resolvedTarget` unchanged
2. If `resolvedTarget` does not match `X:\...` (i.e. `len(resolvedTarget) < 3` or `resolvedTarget[1] != ':'` or `resolvedTarget[2] != '\\'`) → return `resolvedTarget` unchanged
3. If `len(destRoot) == 3` → drive swap: return `string(destRoot[0]) + resolvedTarget[1:]`
4. Otherwise → strip first 3 chars and prepend: return `destRoot + resolvedTarget[3:]`

### New types in `types.go`

```go
type PreflightResult struct { ... }
type PreflightItem struct { ... }
```

### Modified: `SummaryResult`

```go
SourceRootOverride string `json:"sourceRootOverride,omitempty"`
DestRootOverride   string `json:"destRootOverride,omitempty"`
```

---

## Frontend Changes

### `SelectionScreen.svelte`

- Replaces backup banner with `PathBar` component
- Holds `preflightResult: PreflightResult | null`, `preflightDone: bool`, `runAnyway: bool`
- Reactive reset: `$: resetPreflight(sourceRoot, destRoot, selected)` — clears all three on any change
- "Pre-flight Check" button calls `PreflightCheck([...selected], sourceRoot, destRoot)`
- "Start Migration" enabled when: `preflightDone && preflightResult.sourceRootOK && preflightResult.destRootOK && (!preflightResult.hasItemWarnings || runAnyway)`
- Passes `sourceRoot`, `destRoot` to `handleStart`, which forwards them to `StartMigration`

### New component: `PathBar.svelte`

Props:
- `sourceRoot: string` (pre-populated from `configView.backupRoot`)
- `destRoot: string` (starts `""`)

Emits `on:change { sourceRoot, destRoot }` on blur/confirm. Each row: label + text input + Browse button + status icon. Browse opens FolderPicker; on confirm, sets field value and triggers blur. On blur, normalise `D:` → `D:\` then emit change. Status icon uses four-state enum: `blank` (no icon), `unchecked` (`—`), `ok` (`✓`), `error` (`⚠`).

### New component: `PreflightPanel.svelte`

Props:
- `result: PreflightResult | null`

Collapsible per-item list. Summary line: `N/M ready`. Shows "Run anyway" checkbox only when `result.sourceRootOK && result.destRootOK && result.hasItemWarnings`. Emits `on:runAnyway: bool`.

### Modified: `App.svelte`

- `handleStartMigration(selectedIDs, selectivePaths, dryRun, sourceRoot, destRoot)` — passes `sourceRoot` and `destRoot` to the updated `StartMigration` binding
- Stores received `SummaryResult` in a reactive variable `summaryResult`
- `handleRunAgain` reads `summaryResult.sourceRootOverride` and `summaryResult.destRootOverride` from its own scope (no argument changes to `SummaryScreen`'s `onRunAgain` callback); transitions back to `SelectionScreen` passing the two values as props

### Modified: `SelectionScreen.svelte` props

Add two optional props for pre-population:
```js
export let initialSourceRoot = ''   // pre-populated on Run Again
export let initialDestRoot = ''     // pre-populated on Run Again
```
`PathBar` receives `sourceRoot = initialSourceRoot || configView.backupRoot` and `destRoot = initialDestRoot`.

---

## Out of Scope

- Persisting override paths across app restarts (per-session only)
- UNC / network path support
- Per-item destination override
- Modifying the config JSON file
