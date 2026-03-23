# Source Discovery Design

**Date:** 2026-03-22
**Status:** Approved
**Branch:** `feat/source-discovery`

## Problem

The app expects a pre-organized backup folder with a specific directory structure matching the config. Users who have a full drive clone (e.g., a W10 C:\ cloned to E:\) have all the data but not in the expected layout. Manually locating each app's data across a cloned drive is tedious and error-prone.

Additionally, the checkboxes in the selection screen are nearly invisible on the dark background (native browser checkboxes with only `accent-color` set).

## Solution

Auto-scan a source drive to discover where each config item's data actually lives, then present results with assisted fallback for items that couldn't be found automatically. Also fix checkbox visibility.

## User Flow

1. User sets Source path to a drive root (e.g., `E:\`)
2. User clicks a **"Scan"** button (explicit trigger, no auto-scan on blur)
3. Scanning indicator shown while scan runs
4. Selection screen shows results inline:
   - **Found items**: green indicator, pre-checked, resolved source path stored in memory
   - **Not-found items**: greyed out with "Assist" button
5. Clicking "Assist" on a not-found item opens a browse dialog starting at the source drive, or shows partial-match suggestions if the scanner found near-misses
6. User proceeds with normal preflight → migration flow

## Backend

### New package: `internal/discovery`

**`Scan(ctx context.Context, drivePath string, cfg *config.Config) (*Result, error)`**

Algorithm:

1. Check if `<drivePath>\Users\` exists. If not, return all items as not-found with `FoundCount: 0` (graceful empty result, not an error).
2. Enumerate subdirectories of `Users\`, filter out system profiles (`Default`, `Public`, `All Users`, `Default User`). Skip any directory that returns a permission error (log and continue).
3. For each real user profile found, build an env var mapping for the cloned drive:
   - `%APPDATA%` → `<drivePath>\Users\<name>\AppData\Roaming`
   - `%LOCALAPPDATA%` → `<drivePath>\Users\<name>\AppData\Local`
   - `%USERPROFILE%` → `<drivePath>\Users\<name>`
   (Note: `%USERPROFILE%` expansion already covers `AppData\LocalLow` paths like LurkBait's target `%USERPROFILE%/AppData/LocalLow/...`)
4. For each config item, take its **`target` field** (not `source` — the target contains the env vars that tell us where the data lives in a Windows install), resolve it using the cloned-drive env mapping, and check if that resolved path exists on the source drive.
5. **Profile selection**: score each profile by number of items found. The profile with the highest match count wins. Ties broken by most recent directory modification time.
6. If `Users\` doesn't exist or no profiles found, return all items as not-found. (No fallback deep-scan — keep it simple and fast. The "Assist" button handles manual location for edge cases.)

**Return type:**
```go
type Result struct {
    Items      []DiscoveredItem
    FoundCount int
    TotalCount int
}

type DiscoveredItem struct {
    ID         string   // "OBS Studio:scenes & profiles"
    Label      string   // "OBS Studio — scenes & profiles"
    SourcePath string   // absolute path on the source drive (empty if not found)
    Found      bool
    Partial    []string // near-miss paths for assist suggestions (empty if found)
}
```

### New Wails bindings on `*App`

```go
ScanDrive(drivePath string) (*discovery.Result, error)
```

Calls `discovery.Scan(ctx, drivePath, app.cfg)` and returns the result. Context comes from the Wails runtime context for potential cancellation.

### Integration with existing engine — per-item source path overrides

**This is the critical integration point.** The existing `sourceRootOverride` is a single scalar applied uniformly to all items. Discovery produces per-item absolute paths. These are fundamentally incompatible.

**Solution: new `sourcePathMap` parameter.**

Add a new parameter to `StartMigration`:
```go
func (a *App) StartMigration(
    selectedIDs []string,
    selectivePaths map[string][]string,
    dryRun bool,
    sourceRootOverride string,
    destRootOverride string,
    sourcePathMap map[string]string,  // NEW: itemID → absolute source path
) error
```

The runner checks `sourcePathMap[itemID]` first. If present, use that absolute path as the source. Otherwise, fall back to the existing `BuildSourcePath(backupRoot, item.Source)` logic. This preserves full backward compatibility — when `sourcePathMap` is nil or empty, behavior is identical to today.

**PreflightCheck also needs the same parameter:**
```go
func (a *App) PreflightCheck(
    selectedIDs string,
    sourceRoot string,
    destRoot string,
    sourcePathMap map[string]string,  // NEW
) (PreflightResult, error)
```

When checking each item's source existence, if `sourcePathMap[itemID]` exists, check that path instead of `BuildSourcePath(sourceRoot, item.Source)`.

## Frontend

### Selection screen changes

- **"Scan" button** in the PathBar next to Source Browse button. Explicit click to trigger scan (no auto-trigger on blur — avoids debounce/race issues).
- Show a scanning indicator (spinner or progress text) while scan runs
- Store `discoveryMap: Record<string, DiscoveredItem>` in component state
- Each `ItemRow` receives discovery status:
  - **Found**: visible check indicator, item is pre-checked, tooltip shows resolved source path
  - **Not found**: dimmed label, "Assist" button that opens `BrowseForFolder(drivePath)` — user picks the folder manually, which updates `discoveryMap` for that item
- When starting migration/preflight, build `sourcePathMap` from `discoveryMap` and pass it through

### Checkbox visibility fix (bundled opportunistic fix)

Replace the nearly-invisible native `<input type="checkbox">` with custom-styled checkboxes that have:
- Visible border in unchecked state (e.g., `1px solid var(--color-text-secondary)`)
- Filled accent-color background with checkmark in checked state
- Proper sizing for the dark theme

## Scope boundaries

- Discovery is **read-only** — never writes to the source drive or modifies the config file
- Discovered paths are held in frontend state for the session only
- No profile picker UI — profile detection is an internal implementation detail
- System-level paths (`%ProgramData%`, etc.) are out of scope — all current config items use user-profile paths
- Network/UNC paths are out of scope — this targets local drive clones
