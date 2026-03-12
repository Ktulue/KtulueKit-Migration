<script>
  export let result
  export let onClose
  export let onRunAgain = null
  export let manifestPath = ''

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

  function copyManifestPath() {
    if (manifestPath) {
      navigator.clipboard.writeText(manifestPath)
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
    <div class="footer-info">
      <div class="log-path">
        {#if result.logPath}
          <span class="log-label">Log:</span>
          <span class="log-file">{result.logPath}</span>
          <button class="copy-btn" on:click={copyLogPath}>Copy</button>
        {/if}
      </div>
      {#if manifestPath}
        <div class="log-path manifest-path-row">
          <span class="log-label">Manifest:</span>
          <span class="log-file">{manifestPath}</span>
          <button class="copy-btn" on:click={copyManifestPath}>Copy</button>
        </div>
      {/if}
    </div>
    <div class="footer-actions">
      {#if onRunAgain}
        <button class="run-again-btn" on:click={onRunAgain}>Run Again</button>
      {/if}
      <button class="close-btn" on:click={onClose}>Close</button>
    </div>
  </footer>
</div>

<style>
  .summary-screen {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  header {
    padding: 12px 20px;
    background: #111;
    border-bottom: 1px solid #333;
  }

  h2 {
    margin: 0;
    font-size: 18px;
    color: #e0e0e0;
  }

  .stats {
    margin-top: 6px;
    font-size: 13px;
    color: #999;
  }

  .separator {
    margin: 0 8px;
    color: #555;
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: 20px;
  }

  .section {
    margin-bottom: 20px;
  }

  h3 {
    margin: 0 0 8px 0;
    font-size: 14px;
    color: #ccc;
  }

  .item {
    padding: 5px 0 5px 24px;
    font-size: 13px;
    color: #aaa;
    border-bottom: 1px solid #222;
  }

  .manifest-toggle {
    margin: 16px 0;
  }

  .manifest-btn {
    background: transparent;
    color: #2ea043;
    border: 1px solid #2ea043;
    border-radius: 4px;
    padding: 6px 16px;
    font-size: 13px;
    cursor: pointer;
  }

  .manifest-btn:hover {
    background: rgba(46, 160, 67, 0.1);
  }

  .manifest {
    overflow-x: auto;
    margin-top: 8px;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
  }

  th {
    text-align: left;
    padding: 8px;
    background: #222;
    color: #999;
    border-bottom: 1px solid #333;
  }

  td {
    padding: 6px 8px;
    border-bottom: 1px solid #222;
    color: #bbb;
  }

  .path {
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: 12px;
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
    padding: 12px 20px;
    background: #111;
    border-top: 1px solid #333;
  }

  .footer-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .log-path {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
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
    padding: 3px 8px;
    font-size: 12px;
    cursor: pointer;
  }

  .copy-btn:hover {
    color: #e0e0e0;
    border-color: #666;
  }

  .footer-actions {
    display: flex;
    gap: 8px;
  }

  .run-again-btn {
    background: transparent;
    color: #999;
    border: 1px solid #444;
    border-radius: 6px;
    padding: 8px 20px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
  }

  .run-again-btn:hover {
    color: #e0e0e0;
    border-color: #666;
  }

  .close-btn {
    background: #2ea043;
    color: #fff;
    border: none;
    border-radius: 6px;
    padding: 8px 20px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
  }

  .close-btn:hover {
    background: #3ab553;
  }
</style>
