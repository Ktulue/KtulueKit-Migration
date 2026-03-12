# KtulueKit-Migration — TODO

## ✅ Done
- Project scaffold: Go + Wails v2 + Svelte frontend
- 3-screen UI: Selection → Progress → Summary
- Backend packages: config loader, runner, copier, mapper, reporter
- JSON config + JSON Schema for validation
- Category accordion with select-all, profile presets, tooltips
- Real-time progress feed with auto-scroll
- Summary screen: copied/skipped/failed sections, manifest table, log path
- Go dependencies installed (`go mod tidy`)
- Frontend dependencies installed (`npm install`)
- Wails CLI confirmed (v2.11.0)
- LurkBait Twitch Fishing added to config (Games category)
- Design spec approved: `docs/superpowers/specs/2026-03-11-migration-buildout-design.md`
- Implementation plan written: `docs/superpowers/plans/2026-03-11-migration-buildout.md`

---

## Phase 1 — Dev Run & Wiring Verification
- [x] `wails dev` smoke test — window opens, config loads, screens navigate
- [x] Wails JS bindings generated (`frontend/wailsjs/`)
- [x] End-to-end run completes (items skip if no backup present — expected)
- [x] Log file written to `logs/`

---

## Phase 2 — Config Expansion
- [x] Schema updated: strategy enum (mirror/file/selective), category enum added
- [x] `categoryOrder` in `types.go` updated with all new categories
- [x] Dev Tools added: Git, SSH, VS Code, Windows Terminal
- [x] Creative Suite added: GIMP, Krita, Audacity, Blender, Aseprite, Kdenlive, DaVinci Resolve
- [x] Utilities added: ShareX, PowerToys, Claude Desktop
- [x] Communication added: Discord, Spotify
- [x] Personal Files added: Desktop (mirror), Documents/Videos/Pictures/Music (selective)
- [x] OBS Studio + Streamer.bot notes updated (close before backup)
- [x] Full Restore profile updated with all new IDs
- [x] Dev Tools profile added

---

## Phase 3 — Folder Picker Feature
- [x] `copier.CopyPath` added + tests pass
- [x] `reporter.NewNull()` + `reporter.WriteManifest()` + timestamp field added + tests pass
- [x] `runner`: selective strategy, dry-run mode, `SelectedPaths` on result + tests pass
- [x] `types.go`: `Strategy` on `ItemView`, `FolderEntry`, `SelectedPaths` on `ManifestEntry`
- [x] `app.go`: `ListFolder`, `GetSourcePath`, updated `StartMigration` signature
- [x] `FolderPicker.svelte` component created
- [x] `ItemRow.svelte` updated for selective strategy trigger
- [x] `CategoryAccordion.svelte` "Select all" skips selective items
- [x] `App.svelte` manages `selectivePaths`, `pickerItem`, `pickerSourcePath`
- [x] `SelectionScreen.svelte` dry-run toggle added
- [x] Wails bindings regenerated after `ListFolder` + `GetSourcePath` added

---

## Phase 4 — UI Polish
- [x] Spacing aligned with KtulueKit-W11 (12px/20px, px values)
- [x] Dry-run banner in ProgressScreen
- [x] Manifest path shown in SummaryScreen footer
- [x] FolderPicker visually consistent with app theme

---

## Phase 5 — End-to-End Testing
- [ ] `%USERPROFILE%` resolves to E: drive correctly
- [ ] Full Restore run completes on real backup
- [ ] Selective item (Documents): only selected paths copy
- [ ] Dry-run: no files, no log, no manifest written
- [ ] Skipped + failed behaviors surface correctly
- [ ] Manifest JSON valid with `selectedPaths` populated
- [ ] **Manual:** LurkBait images relink in-game after restore

---

## Phase 6 — Build & Manifest Contract
- [x] `wails build` produces `ktuluekit-migration.exe` (11 MB, `build/bin/`)
- [ ] Exe runs standalone (no Go/Wails/Node on target) — verify on fresh W11
- [x] Manifest format documented for KtulueKit-Cleanup dev (`docs/manifest-contract.md`)
- [x] `.gitignore` confirmed to exclude `build/bin/`

---

## Handoff to KtulueKit-Cleanup
- [x] Manifest contract locked: `logs/manifest_<timestamp>.json`
- [x] Schema: version, runAt, items[]{app, label, sourcePath, targetPath, status, bytesCopied, selectedPaths}
- [x] Document contract in `docs/manifest-contract.md`
