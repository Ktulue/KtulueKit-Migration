# Impeccable UI Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply the Impeccable UI design system to KtulueKit-Migration — CSS custom property token system, bundled Nunito font, blue accent, semantic status colors, and the normalize→distill→polish→colorize→animate skill sequence on every screen and component.

**Architecture:** Token foundation first (App.svelte `:root` + `@font-face`), then per-screen skill sequence in order: SelectionScreen → ProgressScreen → SummaryScreen → PathBar+PreflightPanel → FolderPicker → CategoryAccordion+ItemRow → ProgressItem. No backend or Wails binding changes. No feature changes.

**Tech Stack:** Svelte 4, Vite 5, Wails v2. Frontend dev server: `wails dev` from project root. All styling lives in component `<style>` blocks; tokens declared on `:root` in `App.svelte`. Font files at `frontend/src/assets/fonts/`.

**Spec:** `docs/superpowers/specs/2026-03-15-impeccable-ui-design.md`

---

## File Map

| File | What changes |
|------|-------------|
| `frontend/src/App.svelte` | Add `:root` token block, `@font-face` declarations, `:global(body)` reset, box-sizing reset |
| `frontend/src/assets/fonts/nunito-400.woff2` | New — bundled font weight 400 |
| `frontend/src/assets/fonts/nunito-600.woff2` | New — bundled font weight 600 |
| `frontend/src/assets/fonts/nunito-700.woff2` | New — bundled font weight 700 |
| `frontend/src/screens/SelectionScreen.svelte` | Full skill sequence — replace all hardcoded values |
| `frontend/src/screens/ProgressScreen.svelte` | Full skill sequence — replace all hardcoded values |
| `frontend/src/screens/SummaryScreen.svelte` | Full skill sequence — replace all hardcoded values |
| `frontend/src/components/PathBar.svelte` | Full skill sequence — replace all hardcoded values |
| `frontend/src/components/PreflightPanel.svelte` | Full skill sequence — replace all hardcoded values |
| `frontend/src/components/FolderPicker.svelte` | Full skill sequence — replace all hardcoded values |
| `frontend/src/components/CategoryAccordion.svelte` | Full skill sequence — replace all hardcoded values, convert rem to token-based px |
| `frontend/src/components/ItemRow.svelte` | Full skill sequence — replace all hardcoded values |
| `frontend/src/components/ProgressItem.svelte` | Full skill sequence — replace all hardcoded values, convert rem to token-based px |

---

## Chunk 1: Token Foundation

### Task 1: Acquire Nunito font files

**Files:**
- Create: `frontend/src/assets/fonts/nunito-400.woff2`
- Create: `frontend/src/assets/fonts/nunito-600.woff2`
- Create: `frontend/src/assets/fonts/nunito-700.woff2`

- [ ] **Step 1: Create the fonts directory and install @fontsource/nunito temporarily**

```bash
mkdir -p frontend/src/assets/fonts
cd frontend && npm install --save-dev @fontsource/nunito
```

- [ ] **Step 2: Copy the three woff2 files into the assets directory**

```bash
cp frontend/node_modules/@fontsource/nunito/files/nunito-latin-400-normal.woff2 frontend/src/assets/fonts/nunito-400.woff2
cp frontend/node_modules/@fontsource/nunito/files/nunito-latin-600-normal.woff2 frontend/src/assets/fonts/nunito-600.woff2
cp frontend/node_modules/@fontsource/nunito/files/nunito-latin-700-normal.woff2 frontend/src/assets/fonts/nunito-700.woff2
```

- [ ] **Step 3: Uninstall the package — we only needed the files**

```bash
cd frontend && npm uninstall @fontsource/nunito
```

- [ ] **Step 4: Verify all three files exist and are non-zero**

```bash
ls -lh frontend/src/assets/fonts/
```

Expected: three `.woff2` files, each between 30–60 KB.

---

### Task 2: Add token system and font loading to App.svelte

**Files:**
- Modify: `frontend/src/App.svelte` (the `<style>` block, lines 127–149)

Current `<style>` block has a bare `:global(body)` with hardcoded values and no token system. Replace it entirely with the full token foundation.

- [ ] **Step 1: Replace the entire `<style>` block in App.svelte**

Replace:
```svelte
<style>
  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: #1a1a1a;
    color: #e0e0e0;
  }

  main {
    height: 100vh;
    overflow: hidden;
  }

  .error {
    padding: 2rem;
    text-align: center;
  }

  .error h2 {
    color: #e55;
  }
</style>
```

