<script>
  import CategoryAccordion from '../components/CategoryAccordion.svelte'
  import PathBar from '../components/PathBar.svelte'
  import PreflightPanel from '../components/PreflightPanel.svelte'
  import { PreflightCheck, ScanDrive, BrowseForFolder, DetectDestination } from '../../wailsjs/go/main/App'

  export let configView
  export let onStart
  export let onOpenPicker = (item) => {}
  export let onProfileChange = () => {}
  export let initialSourceRoot = ''
  export let initialDestRoot = ''

  let selected = new Set()
  let profileValue = ''
  let dryRun = false

  let sourceRoot = initialSourceRoot || (configView && configView.backupRoot) || ''
  let destRoot = initialDestRoot

  let preflightResult = null
  let preflightDone = false
  let runAnyway = false

  let discoveryMap = {}
  let destMap = {}
  let scanning = false
  let scanDone = false

  async function handleAssist(item) {
    try {
      const chosen = await BrowseForFolder(sourceRoot)
      if (chosen) {
        discoveryMap = {
          ...discoveryMap,
          [item.id]: { ...discoveryMap[item.id], found: true, sourcePath: chosen }
        }
        selected.add(item.id)
        selected = new Set(selected)
      }
    } catch (err) {
      console.error('Assist browse failed:', err)
    }
  }

  async function handleScan(e) {
    const { sourcePath } = e.detail
    if (!sourcePath) return
    scanning = true
    scanDone = false
    discoveryMap = {}
    try {
      const result = await ScanDrive(sourcePath)
      const map = {}
      for (const item of result.items) {
        map[item.id] = item
      }
      discoveryMap = map
      scanDone = true

      // Auto-select found items
      for (const item of result.items) {
        if (item.found) {
          selected.add(item.id)
        }
      }
      selected = new Set(selected)
    } catch (err) {
      console.error('Scan failed:', err)
    } finally {
      scanning = false
    }
  }

  async function handleDetect(appName) {
    const appSourcePaths = {}
    for (const [id, item] of Object.entries(discoveryMap)) {
      if (item.found && item.sourcePath && id.startsWith(appName + ':')) {
        appSourcePaths[id] = item.sourcePath
      }
    }
    if (Object.keys(appSourcePaths).length === 0) return

    try {
      const results = await DetectDestination(appName, appSourcePaths)
      for (const r of results) {
        destMap[r.itemId] = r
      }
      destMap = { ...destMap }
    } catch (err) {
      console.error('Detection failed:', err)
    }
  }

  async function handleDestOverride(itemId) {
    try {
      const current = destMap[itemId]?.destPath || ''
      const chosen = await BrowseForFolder(current)
      if (chosen) {
        destMap[itemId] = {
          ...destMap[itemId],
          itemId,
          destPath: chosen,
          method: 'manual',
          confirmed: true,
          candidates: []
        }
        destMap = { ...destMap }
      }
    } catch (err) {
      console.error('Dest browse failed:', err)
    }
  }

  function buildDestPathMap() {
    const map = {}
    for (const [id, result] of Object.entries(destMap)) {
      if (result.destPath) {
        map[id] = result.destPath
      }
    }
    return map
  }

  function buildSourcePathMap() {
    const map = {}
    for (const [id, item] of Object.entries(discoveryMap)) {
      if (item.found && item.sourcePath) {
        map[id] = item.sourcePath
      }
    }
    return map
  }

  function resetPreflight() {
    preflightResult = null
    preflightDone = false
    runAnyway = false
  }

  // Reset preflight whenever paths or selections change
  $: {
    sourceRoot; destRoot; selected;
    resetPreflight()
  }

  function loadProfile(e) {
    const profileName = e.target.value
    if (!profileName) return

    const profile = configView.profiles.find(p => p.name === profileName)
    if (profile) {
      selected = new Set(profile.ids)
    }
    profileValue = profileName
    onProfileChange()
  }

  function handleToggle() {
    selected = new Set(selected)
  }

  // Wrap onOpenPicker to add the item to selected when picker is confirmed
  function handleOpenPickerWrapped(item) {
    onOpenPicker(item, (itemId) => {
      selected.add(itemId)
      handleToggle()
    })
  }

  async function handlePreflight() {
    try {
      preflightResult = await PreflightCheck([...selected], sourceRoot, destRoot, buildSourcePathMap(), buildDestPathMap())
      preflightDone = true
    } catch (e) {
      preflightDone = false
      console.error('Preflight failed:', e)
    }
  }

  function handleRunAnyway(e) {
    runAnyway = e.detail
  }

  function handleStart() {
    onStart([...selected], {}, dryRun, sourceRoot, destRoot, buildSourcePathMap(), buildDestPathMap())
  }

  $: selectedCount = selected.size

  $: startEnabled = selectedCount > 0 &&
    preflightDone &&
    preflightResult &&
    preflightResult.sourceRootOK &&
    preflightResult.destRootOK &&
    (!preflightResult.hasItemWarnings || runAnyway)
