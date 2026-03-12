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

  function formatSize(bytes) {
    if (bytes === 0) return '—'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1048576) return `${(bytes/1024).toFixed(1)} KB`
    return `${(bytes/1048576).toFixed(1)} MB`
  }
</script>

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
    background: #1e1e1e;
    border: 1px solid #333;
    border-radius: 6px;
    width: 560px;
    max-height: 70vh;
    display: flex;
    flex-direction: column;
  }
  .modal-header {
    padding: 16px 20px 12px;
    border-bottom: 1px solid #333;
  }
  .modal-header h3 { margin: 0 0 4px; font-size: 15px; }
  .path {
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: 11px;
    color: #666;
    word-break: break-all;
  }
  .modal-toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px 20px;
    background: #181818;
    border-bottom: 1px solid #2a2a2a;
  }
  .select-all-btn {
    background: transparent;
    color: #999;
    border: 1px solid #444;
    border-radius: 4px;
    padding: 3px 10px;
    font-size: 12px;
    cursor: pointer;
  }
  .select-all-btn:hover:not(:disabled) { color: #e0e0e0; border-color: #666; }
  .select-all-btn:disabled { opacity: 0.4; cursor: not-allowed; }
  .count { font-size: 12px; color: #666; }
  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: 4px 0;
  }
  .entry-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 20px;
    cursor: pointer;
    font-size: 13px;
  }
  .entry-row:hover { background: #2a2a2a; }
  input[type="checkbox"] { accent-color: #2ea043; flex-shrink: 0; }
  .entry-icon { font-size: 14px; }
  .entry-name { flex: 1; color: #ddd; }
  .entry-size { font-size: 11px; color: #666; font-family: 'Cascadia Code', 'Consolas', monospace; }
  .state-msg { padding: 20px; text-align: center; color: #666; }
  .state-msg.error { color: #e55; }
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 12px 20px;
    border-top: 1px solid #333;
    background: #181818;
  }
  .cancel-btn {
    background: transparent;
    color: #999;
    border: 1px solid #444;
    border-radius: 4px;
    padding: 8px 20px;
    cursor: pointer;
  }
  .cancel-btn:hover { color: #e0e0e0; border-color: #666; }
  .confirm-btn {
    background: #2ea043;
    color: #fff;
    border: none;
    border-radius: 4px;
    padding: 8px 20px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
  }
  .confirm-btn:disabled { background: #333; color: #666; cursor: not-allowed; }
  .confirm-btn:not(:disabled):hover { background: #3ab553; }
</style>
