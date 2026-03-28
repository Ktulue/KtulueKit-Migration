# Smart Destination Detection

**Date:** 2026-03-28
**Status:** Approved
**Builds on:** Source Discovery (2026-03-22), Path Override & Preflight (2026-03-12)

---

## Problem

The tool copies files to hardcoded target paths defined in the config. When those paths are wrong for the user's machine — especially for games, portable apps, or non-standard installs — the migration silently puts files in a location the app never reads from, making the copy pointless.

Users shouldn't have to know the correct `%APPDATA%` subpath for every app. The tool should figure it out.

## Solution

A two-tier destination resolution system that infers where files should go based on where they came from, with fallback detection for apps that don't follow standard patterns.

---

## Design

### Tier 1: Path Pattern Mapping (Automatic)

When source discovery finds files at a path like `E:\Users\Josh\AppData\Roaming\obs-studio\basic`, the tool:

1. Extracts known path segments from the source path:
   - `Users\<name>\AppData\Roaming\...`
   - `Users\<name>\AppData\Local\...`
   - `Users\<name>\Documents\...`
   - `Users\<name>\Desktop\...`
   - `Users\<name>\Videos\...`
   - `Users\<name>\Pictures\...`
   - `Users\<name>\Music\...`
   - `Users\<name>\.ssh\...`
   - `Users\<name>\.gitconfig` (and similar dotfiles)
2. Replaces the drive letter and username with the current machine's equivalents (current user's `%USERPROFILE%`, local drive).
3. Checks if the resolved local path exists on disk.
4. If it exists, returns it as the confirmed destination.

**Coverage:** ~80% of configured apps — anything that stores data in standard Windows user profile locations.

**No config changes required** for Tier 1. It works purely from the discovered source path structure.

### Tier 2: App Detection Hints (Config-Driven)

For apps where Tier 1 fails (custom install locations, games, portable apps), the config provides an optional `detection` block at the `App` level with fallback detection strategies.

**Resolution order within Tier 2:**

1. **Registry lookup** — Query a Windows registry key for the app's install or data path.
2. **Search paths** — Scan common directories across all available drives for a known folder name.
3. **Executable confirmation** — If a folder is found via registry or search, verify it contains the expected executable.

If Tier 2 also fails, the item is marked "not found" and the user is prompted to browse manually.

### Resolution Priority

```
Source path discovered
    |
    v
Tier 1: Pattern map source path -> local equivalent
    |
    +--> Path exists? --> YES --> Destination confirmed
    |
    +--> NO --> Tier 2 detection block exists?
                    |
                    +--> YES --> Try registry, then search paths
                    |               |
                    |               +--> Found? --> Destination confirmed
                    |               |
                    |               +--> Not found --> Mark "not found", prompt browse
                    |
                    +--> NO --> Fall back to config target field
                                    |
                                    +--> Path exists? --> YES --> Destination confirmed
                                    |
                                    +--> NO --> Mark "not found", prompt browse
```

---

## Config Schema Changes

An optional `detection` block is added to the `App` level in `ktuluekit-migration.json`. Apps without it rely entirely on Tier 1 and the existing `target` field fallback.

```json
{
  "name": "LurkBait Twitch Fishing",
  "category": "Games",
  "detection": {
    "registry": "HKCU\\Software\\LurkBait\\InstallPath",
    "searchPaths": ["Program Files", "Program Files (x86)", "SteamLibrary/steamapps/common"],
    "searchTarget": "LurkBait",
    "executable": "LurkBait.exe"
  },
  "items": [...]
}
```

**Fields (all optional within the block):**

| Field | Type | Purpose |
|-------|------|---------|
| `registry` | `string` | Windows registry key to query for install/data path |
| `searchPaths` | `[]string` | Directories to scan (relative to drive roots, tool checks all available drives) |
| `searchTarget` | `string` | Folder name to look for within search paths |
| `executable` | `string` | Exe name to confirm the correct folder was found |

**Rules:**
- The `detection` block itself is optional — most apps won't need one.
- Within the block, all fields are optional. Strategies are tried in field order: registry first, then search paths.
- `executable` is a confirmation check, not a primary search strategy.
- Existing `target` fields on items are unchanged and serve as the final fallback.
- The JSON schema (`schema/ktuluekit-migration.schema.json`) is updated to include the new fields.

**Which apps need detection blocks:**

Only apps where Tier 1 path mapping won't work — primarily games and apps installed to non-standard locations. For the current config, candidates include:
- LurkBait Twitch Fishing
- DaVinci Resolve (if installed outside default)
- Any future game or portable app additions

Standard `%APPDATA%`/`%LOCALAPPDATA%` apps (OBS, VS Code, Discord, etc.) do not need detection blocks.

---

## Backend Architecture

### New Package: `internal/detector/`

Single responsibility: given a discovered source path and optional detection hints, resolve the destination path on the local machine.

**Types:**

```go
// Detection holds app-level detection hints from config.
type Detection struct {
    Registry     string   `json:"registry,omitempty"`
    SearchPaths  []string `json:"searchPaths,omitempty"`
    SearchTarget string   `json:"searchTarget,omitempty"`
    Executable   string   `json:"executable,omitempty"`
}

// DetectRequest is the input for a single app's detection.
type DetectRequest struct {
    AppName    string
    SourcePath string              // Discovered source path
    Detection  *config.Detection   // Optional detection block (nil for most apps)
}

// DetectResult is the output for a single item's detection.
type DetectResult struct {
    ItemID     string   // "AppName:ItemLabel"
    DestPath   string   // Resolved destination (empty if not found)
    Method     string   // "path-mapping", "registry", "search", "manual"
    Confirmed  bool     // Does the resolved path exist on disk?
    Candidates []string // Multiple matches for user to pick from
}
```

