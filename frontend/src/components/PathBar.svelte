<script>
  import { createEventDispatcher } from 'svelte'
  import { BrowseForFolder } from '../../wailsjs/go/main/App'

  export let sourceRoot = ''
  export let destRoot = ''
  export let scanning = false

  const dispatch = createEventDispatcher()

  // icon states: 'blank' | 'unchecked' | 'ok' | 'error'
  let sourceIcon = 'unchecked'
  let destIcon = 'blank'

  // normalise D: → D:\
  function normaliseDrive(path) {
    if (/^[A-Za-z]:$/.test(path)) return path + '\\'
    return path
  }

  function emitChange() {
    dispatch('change', { sourceRoot, destRoot })
  }

  function handleScan() {
    sourceRoot = normaliseDrive(sourceRoot)
    emitChange()
    dispatch('scan', { sourcePath: sourceRoot })
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
    <button class="scan-btn" on:click={handleScan} disabled={!sourceRoot || scanning}>
      {scanning ? 'Scanning...' : 'Scan'}
    </button>
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
    gap: var(--spacing-xs);
    padding: var(--spacing-sm) var(--spacing-2xl);
    background: var(--color-bg-secondary);
    border-bottom: 1px solid var(--color-border);
  }

  .path-row {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
  }

  .path-label {
    width: 80px;
    font-size: var(--font-size-sm);
    color: var(--color-text-secondary);
    flex-shrink: 0;
  }

  .path-input {
    flex: 1;
    background: var(--color-bg-hover);
    color: var(--color-text-primary);
    border: 1px solid var(--color-border-input);
    border-radius: var(--radius);
    padding: var(--spacing-xs) var(--spacing-md);
    font-size: var(--font-size-sm);
    font-family: 'Cascadia Code', 'Consolas', monospace;
    min-width: 0;
    transition: border-color 100ms ease;
  }

  .path-input:focus {
    outline: none;
    border-color: var(--color-border-input);
  }

  .browse-btn {
    background: transparent;
    color: var(--color-text-secondary);
    border: 1px solid var(--color-border-input);
    border-radius: var(--radius);
    padding: var(--spacing-xs) var(--spacing-lg);
    font-size: var(--font-size-sm);
    cursor: pointer;
    flex-shrink: 0;
    transition: color 100ms ease, border-color 100ms ease;
  }

  .browse-btn:hover { color: var(--color-accent); border-color: var(--color-accent); }

  .scan-btn {
    background: transparent;
    color: var(--color-success);
    border: 1px solid var(--color-success);
    border-radius: var(--radius);
    padding: var(--spacing-xs) var(--spacing-lg);
    font-size: var(--font-size-sm);
    cursor: pointer;
    flex-shrink: 0;
    transition: color 100ms ease, border-color 100ms ease, background 100ms ease;
  }
  .scan-btn:hover:not(:disabled) {
    background: rgba(46, 160, 67, 0.1);
  }
  .scan-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .icon {
    width: 18px;
    text-align: center;
    font-size: 13px;
    flex-shrink: 0;
  }

  .icon-ok    { color: var(--color-success); }
  .icon-error { color: var(--color-warning); }
  .icon-unchecked { color: var(--color-border-input); }
  .icon-blank { color: transparent; }
</style>
