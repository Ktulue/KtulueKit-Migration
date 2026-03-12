# KtulueKit-Migration — Manifest Contract

**For:** KtulueKit-Cleanup development
**Written by:** KtulueKit-Migration v1.0
**Status:** Locked

---

## Overview

After every successful (non-dry-run) migration, KtulueKit-Migration writes a manifest JSON file to:

```
logs/manifest_YYYY-MM-DD_HH-MM-SS.json
```

The timestamp matches the `.log` file from the same run (e.g., `migration_2026-03-11_14-30-00.log` + `manifest_2026-03-11_14-30-00.json`).

KtulueKit-Cleanup reads this manifest to know exactly what was copied and where, so it can delete source backup files after the user confirms a successful restore.

---

## File Format

```json
{
  "version": "1.0",
  "runAt": "2026-03-11T14:30:00Z",
  "items": [
    {
      "app": "OBS Studio",
      "label": "scenes & profiles",
      "sourcePath": "D:\\KtulueBackup\\obs-studio\\basic",
      "targetPath": "C:\\Users\\Klute\\AppData\\Roaming\\obs-studio\\basic",
      "status": "copied",
      "bytesCopied": 1048576,
      "selectedPaths": []
    },
    {
      "app": "Personal Files",
      "label": "Documents",
      "sourcePath": "D:\\KtulueBackup\\personal\\Documents",
      "targetPath": "E:\\Users\\Klute's Stream Rig\\Documents",
      "status": "copied",
      "bytesCopied": 524288,
      "selectedPaths": [
        "D:\\KtulueBackup\\personal\\Documents\\StreamScripts",
        "D:\\KtulueBackup\\personal\\Documents\\Notes.txt"
      ]
    },
    {
      "app": "GIMP",
      "label": "brushes & settings",
      "sourcePath": "D:\\KtulueBackup\\gimp",
      "targetPath": "C:\\Users\\Klute\\AppData\\Roaming\\GIMP",
      "status": "skipped",
      "bytesCopied": 0,
      "selectedPaths": []
    }
  ]
}
```

---

## Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Always `"1.0"` for this contract |
| `runAt` | string | ISO 8601 UTC timestamp of when the migration run started |
| `items[].app` | string | App name from config (e.g., `"OBS Studio"`) |
| `items[].label` | string | Item label from config (e.g., `"scenes & profiles"`) |
| `items[].sourcePath` | string | Absolute path to the backup source (fully resolved, no env vars) |
| `items[].targetPath` | string | Absolute path to the restore target (fully resolved) |
| `items[].status` | string | `"copied"` \| `"skipped"` \| `"failed"` |
| `items[].bytesCopied` | number | Bytes written during copy; `0` for skipped/failed |
| `items[].selectedPaths` | array | For `selective` strategy: the sub-paths that were copied. Empty array `[]` for mirror/file items |

---

## Cleanup Rules (for KtulueKit-Cleanup)

**Mirror / file strategy items** (`selectedPaths: []`):
Cleanup deletes the entire `sourcePath` directory (or file). All contents were copied.

**Selective strategy items** (`selectedPaths` non-empty):
Cleanup deletes **only the paths listed in `selectedPaths`**, not the entire `sourcePath`. The user did not select everything — unselected items remain in the backup and must not be touched.

**Skipped items** (`status: "skipped"`):
Source was not found or had no paths selected. Nothing was copied. Cleanup should not touch these.

**Failed items** (`status: "failed"`):
Copy was attempted but errored. Do not delete source — the user may need to retry.

---

## Notes for Cleanup Dev

- All paths in the manifest are **fully resolved absolute paths** — no `%APPDATA%` or other env vars
- `sourcePath` always points into `D:\KtulueBackup\` (or wherever `backup_root` was configured)
- The manifest only contains items that were **selected and attempted** — items the user did not select are absent
- Multiple manifests may exist in `logs/` — Cleanup should let the user pick which run to clean up, or default to the most recent
- The log file (`.log`) alongside the manifest contains human-readable detail for each item including error messages

---

## Where It's Written

`reporter.WriteManifest(path string) error` in `internal/reporter/reporter.go`
Called from `app.go` `StartMigration` goroutine after `r.Run()` completes, skipped when `dryRun == true`.
