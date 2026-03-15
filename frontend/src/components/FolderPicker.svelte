<script>
  import { ListFolder } from '../../wailsjs/go/main/App'

  export let sourcePath = ''
  export let itemId = ''
  export let onConfirm = (id, paths) => {}
  export let onCancel = () => {}

  let entries = []
  let selected = new Set()
  let loading = true
  let error = null

  $: if (sourcePath) {
    loading = true
    error = null
    ListFolder(sourcePath)
      .then(e => { entries = e; loading = false })
      .catch(err => { error = err.toString(); loading = false })
  }

  $: allChecked = entries.length > 0 && entries.every(e => selected.has(e.path))

  function toggleAll() {
    if (allChecked) {
      selected = new Set()
    } else {
      selected = new Set(entries.map(e => e.path))
    }
  }

  function toggle(path) {
    selected = new Set(selected)
    if (selected.has(path)) selected.delete(path)
    else selected.add(path)
  }

  function confirm() {
    onConfirm(itemId, [...selected])
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') onCancel()
  }

  function formatSize(bytes) {
    if (bytes === 0) return '—'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1048576) return `${(bytes/1024).toFixed(1)} KB`
    return `${(bytes/1048576).toFixed(1)} MB`
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="overlay">
  <div class="modal">
    <div class="modal-header">
      <h3>Select items to migrate</h3>
      <span class="path">{sourcePath}</span>
    </div>

    <div class="modal-toolbar">
      <button class="select-all-btn" on:click={toggleAll} disabled={loading}>
        {allChecked ? 'Deselect all' : 'Select all'}
      </button>
      <span class="count">{selected.size} selected</span>
    </div>

    <div class="modal-body">
      {#if loading}
        <p class="state-msg">Loading...</p>
      {:else if error}
        <p class="state-msg error">{error}</p>
      {:else if entries.length === 0}
        <p class="state-msg">Folder is empty</p>
      {:else}
        {#each entries as entry}
          <label class="entry-row">
            <input
              type="checkbox"
              checked={selected.has(entry.path)}
              on:change={() => toggle(entry.path)}
            />
            <span class="entry-icon">{entry.isDir ? '📁' : '📄'}</span>
            <span class="entry-name">{entry.name}</span>
            <span class="entry-size">{formatSize(entry.size)}</span>
          </label>
        {/each}
      {/if}
    </div>

    <div class="modal-footer">
      <button class="cancel-btn" on:click={onCancel}>Cancel</button>
      <button
        class="confirm-btn"
        disabled={selected.size === 0}
        on:click={confirm}
      >
        Confirm ({selected.size})
      </button>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0,0,0,0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }
  .modal {
    background: var(--color-bg-primary);
    border: 1px solid var(--color-border);
    border-radius: var(--radius);
    width: 560px;
    max-height: 70vh;
    display: flex;
    flex-direction: column;
  }
  .modal-header {
    padding: var(--spacing-xl) var(--spacing-2xl) 12px;
    border-bottom: 1px solid var(--color-border);
  }
  .modal-header h3 { margin: 0 0 var(--spacing-xs); font-size: var(--font-size-base); }
  .path {
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: var(--font-size-xs);
    color: var(--color-text-secondary);
    word-break: break-all;
  }
  .modal-toolbar {
    display: flex;
    align-items: center;
    gap: var(--spacing-lg);
    padding: var(--spacing-md) var(--spacing-2xl);
    background: var(--color-bg-secondary);
    border-bottom: 1px solid var(--color-bg-hover);
  }
  .select-all-btn {
    background: transparent;
    color: var(--color-text-secondary);
    border: 1px solid var(--color-border-input);
    border-radius: var(--radius);
    padding: var(--spacing-xs) var(--spacing-lg);
    font-size: var(--font-size-sm);
    cursor: pointer;
    transition: color 100ms ease, border-color 100ms ease;
  }
  .select-all-btn:hover:not(:disabled) { color: var(--color-text-primary); border-color: var(--color-text-secondary); }
  .select-all-btn:disabled { opacity: 0.4; cursor: not-allowed; }
  .count { font-size: var(--font-size-sm); color: var(--color-text-secondary); }
  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: var(--spacing-xs) 0;
  }
  .entry-row {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
    padding: var(--spacing-sm) var(--spacing-2xl);
    cursor: pointer;
    font-size: var(--font-size-sm);
    transition: background 100ms ease;
  }
  .entry-row:hover { background: var(--color-bg-hover); }
  input[type="checkbox"] { accent-color: var(--color-accent); flex-shrink: 0; }
  .entry-icon { font-size: var(--font-size-base); }
  .entry-name { flex: 1; color: var(--color-text-primary); }
  .entry-size { font-size: var(--font-size-xs); color: var(--color-text-secondary); font-family: 'Cascadia Code', 'Consolas', monospace; }
  .state-msg { padding: var(--spacing-2xl) 0 var(--spacing-md) 0; text-align: center; color: var(--color-text-secondary); }
  .state-msg.error { color: var(--color-danger); }
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--spacing-md);
    padding: var(--spacing-lg) var(--spacing-2xl);
    border-top: 1px solid var(--color-border);
    background: var(--color-bg-secondary);
  }
  .cancel-btn {
    background: transparent;
    color: var(--color-text-secondary);
    border: 1px solid var(--color-border-input);
    border-radius: var(--radius);
    padding: var(--spacing-md) var(--spacing-2xl);
    cursor: pointer;
    transition: color 100ms ease, border-color 100ms ease;
  }
  .cancel-btn:hover { color: var(--color-text-primary); border-color: var(--color-text-secondary); }
  .confirm-btn {
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
  .confirm-btn:disabled { background: var(--color-accent-disabled); color: var(--color-text-secondary); cursor: not-allowed; }
  .confirm-btn:not(:disabled):hover { background: var(--color-accent-hover); }
</style>