**Core functions:**

- `PatternMap(sourcePath string) (string, bool)` — Tier 1: extract known path segments, remap to local user, check existence.
- `RegistryLookup(key string) (string, error)` — Tier 2a: query Windows registry.
- `SearchForApp(searchPaths []string, target string, exe string) ([]string, error)` — Tier 2b: scan directories across drives, return candidates.
- `Detect(req DetectRequest) DetectResult` — Orchestrator: try Tier 1, fall back to Tier 2, return result.

### Wails-Bound Method

```go
func (a *App) DetectDestination(appName string, sourcePathMap map[string]string) []DetectResult
```

Takes an app name and its discovered source paths (from source discovery), returns detection results per-item. Called from the frontend when the user clicks "Detect" on an app.

### Changes to Existing Code

| File | Change |
|------|--------|
| `internal/config/config.go` | Add optional `Detection` field to `App` struct |
| `app.go` | Add `DetectDestination()` Wails-bound method |
| `internal/runner/runner.go` | Accept `destPathMap map[string]string` alongside existing `sourcePathMap`; use detected destinations when available |
| `types.go` | Add `DetectResult` to display types |
| `schema/ktuluekit-migration.schema.json` | Add `detection` object schema to app definition |
| `ktuluekit-migration.json` | Add `detection` blocks to apps that need them |

### Runner Integration

The `Runner` struct gets a new `destPathMap` field (same pattern as `sourcePathMap`):

```go
type Runner struct {
    // ... existing fields ...
    destPathMap map[string]string // Per-item destination overrides from detection
}
```

Resolution order in `Run()`:
1. Check `destPathMap[itemID]` — if present, use it
2. Fall back to `ApplyDestOverride(resolvedTarget, destRootOverride)` (existing global override)
3. Fall back to `BuildTargetPath(item.Target)` (existing config target)

---

## Frontend Changes

### Per-App "Detect" Button

Each app row in `CategoryAccordion` / `ItemRow` gets a "Detect" button, visible when source discovery has found items for that app.

**Trigger:** User clicks "Detect" on an app row.
**Action:** Calls `DetectDestination(appName, sourcePathMap)` on the backend.
**Result:** Per-item detection results displayed inline.

### Item Row Changes

After detection, each `ItemRow` shows:

- **Destination path** — displayed beneath the item name as a clickable/editable field
- **Badge:**
  - Green **"confirmed"** — Tier 1 or Tier 2 found the path and it exists
  - Yellow **"not found"** — detection failed, user needs to set manually
  - Blue **"manual"** — user overrode the detected path via browse
- **Browse button** — always available to override the destination
- **Multiple candidates** — if detection found multiple matches, show a dropdown for the user to pick

### Detection State

Frontend maintains a `destMap` (parallel to existing `discoveryMap`):

```typescript
// Map of itemID -> detected destination path
let destMap: Record<string, string> = {};
```

Passed to `PreflightCheck()` and `StartMigration()` alongside the existing `sourcePathMap`.

### What Stays the Same

- Source discovery works exactly as today
- Global destination override in PathBar remains as a quick shortcut
- Per-item detection takes precedence over global override
- Preflight check works the same but now also validates per-item destinations from `destMap`

---

## Testing Strategy

### Unit Tests (`internal/detector/`)

- `PatternMap` correctly remaps known path patterns (AppData\Roaming, AppData\Local, Documents, etc.)
- `PatternMap` returns empty for unrecognized path structures
- `PatternMap` handles different drive letters and usernames
- `RegistryLookup` returns path for existing keys, error for missing keys
- `SearchForApp` finds target folders in search paths
- `SearchForApp` validates with executable when provided
- `SearchForApp` returns multiple candidates when found
- `Detect` tries Tier 1 first, falls back to Tier 2
- `Detect` returns "not found" when both tiers fail

### Integration Tests

- End-to-end detection with real filesystem paths (using temp directories)
- Runner uses `destPathMap` correctly when present
- Runner falls back to existing target resolution when `destPathMap` is empty
- Config loading with and without `detection` blocks

### Manual Testing

- Run with cloned drive, verify Tier 1 auto-maps standard app paths
- Test LurkBait or similar game with detection block, verify Tier 2 kicks in
- Test manual override flow when detection fails
- Verify preflight validates detected destinations

---

## Future Enhancements Roadmap

Prioritized in order of expected value:

### Phase 2: Batch Detection
- "Detect All" button that runs detection for all selected apps in one pass
- Progress indicator for batch detection
- Summary of results: X confirmed, Y not found, Z need manual input

### Phase 3: Auto-Detect on Source Scan
- When source discovery runs, automatically trigger destination detection for all found apps
- Eliminates the manual "Detect" click per app
- Depends on batch detection being reliable

### Phase 4: Persistence
- Save detected/confirmed destination paths to a `ktuluekit-migration.local.json` file
- Load as defaults on next session, re-validate on startup
- Main config stays clean and shareable

### Phase 5: Detection for Unsourced Apps
- Allow detection to run even without a discovered source
- Useful when the user manually sets source paths but wants smart destination resolution
- Falls back to config `target` field for Tier 1 pattern extraction

### Phase 6: Community Detection Database
- Crowdsourced registry keys and search paths for common apps
- Users can contribute detection blocks for apps they've successfully migrated
- Pull from a remote database or accept PR contributions to the config