With:
```svelte
<style>
  @font-face {
    font-family: 'Nunito';
    src: url('./assets/fonts/nunito-400.woff2') format('woff2');
    font-weight: 400;
    font-style: normal;
    font-display: swap;
  }

  @font-face {
    font-family: 'Nunito';
    src: url('./assets/fonts/nunito-600.woff2') format('woff2');
    font-weight: 600;
    font-style: normal;
    font-display: swap;
  }

  @font-face {
    font-family: 'Nunito';
    src: url('./assets/fonts/nunito-700.woff2') format('woff2');
    font-weight: 700;
    font-style: normal;
    font-display: swap;
  }

  :root {
    /* Font */
    --font-primary: 'Nunito', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;

    /* Typography scale */
    --font-size-xs:   11px;
    --font-size-sm:   12px;
    --font-size-base: 15px;
    --font-size-lg:   16px;
    --font-size-xl:   18px;
    --font-size-2xl:  20px;

    /* Shape */
    --radius: 4px;

    /* Spacing (4px grid) */
    --spacing-xs:  4px;
    --spacing-sm:  6px;
    --spacing-md:  8px;
    --spacing-lg:  12px;
    --spacing-xl:  16px;
    --spacing-2xl: 20px;

    /* Colors — backgrounds */
    --color-bg-primary:   #1a1a1a;
    --color-bg-secondary: #111;
    --color-bg-hover:     #2a2a2a;

    /* Colors — borders */
    --color-border:       #333;
    --color-border-input: #555;

    /* Colors — text */
    --color-text-primary:   #e0e0e0;
    --color-text-secondary: #888;
    --color-text-tertiary:  #aaa;

    /* Colors — accent (blue) */
    --color-accent:          #0e7fd4;
    --color-accent-hover:    #1290e8;
    --color-accent-disabled: #444;

    /* Colors — danger */
    --color-danger:        #ff6b6b;
    --color-danger-action: #c0392b;

    /* Colors — success (Migration-specific) */
    --color-success:       #2ea043;
    --color-success-hover: #3ab854;

    /* Colors — warning (Migration-specific) */
    --color-warning:       #e6a817;
    --color-warning-hover: #f0b929;
  }

  :global(*, *::before, *::after) {
    box-sizing: border-box;
  }

  :global(body) {
    margin: 0;
    font-family: var(--font-primary);
    background: var(--color-bg-primary);
    color: var(--color-text-primary);
  }

  main {
    height: 100vh;
    overflow: hidden;
  }

  .error {
    padding: 2rem;
    text-align: center;
  }

  .error h2 {
    color: var(--color-danger);
  }
</style>
```

- [ ] **Step 2: Start the dev server and verify the token foundation loads without errors**

```bash
wails dev
```

Expected: app opens, no console errors, font loads (text visibly uses Nunito — more rounded than the system sans-serif).

- [ ] **Step 3: Commit**

```bash
git add frontend/src/App.svelte frontend/src/assets/fonts/
git commit -m "feat: add Impeccable UI token system and bundle Nunito font"
```

---

## Chunk 2: SelectionScreen

### Task 3: Apply Impeccable UI to SelectionScreen.svelte

**Files:**
- Modify: `frontend/src/screens/SelectionScreen.svelte`

Current file has hardcoded values throughout: `#111`, `#333`, `#2ea043` (green accent), `#d4a017` (ad-hoc warning), rem-free but using literal px. Dry-run toggle uses `accent-color: #d4a017`.

Design decisions from spec:
- Header: `--color-bg-secondary`, `--font-size-2xl` for h1, `--spacing-2xl` horizontal padding
- `.accent` span in h1: `--color-accent` (blue, not green)
- Dry-run checkbox: `accent-color: var(--color-warning)`
- Profile select: `--color-bg-hover`, `--color-border-input`, `--radius`
- Pre-flight Check button (footer, secondary): `--color-warning` text/border, `transparent` background
- Start Migration button (footer, primary): `--color-accent` background, `--color-accent-hover` hover, `scale(0.98)` active, `--color-accent-disabled` when disabled

- [ ] **Step 1: Invoke the `normalize` skill on SelectionScreen.svelte**

Use the `normalize` skill. Target: `frontend/src/screens/SelectionScreen.svelte`. Goal: replace every hardcoded color, spacing, and font-size value in the `<style>` block with the appropriate CSS custom property token from the `:root` definition in App.svelte.

