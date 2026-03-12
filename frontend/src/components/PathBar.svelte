<script>
  import { BrowseForFolder } from '../../wailsjs/go/main/App'

  export let sourceRoot = ''
  export let destRoot = ''

  // icon states: 'blank' | 'unchecked' | 'ok' | 'error'
  let sourceIcon = 'unchecked'
  let destIcon = 'blank'

  // normalise D: → D:\
  function normaliseDrive(path) {
    if (/^[A-Za-z]:$/.test(path)) return path + '\\'
    return path
  }

  import { createEventDispatcher } from 'svelte'
  const dispatch = createEventDispatcher()

  function emitChange() {
    dispatch('change', { sourceRoot, destRoot })
  }

  function handleSourceBlur() {
    sourceRoot = normaliseDrive(sourceRoot)
    sourceIcon = 'unchecked'
    emitChange()
  }

  function handleDestBlur() {
    destRoot = normaliseDrive(destRoot)
    destIcon = destRoot === '' ? 'blank' : 'unchecked'
    emitChange()
  }

  async function browseSource() {
    try {
      const chosen = await BrowseForFolder(sourceRoot)
      if (chosen) {
        sourceRoot = chosen
        sourceIcon = 'unchecked'
        emitChange()
      }
    } catch (e) {
      console.error('Browse failed:', e)
    }
  }

  async function browseDest() {
    try {
      const chosen = await BrowseForFolder(destRoot || '')
      if (chosen) {
        destRoot = chosen
        destIcon = 'unchecked'
        emitChange()
      }
    } catch (e) {
      console.error('Browse failed:', e)
    }
  }

</script>

<div class="path-bar">
  <div class="path-row">
    <span class="path-label">Source:</span>
    <input
      class="path-input"
      type="text"
      bind:value={sourceRoot}
      on:blur={handleSourceBlur}
      placeholder="Backup root (e.g. D:\Backup\W10)"
      spellcheck="false"
    />
    <button class="browse-btn" on:click={browseSource}>Browse</button>
    <span class="icon icon-{sourceIcon}" aria-label={sourceIcon}>
      {#if sourceIcon === 'ok'}✓{:else if sourceIcon === 'error'}⚠{:else if sourceIcon === 'unchecked'}—{/if}
    </span>
  </div>

  <div class="path-row">
    <span class="path-label">Destination:</span>
    <input
      class="path-input"
      type="text"
      bind:value={destRoot}
      on:blur={handleDestBlur}
      placeholder="Override destination root (optional, e.g. D:\)"
      spellcheck="false"
    />
    <button class="browse-btn" on:click={browseDest}>Browse</button>
    <span class="icon icon-{destIcon}" aria-label={destIcon}>
      {#if destIcon === 'ok'}✓{:else if destIcon === 'error'}⚠{:else if destIcon === 'unchecked'}—{/if}
    </span>
  </div>
</div>

<style>
  .path-bar {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 7px 20px;
    background: #141414;
    border-bottom: 1px solid #333;
  }

  .path-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .path-label {
    width: 80px;
    font-size: 12px;
    color: #888;
    flex-shrink: 0;
  }

  .path-input {
    flex: 1;
    background: #2a2a2a;
    color: #e0e0e0;
    border: 1px solid #444;
    border-radius: 4px;
    padding: 4px 8px;
    font-size: 12px;
    font-family: 'Cascadia Code', 'Consolas', monospace;
    min-width: 0;
  }

  .path-input:focus {
    outline: none;
    border-color: #555;
  }

  .browse-btn {
    background: transparent;
    color: #999;
    border: 1px solid #444;
    border-radius: 4px;
    padding: 3px 10px;
    font-size: 12px;
    cursor: pointer;
    flex-shrink: 0;
  }

  .browse-btn:hover { color: #e0e0e0; border-color: #666; }

  .icon {
    width: 18px;
    text-align: center;
    font-size: 13px;
    flex-shrink: 0;
  }

  .icon-ok    { color: #2ea043; }
  .icon-error { color: #d4a017; }
  .icon-unchecked { color: #555; }
  .icon-blank { color: transparent; }
</style>
