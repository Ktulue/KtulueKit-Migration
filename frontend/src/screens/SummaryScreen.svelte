<script>
  export let result
  export let onClose

  let showManifest = false

  const statusSections = [
    { key: 'copied',  label: 'Copied',  icon: '\u2705' },
    { key: 'skipped', label: 'Skipped', icon: '\u23ed\ufe0f' },
    { key: 'failed',  label: 'Failed',  icon: '\u274c' },
  ]

  function copyLogPath() {
    if (result.logPath) {
      navigator.clipboard.writeText(result.logPath)
    }
  }

  function formatBytes(b) {
    if (!b || b === 0) return '0 B'
    const units = ['B', 'KB', 'MB', 'GB']
    let i = 0
    let val = b
    while (val >= 1024 && i < units.length - 1) {
      val /= 1024
      i++
    }
    return `${val.toFixed(i > 0 ? 1 : 0)} ${units[i]}`
  }
</script>

<div class="summary-screen">
  <header>
    <h2>Migration Summary</h2>
    <div class="stats">
      <span>{result.elapsed || '—'} elapsed</span>
      <span class="separator">|</span>
      <span>{formatBytes(result.bytes)} transferred</span>
    </div>
  </header>

  <div class="content">
    {#each statusSections as section}
      {#if result[section.key] && result[section.key].length > 0}
        <div class="section">
          <h3>{section.icon} {section.label} ({result[section.key].length})</h3>
          {#each result[section.key] as name}
            <div class="item">{name}</div>
          {/each}
        </div>
      {/if}
    {/each}

    {#if result.manifest && result.manifest.length > 0}
      <div class="manifest-toggle">
        <button class="manifest-btn" on:click={() => showManifest = !showManifest}>
          {showManifest ? 'Hide' : 'Show'} Full Manifest
        </button>
      </div>

      {#if showManifest}
        <div class="manifest">
          <table>
            <thead>
              <tr>
                <th>App</th>
                <th>Item</th>
                <th>Source</th>
                <th>Target</th>
                <th>Status</th>
                <th>Size</th>
              </tr>
            </thead>
            <tbody>
              {#each result.manifest as entry}
                <tr class="status-{entry.status}">
                  <td>{entry.app}</td>
                  <td>{entry.label}</td>
                  <td class="path">{entry.sourcePath}</td>
                  <td class="path">{entry.targetPath}</td>
                  <td>{entry.status}</td>
                  <td>{formatBytes(entry.bytesCopied)}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    {/if}
  </div>

  <footer>
    <div class="log-path">
      {#if result.logPath}
        <span class="log-label">Log:</span>
        <span class="log-file">{result.logPath}</span>
        <button class="copy-btn" on:click={copyLogPath}>Copy</button>
      {/if}
    </div>
    <button class="close-btn" on:click={onClose}>Close</button>
  </footer>
</div>

<style>
  .summary-screen {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  header {
    padding: 1rem 1.5rem;
    background: #111;
    border-bottom: 1px solid #333;
  }

  h2 {
    margin: 0;
    font-size: 1.2rem;
    color: #e0e0e0;
  }

  .stats {
    margin-top: 0.4rem;
    font-size: 0.85rem;
    color: #999;
  }

  .separator {
    margin: 0 0.5rem;
    color: #555;
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: 1rem 1.5rem;
  }

  .section {
    margin-bottom: 1.2rem;
  }

  h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
    color: #ccc;
  }

  .item {
    padding: 0.3rem 0 0.3rem 1.5rem;
    font-size: 0.9rem;
    color: #aaa;
    border-bottom: 1px solid #222;
  }

  .manifest-toggle {
    margin: 1rem 0;
  }

  .manifest-btn {
    background: transparent;
    color: #2ea043;
    border: 1px solid #2ea043;
    border-radius: 4px;
    padding: 0.4rem 1rem;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .manifest-btn:hover {
    background: rgba(46, 160, 67, 0.1);
  }

  .manifest {
    overflow-x: auto;
    margin-top: 0.5rem;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.8rem;
  }

  th {
    text-align: left;
    padding: 0.5rem;
    background: #222;
    color: #999;
    border-bottom: 1px solid #333;
  }

  td {
    padding: 0.4rem 0.5rem;
    border-bottom: 1px solid #222;
    color: #bbb;
  }

  .path {
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: 0.75rem;
    color: #888;
    max-width: 250px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  tr.status-copied td:nth-child(5) { color: #2ea043; }
  tr.status-skipped td:nth-child(5) { color: #d4a017; }
  tr.status-failed td:nth-child(5) { color: #e55; }

  footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    background: #111;
    border-top: 1px solid #333;
  }

  .log-path {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.8rem;
  }

  .log-label {
    color: #666;
  }

  .log-file {
    color: #999;
    font-family: 'Cascadia Code', 'Consolas', monospace;
  }

  .copy-btn {
    background: transparent;
    color: #999;
    border: 1px solid #444;
    border-radius: 4px;
    padding: 0.2rem 0.6rem;
    font-size: 0.75rem;
    cursor: pointer;
  }

  .copy-btn:hover {
    color: #e0e0e0;
    border-color: #666;
  }

  .close-btn {
    background: #2ea043;
    color: #fff;
    border: none;
    border-radius: 6px;
    padding: 0.6rem 1.5rem;
    font-size: 0.95rem;
    font-weight: 600;
    cursor: pointer;
  }

  .close-btn:hover {
    background: #3ab553;
  }
</style>