Key mappings for normalize pass:
- `#111` → `var(--color-bg-secondary)`
- `#333` → `var(--color-border)`
- `#2ea043` (accent green) → `var(--color-accent)` (blue)
- `#3ab553` (green hover) → `var(--color-accent-hover)`
- `#d4a017` (warning amber) → `var(--color-warning)`
- `rgba(212, 160, 23, 0.1)` → `rgba(230, 168, 23, 0.1)` (warning tint, keep inline — close enough)
- `#999` / `#888` → `var(--color-text-secondary)`
- `#e0e0e0` → `var(--color-text-primary)`
- `#2a2a2a` → `var(--color-bg-hover)`
- `#444` → `var(--color-border-input)`
- `12px` / `13px` → `var(--font-size-sm)`
- `14px` → `var(--font-size-base)` (close enough; base is 15px — use base)
- `20px` font-size on h1 → `var(--font-size-2xl)`
- `12px 20px` padding → `var(--spacing-lg) var(--spacing-2xl)`
- `16px` gap → `var(--spacing-xl)`
- `8px` gap → `var(--spacing-md)`
- `6px` gap → `var(--spacing-sm)`
- `6px` border-radius → `var(--radius)` (spec uses 4px — normalize to `--radius`)
- `#333` disabled → `var(--color-accent-disabled)`
- `#666` disabled text → `var(--color-text-secondary)`

- [ ] **Step 2: Invoke the `distill` skill on SelectionScreen.svelte**

Use the `distill` skill. Remove unnecessary styles, simplify redundant rules, clean up anything that doesn't serve the user.

- [ ] **Step 3: Invoke the `polish` skill on SelectionScreen.svelte**

Use the `polish` skill. Fix alignment, spacing consistency, typographic hierarchy. Verify footer layout, header layout, button sizing are consistent.

- [ ] **Step 4: Invoke the `colorize` skill on SelectionScreen.svelte**

Use the `colorize` skill. Ensure accent blue is applied correctly to the `.accent` span in h1, the Start Migration CTA, and checkboxes. Pre-flight Check button should be warning amber (text + border). Dry-run toggle checkbox uses `--color-warning` accent-color.

- [ ] **Step 5: Invoke the `animate` skill on SelectionScreen.svelte**

Use the `animate` skill. Ensure:
- Button hover transitions: `100ms ease`
- Start button active state: `transform: scale(0.98)`
- Select/input focus transitions: `100ms ease`

- [ ] **Step 6: Verify in dev server**

```bash
wails dev
```

Visually confirm:
- Header is dark (`#111`), title uses Nunito at 20px, "Migration" word is blue not green
- Pre-flight Check button is amber-outlined
- Start Migration button is blue
- Dry-run checkbox tick is amber

- [ ] **Step 7: Commit**

```bash
git add frontend/src/screens/SelectionScreen.svelte
git commit -m "feat: apply Impeccable UI to SelectionScreen"
```

---

## Chunk 3: ProgressScreen

### Task 4: Apply Impeccable UI to ProgressScreen.svelte

**Files:**
- Modify: `frontend/src/screens/ProgressScreen.svelte`

Current file uses `#2ea043` (green) for progress bar, `#2a2000`/`#3a3000` ad-hoc dark amber for dry-run banner background.

Design decisions from spec:
- Progress bar: `--color-accent` (blue)
- Dry-run banner: `rgba` tint of `--color-warning` background, `--color-warning` text, `--color-border` bottom border
- Feed items: entrance fade 150ms, 50ms stagger

- [ ] **Step 1: Invoke the `normalize` skill on ProgressScreen.svelte**

Use the `normalize` skill. Key mappings:
- `#111` → `var(--color-bg-secondary)`
- `#333` → `var(--color-border)`
- `#2ea043` (progress bar) → `var(--color-accent)`
- `#e0e0e0` → `var(--color-text-primary)`
- `#999` → `var(--color-text-secondary)`
- `#2a2000` (dry-run bg) → `rgba(230, 168, 23, 0.08)` (warning tint — keep inline)
- `#d4a017` (dry-run text) → `var(--color-warning)`
- `#3a3000` (dry-run border) → `var(--color-border)`
- `12px 20px` padding → `var(--spacing-lg) var(--spacing-2xl)`
- `18px` h2 → `var(--font-size-xl)`
- `12px` progress-text → `var(--font-size-sm)`
- `6px` banner padding-vertical → `var(--spacing-sm)`

- [ ] **Step 2: Invoke the `distill` skill on ProgressScreen.svelte**

