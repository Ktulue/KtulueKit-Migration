# KtulueKit-Migration: Build-Out Design Spec
**Date:** 2026-03-11
**Status:** Approved
**Author:** Ktulue

---

## Overview

KtulueKit-Migration is the second tool in the KtulueKit trilogy (W11 → Migration → Cleanup). It restores personal app state from a backup location to a fresh Windows 11 install. The tool is purely additive — it never deletes anything. Destructive operations belong to KtulueKit-Cleanup.

The scaffold is complete. This spec defines the full build-out to a shippable, polished tool.

**Tech stack:** Go + Wails v2 + Svelte. Same patterns as KtulueKit-W11.

---

## Definition of Done

The tool is done when:
1. It runs end-to-end without errors
2. All personal apps and files are represented in the config
3. The folder picker feature works for personal file directories
4. The UI is visually aligned with KtulueKit-W11
5. A full migration has been tested against real backup data on W11
6. The manifest contract is locked and the `.exe` is built

---

## Build Order

### Phase 1 — Dev Run & Wiring Verification

**Goal:** Confirm the full 3-screen flow runs without crashes.

**Done when:**
- `wails dev` opens the window and SelectionScreen loads the config correctly
- Wails JS bindings (`frontend/wailsjs/`) are generated and up to date
- `GetConfig` and `StartMigration` calls work from the frontend
- A migration run completes and lands on SummaryScreen with manifest entries
- Log file is written to `logs/`

**Note:** When new Go methods are added in later phases (e.g. `ListFolder` in Phase 3), `wails dev` must be restarted to regenerate bindings before the Svelte side can call them.

**Scope:** Fix wiring bugs only. No polish, no new features.

---

### Phase 2 — Config Expansion

**Goal:** Every app the user needs restored is in `ktuluekit-migration.json`.

**Apps to add:**

#### Dev Tools
| App | Label | Source | Target |
|-----|-------|--------|--------|
| Git | global config | `git/.gitconfig` | `%USERPROFILE%/.gitconfig` |
| SSH | keys & config | `ssh` | `%USERPROFILE%/.ssh` |
| VS Code | settings & keybindings | `vscode/User` | `%APPDATA%/Code/User` |
| Windows Terminal | settings | `terminal/settings.json` | `%LOCALAPPDATA%/Packages/Microsoft.WindowsTerminal_8wekyb3d8bbwe/LocalState/settings.json` |

> **Note — Windows Terminal path:** The `8wekyb3d8bbwe` suffix is the Store publisher ID and is stable. This path only applies when Terminal is installed via the Microsoft Store or WinGet (which installs the Store version). Note this in the config entry's `notes` field.

#### Creative Suite
| App | Label | Source | Target |
|-----|-------|--------|--------|
| GIMP | brushes & settings | `gimp` | `%APPDATA%/GIMP` |
| Krita | brushes & settings | `krita` | `%APPDATA%/krita` |
| Audacity | settings | `audacity` | `%APPDATA%/audacity` |
| Blender | preferences & addons | `blender` | `%APPDATA%/Blender Foundation` |
| Aseprite | settings | `aseprite` | `%APPDATA%/Aseprite` |
| Kdenlive | settings | `kdenlive` | `%APPDATA%/kdenlive` |
| DaVinci Resolve | project library | `resolve` | `%USERPROFILE%/Documents/DaVinci Resolve` |

#### Utilities
| App | Label | Source | Target |
|-----|-------|--------|--------|
| ShareX | settings & screenshots | `sharex` | `%APPDATA%/ShareX` |
| PowerToys | settings | `powertoys` | `%LOCALAPPDATA%/Microsoft/PowerToys` |
| Claude Desktop | MCP config | `claude-desktop` | `%APPDATA%/Claude` |

#### Communication
| App | Label | Source | Target |
|-----|-------|--------|--------|
| Discord | local settings | `discord` | `%APPDATA%/discord` |
| Spotify | local files | `spotify` | `%LOCALAPPDATA%/Spotify` |

