<script>
  import CategoryAccordion from '../components/CategoryAccordion.svelte'
  import PathBar from '../components/PathBar.svelte'
  import PreflightPanel from '../components/PreflightPanel.svelte'
  import { PreflightCheck } from '../../wailsjs/go/main/App'

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
      preflightResult = await PreflightCheck([...selected], sourceRoot, destRoot)
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
    onStart([...selected], {}, dryRun, sourceRoot, destRoot)
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
    on:change={(e) => { sourceRoot = e.detail.sourceRoot; destRoot = e.detail.destRoot }}
  />

  <PreflightPanel result={preflightResult} on:runAnyway={handleRunAnyway} />

  <div class="content">
    {#each configView.categories as category}
      <CategoryAccordion {category} {selected} onToggle={handleToggle} onOpenPicker={handleOpenPickerWrapped} />
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
    padding: 12px 20px;
    background: #111;
    border-bottom: 1px solid #333;
  }

  h1 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: #e0e0e0;
  }

  .accent {
    color: #2ea043;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .dry-run-label {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #999;
    cursor: pointer;
  }

  .dry-run-label input { accent-color: #d4a017; }

  select {
    background: #2a2a2a;
    color: #e0e0e0;
    border: 1px solid #444;
    border-radius: 4px;
    padding: 6px 10px;
    font-size: 13px;
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: 12px 20px;
  }

  footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 20px;
    background: #111;
    border-top: 1px solid #333;
  }

  .count {
    font-size: 13px;
    color: #999;
  }

  .footer-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .preflight-btn {
    background: transparent;
    color: #d4a017;
    border: 1px solid #d4a017;
    border-radius: 6px;
    padding: 8px 16px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s;
  }

  .preflight-btn:hover:not(:disabled) {
    background: rgba(212, 160, 23, 0.1);
  }

  .preflight-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .start-btn {
    background: #2ea043;
    color: #fff;
    border: none;
    border-radius: 6px;
    padding: 8px 20px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s;
  }

  .start-btn:hover:not(:disabled) {
    background: #3ab553;
  }

  .start-btn:disabled {
    background: #333;
    color: #666;
    cursor: not-allowed;
  }
</style>