- [ ] **Step 3: Invoke the `polish` skill on ProgressScreen.svelte**

- [ ] **Step 4: Invoke the `colorize` skill on ProgressScreen.svelte**

Verify: progress bar is blue, dry-run banner is amber-tinted with amber text.

- [ ] **Step 5: Invoke the `animate` skill on ProgressScreen.svelte**

Ensure: `transition: width 0.3s ease` on `.progress-bar`. Add entrance fade for `.feed` items if not already present — this may require a Svelte `in:fade` directive or a CSS animation on `.progress-item` (coordinate with ProgressItem task).

- [ ] **Step 6: Verify in dev server**

Visually confirm: progress bar is blue, dry-run banner is amber.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/screens/ProgressScreen.svelte
git commit -m "feat: apply Impeccable UI to ProgressScreen"
```

---

## Chunk 4: SummaryScreen

### Task 5: Apply Impeccable UI to SummaryScreen.svelte

**Files:**
- Modify: `frontend/src/screens/SummaryScreen.svelte`

Current file: `#2ea043` (green) on manifest toggle button and Close button, `#d4a017` on status-skipped rows, ad-hoc monospace font family strings.

Design decisions from spec:
- Close button (primary): `--color-accent` background
- Run Again button (secondary): `--color-bg-hover` bg, `--color-text-secondary` text, `--color-border` border
- Manifest toggle button: `--color-accent` text + border (replaces green)
- `tr.status-copied td:nth-child(5)`: `var(--color-success)` (stays green — this is status, not accent)
- `tr.status-skipped td:nth-child(5)`: `var(--color-warning)` (replaces hardcoded `#d4a017`)
- `tr.status-failed td:nth-child(5)`: `var(--color-danger)`
- Section badges: accent/secondary/danger `rgba` tint pattern (distill pass may add these)
- Monospace font family: keep as-is (path display — functional, not UI font)
- Copy buttons: `--color-text-tertiary`, hover to `--color-text-primary`

- [ ] **Step 1: Invoke the `normalize` skill on SummaryScreen.svelte**

Key mappings:
- `#111` → `var(--color-bg-secondary)`
- `#333` → `var(--color-border)`
- `#2ea043` (Close btn, manifest-btn, status-copied) → `var(--color-accent)` for buttons; `var(--color-success)` for status cell color
- `#3ab553` (hover) → `var(--color-accent-hover)` for Close btn hover; `var(--color-success-hover)` for status-copied hover if any
- `#d4a017` (status-skipped) → `var(--color-warning)`
- `#e55` (status-failed) → `var(--color-danger)`
- `#e0e0e0` → `var(--color-text-primary)`
- `#ccc` → `var(--color-text-primary)`
- `#999` / `#888` → `var(--color-text-secondary)`
- `#aaa` / `#bbb` → `var(--color-text-tertiary)`
- `#666` → `var(--color-text-secondary)` or `var(--color-text-tertiary)` depending on context
- `#555` (separator) → `var(--color-border-input)`
- `#222` (item border, table row border) → `var(--color-bg-hover)`
- `18px` → `var(--font-size-xl)`
- `14px` → `var(--font-size-base)`
- `13px` / `12px` → `var(--font-size-sm)`
- `12px 20px` padding → `var(--spacing-lg) var(--spacing-2xl)`
- `20px` content padding → `var(--spacing-2xl)`
- `8px` gaps/padding → `var(--spacing-md)`
- `6px` border-radius → `var(--radius)`

- [ ] **Step 2: Invoke the `distill` skill on SummaryScreen.svelte**

The manifest table section is complex. Distill will tighten column widths, remove visual noise, simplify the toggle button area.

- [ ] **Step 3: Invoke the `polish` skill on SummaryScreen.svelte**

- [ ] **Step 4: Invoke the `colorize` skill on SummaryScreen.svelte**

Verify: Close is blue, Run Again is neutral secondary, manifest toggle is blue-outlined, status cells use correct semantic colors (green copied, amber skipped, red failed). Section headers (Copied/Skipped/Failed) should get `rgba` badge treatment if the distill pass hasn't already applied it.

- [ ] **Step 5: Invoke the `animate` skill on SummaryScreen.svelte**

Entrance fades on sections (150ms, 50ms stagger). Copy button hover transition 100ms.

- [ ] **Step 6: Verify in dev server**

