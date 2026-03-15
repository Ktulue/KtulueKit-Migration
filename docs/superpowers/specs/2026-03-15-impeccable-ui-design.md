# KtulueKit-Migration — Impeccable UI Design Spec

**Date:** 2026-03-15
**Branch:** `feat/impeccable-ui`
**Status:** Approved

---

## Overview

Apply the Impeccable UI design system to KtulueKit-Migration. The tool is feature-complete (all phases 1–4b done, Go tests passing, `.exe` builds). This pass replaces all hardcoded style values with a CSS custom property token system matching KtulueKit-W11, bundles the Nunito font for offline Wails use, swaps the green accent for blue `#0e7fd4`, and extends the palette with two Migration-specific semantic tokens. Every screen and component receives the full normalize → distill → polish → colorize → animate skill sequence.

---

## Goals

- Consistent visual identity across the KtulueKit suite (W11, Migration, Cleanup)
- Zero hardcoded values in component files — all styles reference CSS custom properties
- Nunito font bundled as `.woff2` (no CDN dependency — Wails apps run offline)
- Blue accent (`#0e7fd4`) replacing green (`#2ea043`) as the primary action color
- Semantic status colors for Migration-specific states (mounted, not-mounted, preflight)
- Per-screen skill sequence: normalize → distill → polish → colorize → animate

---

## Non-Goals

- No feature changes — UI pass only
- No layout restructuring beyond what the skill sequence naturally produces
- No changes to Go backend or Wails bindings
- No E2E testing (phase 5) in this branch — that follows separately

---

## Architecture

### Token System

All design tokens declared on `:root` in `App.svelte`. Component files reference only `var(--token-name)` — no hardcoded hex, px, or font-size values anywhere except the `:root` block itself.

#### Color Tokens (W11-inherited)

| Token | Value | Purpose |
|---|---|---|
| `--color-bg-primary` | `#1a1a1a` | Main screen background |
| `--color-bg-secondary` | `#111` | Headers, footers, setup zone |
| `--color-bg-hover` | `#2a2a2a` | Row hover, button hover background |
| `--color-border` | `#333` | Structural dividers, accordion borders |
| `--color-border-input` | `#555` | Input/select field borders |
| `--color-text-primary` | `#e0e0e0` | Body text, item names |
| `--color-text-secondary` | `#888` | Counts, labels, inactive states |
| `--color-text-tertiary` | `#aaa` | Timestamps, copy buttons, metadata |
| `--color-accent` | `#0e7fd4` | CTAs, checkboxes, progress bar, links |
| `--color-accent-hover` | `#1290e8` | Accent hover state |
| `--color-accent-disabled` | `#444` | Disabled button background |
| `--color-danger` | `#ff6b6b` | Failed status, hard preflight failures |
| `--color-danger-action` | `#c0392b` | Destructive action buttons |

#### Color Tokens (Migration-specific additions)

| Token | Value | Purpose |
|---|---|---|
| `--color-success` | `#2ea043` | Backup root mounted ✓, preflight pass, copied status |
| `--color-success-hover` | `#3ab854` | Success hover state |
| `--color-warning` | `#e6a817` | Backup root not mounted ⚠, preflight warn, dry-run banner |
| `--color-warning-hover` | `#f0b929` | Warning hover state |

#### Spacing Scale (4px grid)

| Token | Value |
|---|---|
| `--spacing-xs` | `4px` |
| `--spacing-sm` | `6px` |
| `--spacing-md` | `8px` |
| `--spacing-lg` | `12px` |
| `--spacing-xl` | `16px` |
| `--spacing-2xl` | `20px` |

#### Typography Scale

| Token | Value | Usage |
|---|---|---|
| `--font-size-xs` | `11px` | Timestamps, log detail, raw output |
| `--font-size-sm` | `12px` | Metadata, counts, copy buttons |
| `--font-size-base` | `15px` | Body text, item names, labels |
| `--font-size-lg` | `16px` | Screen titles, section headers |
| `--font-size-xl` | `18px` | App header |

#### Shape

| Token | Value | Usage |
|---|---|---|
| `--radius` | `4px` | All interactive elements |

#### Font

| Token | Value |
|---|---|
| `--font-primary` | `"Nunito", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif` |

### Font Loading

Nunito bundled as `.woff2` files in `frontend/src/assets/fonts/`:
- `nunito-400.woff2` (regular)
- `nunito-600.woff2` (semi-bold)
- `nunito-700.woff2` (bold)

Loaded via `@font-face` in App.svelte `<style>` block. No CDN. No `@import url(fonts.googleapis.com/...)`.

### style.css

`frontend/src/style.css` stripped to a minimal stub:
```css
*, *::before, *::after {
  box-sizing: border-box;
}
```
All real styles live in component `<style>` blocks or the App.svelte `:root`.

---