#### Personal Files
| App | Label | Source | Target | Strategy |
|-----|-------|--------|--------|----------|
| Personal Files | Desktop | `personal/Desktop` | `%USERPROFILE%/Desktop` | `mirror` |
| Personal Files | Documents | `personal/Documents` | `%USERPROFILE%/Documents` | `selective` |
| Personal Files | Videos | `personal/Videos` | `%USERPROFILE%/Videos` | `selective` |
| Personal Files | Pictures | `personal/Pictures` | `%USERPROFILE%/Pictures` | `selective` |
| Personal Files | Music | `personal/Music` | `%USERPROFILE%/Music` | `selective` |

**OBS Studio & Streamer.bot note:** Already in the config. Both apps store all config as plain files — no special export or CLI step needed. Neither app needs to be running for backup or restore. The only requirement is that both apps are **closed** before the backup step runs, to avoid SQLite file locks on Streamer.bot's `.db` files. Add this to each entry's `notes` field.

**Done when:**
- All apps above are in the config with correct source/target paths
- Profiles (`Full Restore`, `Streaming Only`, etc.) updated to include new entries
- New categories added to `categoryOrder` in `types.go` (source of truth — not just the spec)

---

### Phase 3 — Folder Picker Feature

**Goal:** Personal Files items with `strategy: "selective"` open a folder browser instead of mirroring blindly.

#### Go changes

**New `ItemView` field:**
```go
type ItemView struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Notes       string `json:"notes"`
    Strategy    string `json:"strategy"` // "mirror" | "file" | "selective"
}
```
`app.go`'s `GetConfig` must populate `Strategy` from the config item.

**New method on App:**
```go
func (a *App) ListFolder(path string) ([]FolderEntry, error)

type FolderEntry struct {
    Name  string `json:"name"`
    Path  string `json:"path"`
    IsDir bool   `json:"isDir"`
    Size  int64  `json:"size"`
}
```
Returns immediate (non-recursive) contents of a directory. Used by the FolderPicker component.

**Updated `StartMigration` signature:**
```go
func (a *App) StartMigration(selectedIDs []string, selectivePaths map[string][]string) error
```
`selectivePaths` maps item ID → list of absolute paths the user selected in the picker. For non-selective items, their ID is absent from this map and the runner uses MirrorDir as normal.

**New runner strategy:**
When `strategy == "selective"`, the runner iterates `selectivePaths[itemID]` and calls `copier.CopyPath` on each. `CopyPath` handles both files and directories. MirrorDir is used only for `strategy == "mirror"`.

#### Svelte changes

**`App.svelte` update:**
- Add `selectivePaths = {}` state (object mapping item ID → selected path array)
- Pass `selectivePaths` to `StartMigration`
- Pass a `onPickerConfirm` callback down to SelectionScreen

**New `FolderPicker.svelte` component:**
- Modal that opens when a `selective` item's checkbox is checked
- Calls `ListFolder(sourcePath)` to populate contents
- Checkbox per entry (file or subfolder), with select all / deselect all
- Confirm closes the modal and stores selections in `selectivePaths`
- Cancel unchecks the item

> **UX note:** The picker opens on checkbox check. If the user checks the box and cancels the picker, the item is unchecked. This makes the checkbox state authoritative — a checked selective item always has confirmed selections.

**`ItemRow.svelte` update:**
- Receives `strategy` prop from `ItemView`
- For `selective` items, checking triggers the picker via a callback rather than directly toggling the Set

**Done when:**
- `ListFolder` Go method works and returns correct entries
- `ItemView.Strategy` is populated and passed to the frontend
- FolderPicker modal opens, populates, allows selection, and stores correctly
- `StartMigration` accepts and uses `selectivePaths`
- Runner copies only selected paths for selective strategy
- Desktop still mirrors as a full item (no picker)
- Wails bindings regenerated after `ListFolder` is added

---

### Phase 4 — UI Polish

**Goal:** Visual alignment with KtulueKit-W11 patterns, plus dry-run mode.

**Visual changes:**
- Tighten spacing to match W11 (`padding: 12px 20px`, pixel values over rem)
- Review font sizes for consistency with W11
- FolderPicker modal follows the same dark theme
- Any rough edges identified during Phase 1-3 work

**Dry-Run Mode:**

A toggle in the SelectionScreen header. When enabled:
- ProgressScreen shows what would be copied without writing anything
- Summary shows estimated sizes per item
- No files written, no log created
- Clear visual banner indicating dry-run mode