Run a dry-run migration or review the SummaryScreen directly. Confirm Close is blue, Run Again is secondary, status colors are correct.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/screens/SummaryScreen.svelte
git commit -m "feat: apply Impeccable UI to SummaryScreen"
```

---

## Chunk 5: PathBar + PreflightPanel

### Task 6: Apply Impeccable UI to PathBar.svelte

**Files:**
- Modify: `frontend/src/components/PathBar.svelte`

Current file uses `#141414` as background (not in token system — maps to `--color-bg-secondary`), `#2ea043`/`#d4a017` for icon status colors.

Design decisions from spec:
- `.path-bar` background: `--color-bg-secondary` (replaces `#141414`)
- `.icon-ok`: `--color-success`
- `.icon-error`: `--color-warning`
- Browse buttons: `--color-text-secondary`, hover to `--color-accent` (per spec)

- [ ] **Step 1: Invoke the `normalize` skill on PathBar.svelte**

Key mappings:
- `#141414` → `var(--color-bg-secondary)`
- `#333` → `var(--color-border)`
- `#444` → `var(--color-border-input)`
- `#888` → `var(--color-text-secondary)`
- `#e0e0e0` → `var(--color-text-primary)`
- `#999` → `var(--color-text-secondary)`
- `#2a2a2a` (input bg) → `var(--color-bg-hover)`
- `#555` (input focus border) → `var(--color-border-input)`
- `#2ea043` (icon-ok) → `var(--color-success)`
- `#d4a017` (icon-error) → `var(--color-warning)`
- `#555` (icon-unchecked) → `var(--color-border-input)`
- `4px 8px` input padding → `var(--spacing-xs) var(--spacing-md)`
- `3px 10px` browse btn padding → `var(--spacing-xs) var(--spacing-lg)`
- `7px 20px` path-bar padding → `var(--spacing-sm) var(--spacing-2xl)`
- `4px` gap → `var(--spacing-xs)`
- `8px` gap → `var(--spacing-md)`
- `12px` font-size → `var(--font-size-sm)`
- `4px` border-radius → `var(--radius)`

- [ ] **Step 2: Invoke `distill` → `polish` → `colorize` → `animate` on PathBar.svelte**

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/PathBar.svelte
git commit -m "feat: apply Impeccable UI to PathBar"
```

### Task 7: Apply Impeccable UI to PreflightPanel.svelte

**Files:**
- Modify: `frontend/src/components/PreflightPanel.svelte`

Current file uses `#181818` (between bg-primary and bg-secondary), `#e55` for errors, `#d4a017` for warnings.

Design decisions from spec:
- `.preflight-panel` background: `--color-bg-secondary` (replaces `#181818`)
- `.err`: `--color-danger` (replaces `#e55`)
- `.warn`: `--color-warning` (replaces `#d4a017`)
- `.run-anyway-label`: `--color-warning`, `accent-color: var(--color-warning)`
- `.item-row.item-missing`: `--color-warning`
- `.item-path`: keep monospace font-family (functional); use `--font-size-xs`

- [ ] **Step 1: Invoke the `normalize` skill on PreflightPanel.svelte**

Key mappings:
- `#181818` → `var(--color-bg-secondary)`
- `#333` → `var(--color-border)`
- `#e55` → `var(--color-danger)`
- `#d4a017` → `var(--color-warning)`
- `#999` → `var(--color-text-secondary)`
- `#555` → `var(--color-border-input)`
- `#666` → `var(--color-text-secondary)`
- `6px 20px` padding → `var(--spacing-sm) var(--spacing-2xl)`
- `12px` font-size → `var(--font-size-sm)`
- `12px` gap → `var(--spacing-lg)`
- `5px` gap → `var(--spacing-sm)`
- `4px 0` item padding → `var(--spacing-xs) 0`
- `6px` gap → `var(--spacing-sm)`
- `11px` item-path font-size → `var(--font-size-xs)`
- `16px` list padding → `var(--spacing-xl)`

- [ ] **Step 2: Invoke `distill` → `polish` → `colorize` → `animate` on PreflightPanel.svelte**

- [ ] **Step 3: Verify PathBar + PreflightPanel together in dev server**

Both components should visually read as a unified "setup zone" with consistent `--color-bg-secondary` background and `--color-border` bottom borders.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/PreflightPanel.svelte
git commit -m "feat: apply Impeccable UI to PreflightPanel"
```

---

## Chunk 6: FolderPicker

### Task 8: Apply Impeccable UI to FolderPicker.svelte

**Files:**
- Modify: `frontend/src/components/FolderPicker.svelte`

Current file uses `#1e1e1e` (modal card — close to bg-primary but slightly lighter), `#181818` (toolbar/footer), `#2ea043` (confirm button and checkboxes).

