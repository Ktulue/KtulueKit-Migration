<script>
  import CategoryAccordion from '../components/CategoryAccordion.svelte'

  export let configView
  export let onStart
  export let onOpenPicker = (item) => {}

  let selected = new Set()
  let profileValue = ''
  let dryRun = false

  function loadProfile(e) {
    const profileName = e.target.value
    if (!profileName) return

    const profile = configView.profiles.find(p => p.name === profileName)
    if (profile) {
      selected = new Set(profile.ids)
    }
    profileValue = profileName
  }

  function handleToggle() {
    selected = new Set(selected)
  }

  function handleStart() {
    onStart([...selected], {}, dryRun)
  }

  $: selectedCount = selected.size
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

  <div class="content">
    {#each configView.categories as category}
      <CategoryAccordion {category} {selected} onToggle={handleToggle} {onOpenPicker} />
    {/each}
  </div>

  <footer>
    <span class="count">{selectedCount} item{selectedCount !== 1 ? 's' : ''} selected</span>
    <button
      class="start-btn"
      disabled={selectedCount === 0}
      on:click={handleStart}
    >
      Start Migration
    </button>
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