**Implementation:**
- `StartMigration` receives a `dryRun bool` parameter (added to existing signature)
- When `dryRun == true`, the runner resolves paths and estimates sizes but skips all copy operations
- The reporter uses a **null reporter** (no-op implementation of the reporter interface) — no log file is opened or written
- The reporter package exposes a `NewNullReporter()` for this purpose

**Done when:**
- UI matches W11 spacing and style conventions
- FolderPicker fits the design language
- Dry-run mode runs end-to-end with no files written and no log created

---

### Phase 5 — End-to-End Testing

**Goal:** A full real migration run completes successfully on W11.

**Test checklist:**
- [ ] All items resolve paths correctly on W11 (`%USERPROFILE%` expands to E: drive profile)
- [ ] Folder picker: selected items copy, unselected items don't
- [ ] Dry-run: no files written, estimates shown, no log created
- [ ] Skipped behavior: missing backup source surfaces correctly in summary
- [ ] Failed behavior: permission errors surface in summary with detail
- [ ] Log file and manifest JSON written to `logs/` and paths shown in summary
- [ ] Full Restore profile selects everything expected
- [ ] **Manual:** Launch LurkBait after migration and confirm images relink in-game (out of tool scope — manual verification only)

---

### Phase 6 — Manifest Contract & Build

**Goal:** Lock the manifest format for KtulueKit-Cleanup, write it as JSON, then ship the `.exe`.

**Manifest JSON format (the Cleanup contract):**
```json
{
  "version": "1.0",
  "runAt": "2026-03-11T14:30:00Z",
  "items": [
    {
      "app": "OBS Studio",
      "label": "scenes & profiles",
      "sourcePath": "D:\\KtulueBackup\\obs-studio\\basic",
      "targetPath": "C:\\Users\\...\\AppData\\Roaming\\obs-studio\\basic",
      "status": "copied",
      "bytesCopied": 1048576,
      "selectedPaths": []
    },
    {
      "app": "Personal Files",
      "label": "Documents",
      "sourcePath": "D:\\KtulueBackup\\personal\\Documents",
      "targetPath": "E:\\Users\\...\\Documents",
      "status": "copied",
      "bytesCopied": 524288,
      "selectedPaths": [
        "D:\\KtulueBackup\\personal\\Documents\\StreamScripts",
        "D:\\KtulueBackup\\personal\\Documents\\Notes.txt"
      ]
    }
  ]
}
```

`selectedPaths` is empty for `mirror`/`file` strategy items (Cleanup deletes at `sourcePath` level). For `selective` items it lists exactly which sub-paths were copied — Cleanup deletes only those, leaving unselected items in the backup untouched.

**Who writes the manifest file:**
The `reporter` package owns this. `reporter.WriteManifest(path string)` serializes the collected results to JSON after the run completes. Called from `app.go` in the `complete` event emission block, alongside `rep.LogPath()`.

**Manifest path:** `logs/manifest_YYYY-MM-DD_HH-MM-SS.json` — same timestamp as the `.log` file. Both paths shown in SummaryScreen.

**Build:**
- `wails build` produces `ktuluekit-migration.exe`
- `ktuluekit-migration.json` ships alongside the `.exe`
- Confirm `.exe` runs standalone (no Go/Wails/Node required on target machine)

**Done when:**
- `reporter.WriteManifest` writes valid JSON matching the contract above
- Manifest path shown in SummaryScreen alongside log path
- Manifest format documented for KtulueKit-Cleanup dev
- `.exe` builds and runs clean

---

## Config Schema Updates

- Add `"strategy"` field to Item: `"mirror"` (default) | `"file"` | `"selective"`. **Retire `"glob"`** — it was listed in the schema but never implemented in the runner. Remove from enum. No existing config entries use it.
- Add new categories to schema enum: `"Dev Tools"`, `"Creative Suite"`, `"Communication"`, `"Games"`, `"Personal Files"`

---

## New Category Order

Update `types.go` — this is the source of truth, not the spec:

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

---

## What This Tool Does NOT Do

- Delete anything (that's KtulueKit-Cleanup)
- Export from running apps (OBS, Streamer.bot export is a manual pre-step)
- Install software (that's KtulueKit-W11)
- Sync or watch for changes (one-shot restore only)