Design decisions from spec:
- Overlay backdrop: `rgba(0, 0, 0, 0.6)` — intentional one-off, keep as literal value
- Modal card: `--color-bg-primary` background, `--color-border` border, `--radius`
- Toolbar/footer bg: `--color-bg-secondary`
- Checkboxes: `accent-color: var(--color-accent)` (blue)
- Confirm button: `--color-accent` background
- Cancel button: `--color-bg-hover` background, `--color-text-secondary` text

- [ ] **Step 1: Invoke the `normalize` skill on FolderPicker.svelte**

Key mappings:
- `#1e1e1e` (modal) → `var(--color-bg-primary)`
- `#181818` (toolbar/footer) → `var(--color-bg-secondary)`
- `#2a2a2a` (entry hover) → `var(--color-bg-hover)`
- `#333` → `var(--color-border)`
- `#444` → `var(--color-border-input)`
- `#999` → `var(--color-text-secondary)`
- `#666` → `var(--color-text-secondary)`
- `#ddd` → `var(--color-text-primary)`
- `#e0e0e0` → `var(--color-text-primary)`
- `#2ea043` (confirm btn, checkboxes) → `var(--color-accent)`
- `#3ab553` (confirm hover) → `var(--color-accent-hover)`
- `rgba(46, 160, 67, 0.1)` can be removed — confirm button won't use this pattern
- `#333` disabled → `var(--color-accent-disabled)`
- `6px` border-radius (modal) → `var(--radius)` (spec says 4px — normalize to `--radius`)
- `4px` border-radius (buttons) → `var(--radius)`
- `16px 20px` modal-header padding → `var(--spacing-xl) var(--spacing-2xl)`
- `8px 20px` toolbar padding → `var(--spacing-md) var(--spacing-2xl)`
- `12px 20px` footer padding → `var(--spacing-lg) var(--spacing-2xl)`
- `6px 20px` entry padding → `var(--spacing-sm) var(--spacing-2xl)`
- `8px 20px` cancel/confirm padding → `var(--spacing-md) var(--spacing-2xl)`
- `3px 10px` select-all padding → `var(--spacing-xs) var(--spacing-lg)`
- `12px` gap toolbar → `var(--spacing-lg)`
- `8px` gap entries → `var(--spacing-md)`
- `4px 0` modal-body padding → `var(--spacing-xs) 0`
- `15px` h3 → `var(--font-size-base)`
- `13px` entry → `var(--font-size-sm)`
- `12px` toolbar text → `var(--font-size-sm)`
- `11px` path/size → `var(--font-size-xs)`
- `14px` confirm font-size → `var(--font-size-base)`

- [ ] **Step 2: Invoke `distill` → `polish` → `colorize` → `animate` on FolderPicker.svelte**

- [ ] **Step 3: Verify in dev server**

Open a selective item to trigger the picker. Confirm: blue checkboxes, blue Confirm button, neutral Cancel, dark modal card.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/FolderPicker.svelte
git commit -m "feat: apply Impeccable UI to FolderPicker"
```

---

## Chunk 7: CategoryAccordion + ItemRow

### Task 9: Apply Impeccable UI to CategoryAccordion.svelte

**Files:**
- Modify: `frontend/src/components/CategoryAccordion.svelte`

Current file uses **rem units** (not px) — these must be converted to token-based px. Uses `#2a2a2a` accordion header (correct — maps to `--color-bg-hover`), `#333` hover, `#888`/`#999`/`#555` text.

- [ ] **Step 1: Invoke the `normalize` skill on CategoryAccordion.svelte**

Key mappings (including rem → token px conversions):
- `0.25rem` margin → `var(--spacing-xs)`
- `0.6rem 0.75rem` header padding → `var(--spacing-sm) var(--spacing-lg)`
- `0.5rem` gap → `var(--spacing-md)`
- `0.7rem` arrow font-size → `var(--font-size-sm)`
- `0.95rem` name font-size → `var(--font-size-base)`
- `0.85rem` count font-size → `var(--font-size-sm)`
- `0.15rem 0.5rem` select-all padding → `2px var(--spacing-sm)` (2px is below --spacing-xs — use literal 2px)
- `0.75rem` select-all font-size → `var(--font-size-sm)`
- `1.5rem` items padding-left → `var(--spacing-2xl)`
- `#2a2a2a` header bg → `var(--color-bg-hover)` (correct — accordion headers are NOT bg-secondary)
- `#333` header hover → `var(--color-border)` is too subtle; use `var(--color-border-input)` for hover
- `#888` → `var(--color-text-secondary)`
- `#999` → `var(--color-text-secondary)`
- `#555` → `var(--color-border-input)`
- `#e0e0e0` → `var(--color-text-primary)`
- `4px` border-radius → `var(--radius)`
- `0.15s` transition → keep as `150ms` (or `0.15s` — consistent with rest of app)