## Implementation Plan

### Commit 1 — Token foundation

**Files:** `App.svelte`, `frontend/src/style.css`, `frontend/src/assets/fonts/` (new)

- Add `:root` token block to App.svelte (all tokens above)
- Add `@font-face` declarations for Nunito 400/600/700
- Add `:global(body)` reset
- Strip style.css to box-sizing stub
- Download and commit the three `.woff2` font files
- **No component changes in this commit** — purely additive infrastructure

### Commits 2–8 — Per-screen/component skill sequence

Each gets the full **normalize → distill → polish → colorize → animate** pass:

| Commit | Target |
|---|---|
| 2 | `SelectionScreen.svelte` |
| 3 | `ProgressScreen.svelte` |
| 4 | `SummaryScreen.svelte` |
| 5 | `PathBar.svelte` + `PreflightPanel.svelte` |
| 6 | `FolderPicker.svelte` |
| 7 | `CategoryAccordion.svelte` + `ItemRow.svelte` |
| 8 | `ProgressItem.svelte` |

---

## Per-Screen Design Decisions

### SelectionScreen

- Header: `--color-bg-secondary` background, `--font-size-xl` title, `--spacing-2xl` horizontal padding
- PathBar + PreflightPanel sit below header in a shared `--color-bg-secondary` container — visually grouped as the "setup zone"
- Preflight result rows: `--color-success` for mounted ✓, `--color-warning` for not-mounted ⚠, `--color-danger` for hard failures
- Profile select + dry-run toggle: all controls use `--color-border-input` borders, `--radius`, token-referenced padding
- Start Migration CTA: `--color-accent` background, `--color-accent-hover` on hover, `transform: scale(0.98)` on `:active`
- Dry-run mode: button shifts to `--color-accent-disabled`; dry-run indicator uses `--color-warning` tint

### ProgressScreen

- Dry-run banner: `--color-warning` background tint (amber, not the current ad-hoc value)
- Progress feed item entrance: fade-in at `150ms`, `50ms` stagger per item
- Status icons: copied → `--color-success`, skipped → `--color-text-secondary`, failed → `--color-danger`
- Progress bar: `--color-accent`, `transition: width 0.3s ease`

### SummaryScreen

- Section badges (Copied / Skipped / Failed): accent/secondary/danger using `rgba` tint pattern from W11
  - Accent badge: `rgba(14, 127, 212, 0.15)` fill, `rgba(14, 127, 212, 0.35)` border
  - Danger badge: `rgba(255, 107, 107, 0.15)` fill, `rgba(255, 107, 107, 0.35)` border
  - Secondary badge: `--color-bg-hover` fill, `--color-border` border
- Manifest table: distill pass tightens column widths, removes visual noise
- Log/manifest copy buttons: `--color-text-tertiary`, hover to `--color-text-primary`, `100ms ease`
- Run Again button: `--color-accent`

### PathBar + PreflightPanel

- Both live in the same `--color-bg-secondary` container, `--color-border` bottom border separating from item list
- Inputs: `--color-border-input` border, `--color-bg-hover` on focus, `--radius`
- Refresh button: `--color-text-secondary`, hover to `--color-accent`

### FolderPicker

- Backdrop: `rgba(0, 0, 0, 0.6)` (intentional one-off — overlay not tokenized)
- Modal card: `--color-bg-primary` background, `--color-border` border, `--radius`
- Confirm: `--color-accent`; Cancel: `--color-bg-hover` background, `--color-text-secondary` text

### CategoryAccordion + ItemRow

- Accordion headers: `--color-bg-secondary`, `--color-border` bottom border, `--spacing-lg` vertical padding
- Checkboxes: `--color-accent` (single mode — no uninstall dual-accent complexity)
- Row hover: `--color-bg-hover`, `100ms ease` transition
- Selective strategy trigger button: `--color-accent` text, hover underline

### ProgressItem

- Compact row: `--spacing-sm` vertical padding, `--spacing-xl` left padding
- Status dot/icon: success/secondary/danger tokens per status
- Entrance fade: `150ms` (coordinated with ProgressScreen stagger)

---

## Animation Reference (matching W11 patterns)

| Use | Duration | Easing |
|---|---|---|
| Button/tab/border hover | `100ms` | `ease` |
| Item entrance fade | `150ms` | `ease` |
| Summary category stagger | `50ms` per item | `ease` |
| Progress bar width | `0.3s` | `ease` |
| Button active scale | `transform: scale(0.98)` | — |

No bounce, no spring, no glow, no box-shadow spread.

---

## W11 Follow-Up (out of scope for this branch)

W11 currently loads Nunito via Google Fonts CDN (`@import url(...)`). A separate `maint/bundle-nunito-font` branch on KtulueKit-W11 should mirror this spec's `@font-face` approach. Migration sets the correct pattern; W11 catches up.
