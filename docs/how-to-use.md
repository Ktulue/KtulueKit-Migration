# KtulueKit Migration — How to Use

> **Last updated:** 2026-03-28
> **App version covers:** v1.0 feature set + path override / pre-flight check + source discovery + destination detection

---

## What this tool does

KtulueKit Migration copies your app configs, profiles, and personal files from a backup drive to the correct locations on a fresh Windows install. It reads a config file (`ktuluekit-migration.json`) that maps each item from its backup source to its Windows destination. You pick what you want to restore, run a pre-flight check, then start the migration.

---

## Before you start

1. **Mount your backup drive.** The app needs the drive to be accessible. If the drive letter changed between your old PC and your new one, you'll set the correct path in the Source field (see below).
2. **Close any apps you're about to restore.** OBS, Streamer.bot, Brave, etc. should be closed so nothing has their config files locked.
3. **Launch `ktuluekit-migration.exe`** from `build/bin/`.

---

## The Selection Screen

This is the main screen. Everything happens here before you start.

### Path Bar (top, below the title)

Two fields:

| Field | What it does |
|-------|-------------|
| **Source** | Where your backup files live. Pre-filled from `backup_root` in the config. Change this if your backup drive has a different letter or path on this machine. |
| **Destination** | Where files will be copied *to*. Leave blank to use the paths in the config as-is. Set this to override the destination drive (e.g. if your Windows drive is `D:\` instead of `C:\`). |

**Browse** buttons open a native folder picker for each field.

**Scan** button (next to Source Browse) scans the source drive for all config items. See [Source Discovery](#source-discovery-scanning-a-cloned-drive) below.

**Drive normalisation:** If you type just `D:` and click away, it automatically becomes `D:\`.

**Destination override rules:**
- `D:\` (just a drive root, 3 characters) → swaps the drive letter on all target paths. `C:\Users\Foo\...` becomes `D:\Users\Foo\...`
- `D:\Restored\` (a longer path) → replaces the `C:\` prefix entirely. `C:\Users\Foo\...` becomes `D:\Restored\Users\Foo\...`
- Blank → no override; all destination paths come straight from the config

### Selecting items

Items are grouped by category (Streaming, Dev Tools, etc.). Click any item's checkbox to include it in the migration. Click a category header to collapse or expand it.

**Profiles (top-right dropdown):** Load a saved selection preset. The config ships with:
- **Full Restore** — everything
- **Streaming Only** — OBS, Streamer.bot, Stream Deck, media assets
- **Browser & Identity** — Brave, PowerShell profile
- **Dev Tools** — Git, SSH, VS Code, Windows Terminal

### Selective items (folder picker)

Items marked `selective` in the config (Documents, Videos, Pictures, etc.) have a **pick files** button instead of a plain checkbox. Clicking it opens a folder picker inside the backup so you can choose specific subfolders or files rather than copying everything.

### Dry Run toggle

Enable **Dry run** (top-right, next to the profile dropdown) to simulate the migration without writing any files. The progress and summary screens will show exactly what *would* have happened. Use this to verify your setup before a real run.

---

## Source Discovery (scanning a cloned drive)

If your source is a full clone of a Windows drive (e.g., your old W10 C:\ cloned to E:\), the app can automatically find where each item's data lives.

### How to use it

1. Set the **Source** field to the root of your cloned drive (e.g., `E:\`)
2. Click the **Scan** button (green, next to Browse)
3. The app scans `E:\Users\*` for user profiles, resolves each item's expected Windows path against the cloned drive, and reports results

### What you'll see

After scanning, each item in the selection list shows a status badge:

| Badge | Meaning |
|-------|---------|
| **found** (green) | The item's data was located on the source drive. The item is automatically checked. |
| **not found** (grey) | The item was not found. The item is dimmed and unchecked. |

Items marked **not found** have a **Locate** button. Click it to manually browse the source drive and point the app to the correct folder.

### How it works under the hood

The scanner looks at each config item's *target* path (e.g., `%APPDATA%/obs-studio/basic`) and resolves the environment variables against the cloned drive's user profile structure. For example, `%APPDATA%` becomes `E:\Users\YourName\AppData\Roaming`. If multiple user profiles exist on the cloned drive, the one with the most matches wins.

Discovered paths are held in memory for the session — the config file is never modified.

---

## Destination Detection (smart target resolution)

After scanning your source drive, the app can figure out where each item should go on your current machine — so you don't have to know the exact `%APPDATA%` path for every app.

### How to use it

1. **Scan your source drive** first (see above) — items need discovered source paths before detection works
2. In the item list, you'll see **Detect** buttons next to each app name that has discovered sources
3. Click **Detect** on an app (e.g., "OBS Studio") — the tool resolves where those files should land on this machine
4. Results appear inline on each item as destination badges and a clickable path

### What you'll see

After detection, each item shows a destination badge:

| Badge | Meaning |
|-------|---------|
| **confirmed** (green) | The destination path was resolved and the folder exists on this machine. |
| **unconfirmed** (yellow) | The destination path was resolved but the folder doesn't exist yet. It will be created during migration. |
| **dest not found** (red) | The tool couldn't figure out where this app's data goes. You need to set it manually. |

Next to the badge, the resolved destination path is shown as a clickable link. Click it to **override** the destination with a folder picker if the auto-resolved path isn't right.

For items marked **dest not found**, a **Set destination** button lets you browse to the correct location.

### How it works under the hood

**Tier 1 — Path pattern mapping (automatic):** The tool looks at where the source files were found (e.g., `E:\Users\Josh\AppData\Roaming\obs-studio\basic`) and remaps the drive letter and username to your current machine (e.g., `C:\Users\YourName\AppData\Roaming\obs-studio\basic`). This handles ~80% of apps — anything that stores data in standard Windows user profile locations.

**Tier 2 — Detection hints (config-driven):** For apps that don't follow standard paths (games, portable installs), the config can include a `detection` block with registry keys to check or directories to search. The tool tries these as a fallback when Tier 1 doesn't match.

Detection results are held in memory for the session. Per-item detected destinations take precedence over the global Destination override in the path bar.

---

## Pre-flight Check

Before you can start a migration, you must run a pre-flight check. This validates:

1. The **source root** exists and is a directory
2. The **destination root** exists (or its parent does, so it can be created)
3. Every selected item's **source path** can be found on disk

Click **Pre-flight Check** in the footer. The result panel appears below the path bar.

| Result | What it means |
|--------|--------------|
| **N/M ready** (all green) | Everything found — Start Migration is now enabled |
| **N/M ready** with warnings | Some source paths not found. A **Run anyway** checkbox appears. Check it to proceed without the missing items. |
| **Source root not found** | The source path doesn't exist. Fix the Source field and re-run. |
| **Destination root not found** | The destination drive/path doesn't exist. Fix the Destination field or leave it blank. |

> **Note:** Changing the source path, destination path, or any item selection automatically clears the pre-flight result and disables Start Migration. Just run pre-flight again after making changes.

---

## Running the Migration

Once pre-flight passes (and you've checked **Run anyway** if needed), click **Start Migration**.

The **Progress Screen** shows each item as it copies, with live byte counts and elapsed time.

If the migration is interrupted or errors out, affected items are logged as `failed` — your source files are never deleted by this tool.

---

## Summary Screen

After the migration completes:

- **Copied / Skipped / Failed** counts at a glance
- Per-item results — expand to see what was skipped or failed and why
- **View Log** — opens the `.log` file from this run in `logs/`
- **Run Again** — returns to the selection screen with your source/destination paths pre-filled from this run, so you can tweak and re-run without re-entering everything
- **Close** — exits the app

---

## Log Files

Every run (non-dry-run) writes two files to `logs/`:

| File | Contents |
|------|----------|
| `migration_YYYY-MM-DD_HH-MM-SS.log` | Human-readable log — each item, its paths, status, and any error messages |
| `manifest_YYYY-MM-DD_HH-MM-SS.json` | Machine-readable record of what was copied — used by KtulueKit-Cleanup |

Dry runs do not write log files.

---

## Config File (`ktuluekit-migration.json`)

The config lives next to the `.exe`. It defines:

- `backup_root` — default source path (overridable at runtime via the Source field)
- `apps` — list of apps, each with one or more items mapping a backup source to a Windows target
- `profiles` — saved selection presets

Targets can use Windows environment variables: `%APPDATA%`, `%LOCALAPPDATA%`, `%USERPROFILE%`. These are expanded at runtime using the current user's environment.

**You should not need to edit this file just because your drive letter changed** — use the Source/Destination fields in the app instead.

---

## Troubleshooting

**Source root not found after pre-flight**
Your backup drive letter differs from what's in `backup_root`. Set the correct path in the **Source** field and re-run pre-flight.

**Some items show "not found" in pre-flight**
Those items weren't in your backup (or their folder name differs). Use **Run anyway** to skip them and restore everything else.

**SSH keys: permission warning after restore**
Windows doesn't preserve Unix-style file permissions. After restoring SSH keys, set the correct permissions manually:
`icacls "%USERPROFILE%\.ssh\id_rsa" /inheritance:r /grant:r "%USERNAME%:R"`

**Brave / browser extensions need re-enabling**
This is normal — extensions may require manual re-enabling or re-installation after a profile copy.

**OBS or Streamer.bot config not loading correctly**
Make sure those apps were closed before the migration ran. If they were open, restore again (the source files are unchanged).

---

## KtulueKit Trilogy

| Tool | Purpose |
|------|---------|
| **KtulueKit-Migration** *(this tool)* | Copy files from backup to their Windows locations |
| **KtulueKit-Cleanup** *(planned)* | Delete source backup files after a confirmed successful restore |
| **W11 Setup Tool** *(planned)* | First-run Windows 11 configuration automation |

KtulueKit-Cleanup reads the manifest files written by this tool to know exactly what was copied and where.