- [ ] **Step 2: Invoke `distill` → `polish` → `colorize` → `animate` on CategoryAccordion.svelte**

The animate pass should ensure the arrow rotation transition (`transform 0.15s`) is consistent with the app's `100ms ease` standard — either keep at 150ms or normalize to 100ms.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/CategoryAccordion.svelte
git commit -m "feat: apply Impeccable UI to CategoryAccordion"
```

### Task 10: Apply Impeccable UI to ItemRow.svelte

**Files:**
- Modify: `frontend/src/components/ItemRow.svelte`

Current file: `#2ea043` (green) for checkbox accent-color and picker button. No rem units — px throughout.

Design decisions from spec:
- Checkboxes: `accent-color: var(--color-accent)` (blue)
- Picker button (`.picker-btn`): `--color-accent` text + border, `rgba(14, 127, 212, 0.1)` hover tint

- [ ] **Step 1: Invoke the `normalize` skill on ItemRow.svelte**

Key mappings:
- `#2ea043` (checkbox accent) → `var(--color-accent)`
- `#2ea043` (picker-btn color/border) → `var(--color-accent)`
- `rgba(46,160,67,0.1)` (picker hover) → `rgba(14, 127, 212, 0.1)` (accent tint — keep inline)
- `#444` (tooltip bg) → `var(--color-bg-hover)`
- `#999` (tooltip trigger text) → `var(--color-text-secondary)`
- `#333` (tooltip bg) → `var(--color-border)` — actually tooltip bg should be `var(--color-bg-hover)`, border implicit
- `#ddd` (tooltip text) → `var(--color-text-primary)`
- `13px` label → `var(--font-size-sm)`
- `11px` → `var(--font-size-xs)`
- `12px` tooltip → `var(--font-size-sm)`
- `6px 0` item padding → `var(--spacing-sm) 0`
- `8px` gaps → `var(--spacing-md)`
- `1px 8px` picker-btn padding → `1px var(--spacing-md)`
- `8px 12px` tooltip padding → `var(--spacing-md) var(--spacing-lg)`
- `3px` border-radius picker → `var(--radius)`
- `4px` border-radius tooltip → `var(--radius)`

- [ ] **Step 2: Invoke `distill` → `polish` → `colorize` → `animate` on ItemRow.svelte**

- [ ] **Step 3: Verify CategoryAccordion + ItemRow in dev server**

SelectionScreen should now show: dark-grey accordion headers (bg-hover), blue checkboxes, blue "Edit selection" picker buttons.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ItemRow.svelte
git commit -m "feat: apply Impeccable UI to ItemRow"
```

---

## Chunk 8: ProgressItem

### Task 11: Apply Impeccable UI to ProgressItem.svelte

**Files:**
- Modify: `frontend/src/components/ProgressItem.svelte`

Current file uses **rem units** throughout and `#2ea043` (green) for detail toggle button.

Design decisions from spec:
- `.detail-toggle`: `--color-accent` (blue, replaces green)
- `.detail` block: `--color-bg-secondary` background
- Entrance fade: `150ms` (coordinate with ProgressScreen)

- [ ] **Step 1: Invoke the `normalize` skill on ProgressItem.svelte**

Key mappings (rem → token px):
- `0.5rem 0` padding → `var(--spacing-md) 0`
- `1rem` main-line font-size → `var(--font-size-base)`
- `0.5rem` gap → `var(--spacing-md)`
- `0.8rem` counter/elapsed font-size → `var(--font-size-sm)`
- `0.75rem` detail-toggle font-size → `var(--font-size-sm)`
- `0.2rem 0` detail-toggle padding → `2px 0`
- `0.2rem` margin-top → `2px`
- `0.5rem` detail padding → `var(--spacing-md)`
- `0.3rem` margin-top → `var(--spacing-xs)`
- `0.75rem` detail font-size → `var(--font-size-sm)`
- `#222` (border-bottom) → `var(--color-bg-hover)`
- `#2ea043` (detail-toggle) → `var(--color-accent)`
- `#111` (detail bg) → `var(--color-bg-secondary)`
- `#999` (detail text) → `var(--color-text-secondary)`
- `#666` (counter/elapsed) → `var(--color-text-secondary)`
- `#ddd` (name) → `var(--color-text-primary)`
- `4px` border-radius detail → `var(--radius)`
- `50px` min-width counter → keep as literal (layout constraint, not a token)