</script>

<div class="selection-screen">
  <header>
    <div class="header-left">
      <h1>KtulueKit <span class="accent">Migration</span></h1>
    </div>
    <div class="header-right">
      <label class="dry-run-label">
        <input type="checkbox" bind:checked={dryRun} />
        Dry run
      </label>
      <select on:change={loadProfile} bind:value={profileValue}>
        <option value="">Load profile...</option>
        {#each configView.profiles as profile}
          <option value={profile.name}>{profile.name}</option>
        {/each}
      </select>
    </div>
  </header>

  <PathBar
    {sourceRoot}
    {destRoot}
    {scanning}
    on:change={(e) => { sourceRoot = e.detail.sourceRoot; destRoot = e.detail.destRoot }}
    on:scan={handleScan}
  />

  <PreflightPanel result={preflightResult} on:runAnyway={handleRunAnyway} />

  <div class="content">
    {#each configView.categories as category}
      <CategoryAccordion {category} {selected} {discoveryMap} {destMap} onToggle={handleToggle} onOpenPicker={handleOpenPickerWrapped} onAssist={handleAssist} onDetect={handleDetect} onDestOverride={handleDestOverride} />
    {/each}
  </div>

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
</div>

<style>
  .selection-screen {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--spacing-lg) var(--spacing-2xl);
    background: var(--color-bg-secondary);
    border-bottom: 1px solid var(--color-border);
  }

  h1 {
    margin: 0;
    font-size: var(--font-size-2xl);
    font-weight: 600;
    color: var(--color-text-primary);
  }

  .accent {
    color: var(--color-accent);
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: var(--spacing-xl);
  }

  .dry-run-label {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    font-size: var(--font-size-sm);
    color: var(--color-text-secondary);
    cursor: pointer;
  }

  .dry-run-label input { accent-color: var(--color-warning); }

  select {
    background: var(--color-bg-hover);
    color: var(--color-text-primary);
    border: 1px solid var(--color-border-input);
    border-radius: var(--radius);
    padding: var(--spacing-sm) 10px;
    font-size: var(--font-size-sm);
    transition: border-color 100ms ease;
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: var(--spacing-lg) var(--spacing-2xl);
  }

  footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--spacing-lg) var(--spacing-2xl);
    background: var(--color-bg-secondary);
    border-top: 1px solid var(--color-border);
  }

  .count {
    font-size: var(--font-size-sm);
    color: var(--color-text-secondary);
  }

  .footer-actions {
    display: flex;
    gap: var(--spacing-md);
    align-items: center;
  }

  .preflight-btn {
    background: transparent;
    color: var(--color-warning);
    border: 1px solid var(--color-warning);
    border-radius: var(--radius);
    padding: var(--spacing-md) var(--spacing-xl);
    font-size: var(--font-size-base);
    font-weight: 600;
    cursor: pointer;
    transition: background 100ms ease;
  }

  .preflight-btn:hover:not(:disabled) {
    background: rgba(230, 168, 23, 0.1);
  }

  .preflight-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .start-btn {
    background: var(--color-accent);
    color: #fff;
    border: none;
    border-radius: var(--radius);
    padding: var(--spacing-md) var(--spacing-2xl);
    font-size: var(--font-size-base);
    font-weight: 600;
    cursor: pointer;
    transition: background 100ms ease;
  }

  .start-btn:hover:not(:disabled) {
    background: var(--color-accent-hover);
  }

  .start-btn:active:not(:disabled) {
    transform: scale(0.98);
  }

  .start-btn:disabled {
    background: var(--color-accent-disabled);
    color: var(--color-text-secondary);
    cursor: not-allowed;
  }
</style>