- [ ] **Step 2: Invoke `distill` → `polish` → `colorize` → `animate` on ProgressItem.svelte**

The animate pass should add entrance fade (`animation: fadeIn 150ms ease`). Define a `@keyframes fadeIn` in the component or apply Svelte's `in:fade={{ duration: 150 }}` directive in the template.

- [ ] **Step 3: Verify in dev server**

Run a migration (or dry-run) and watch the progress feed. Items should fade in, detail toggle should be blue.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ProgressItem.svelte
git commit -m "feat: apply Impeccable UI to ProgressItem"
```

---

## Chunk 9: Final Verification + Cleanup

### Task 12: Full visual pass and hardcoded value audit

- [ ] **Step 1: Grep for any remaining hardcoded values across all component files**

```bash
grep -rn "#[0-9a-fA-F]\{3,6\}\|font-size: [0-9]\|padding: [0-9]\|gap: [0-9]\|margin: [0-9]" \
  frontend/src/screens/ frontend/src/components/ \
  --include="*.svelte"
```

Expected: zero matches for color hex values. A small number of layout px values (e.g., `min-width: 50px`, `width: 18px`, `width: 250px`) are acceptable if they are structural constraints rather than design tokens.

- [ ] **Step 2: Start `wails dev` and walk through all three screens**

Walk-through checklist:
- SelectionScreen loads: Nunito font visible, header dark, "Migration" in blue, accordion headers dark-grey, checkboxes blue
- Load a profile: items check in blue
- Pre-flight Check button is amber-outlined
- Run preflight: result panel shows in bg-secondary, warnings in amber, errors in red
- Dry-run mode: start button shifts to disabled grey
- ProgressScreen: progress bar is blue, dry-run banner is amber if active, items fade in
- SummaryScreen: Close is blue, Run Again is secondary/neutral, status cells use correct colors, manifest table renders cleanly

- [ ] **Step 3: Commit any final touch-up fixes found during walk-through**

```bash
git add -p  # stage only relevant changes
git commit -m "fix: final Impeccable UI touch-ups from visual walk-through"
```

### Task 13: Update how-to-use.md if needed

Per project feedback memory: `docs/how-to-use.md` must be updated whenever user-facing features change.

- [ ] **Step 1: Review docs/how-to-use.md for any references that need updating**

The Impeccable UI pass is visual only — no feature changes. Check if the doc references any visual elements (screenshots, color descriptions) that are now outdated. If nothing references specific colors or visual appearance, no update is needed.

- [ ] **Step 2: Commit if changes were made**

```bash
git add docs/how-to-use.md
git commit -m "docs: update how-to-use for Impeccable UI visual changes"
```

### Task 14: Open PR

- [ ] **Step 1: Push the branch**

```bash
git push -u origin feat/impeccable-ui
```

- [ ] **Step 2: Open PR**

```bash
gh pr create \
  --title "feat: apply Impeccable UI design system to KtulueKit-Migration" \
  --body "$(cat <<'EOF'
## Summary

- Token system: full CSS custom property `:root` block in App.svelte (colors, spacing, typography, shape)
- Nunito font bundled as `.woff2` (400/600/700) — no CDN dependency, works offline in Wails
- Blue accent `#0e7fd4` replaces green `#2ea043` throughout
- Two new semantic tokens: `--color-success` (mounted/copied) and `--color-warning` (not-mounted/dry-run/preflight warn)
- Every screen and component received full normalize→distill→polish→colorize→animate skill sequence
- Zero hardcoded hex/px values in component files — all reference CSS custom properties

## Test plan

- [ ] `wails dev` — app opens, Nunito font loads, no console errors
- [ ] SelectionScreen: header dark, "Migration" blue, checkboxes blue, preflight button amber, start button blue
- [ ] ProgressScreen: progress bar blue, dry-run banner amber
- [ ] SummaryScreen: Close blue, Run Again secondary/neutral, status cells green/amber/red
- [ ] FolderPicker: blue confirm button, blue checkboxes
- [ ] Grep confirms no hardcoded hex values in component files

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 3: Stop — do not merge. Report PR URL to user and wait for approval.**
