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
  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(4px); }
    to   { opacity: 1; transform: translateY(0); }
  }

  .summary-screen {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  header {
    padding: var(--spacing-lg) var(--spacing-2xl);
    background: var(--color-bg-secondary);
    border-bottom: 1px solid var(--color-border);
  }

  h2 {
    margin: 0;
    font-size: var(--font-size-xl);
    color: var(--color-text-primary);
  }

  .stats {
    margin-top: var(--spacing-sm);
    font-size: var(--font-size-sm);
    color: var(--color-text-secondary);
  }

  .separator {
    margin: 0 var(--spacing-md);
    color: var(--color-border-input);
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: var(--spacing-2xl);
  }

  .section {
    margin-bottom: var(--spacing-2xl);
    animation: fadeIn 150ms ease both;
  }

  .section:nth-child(1) { animation-delay: 0ms; }
  .section:nth-child(2) { animation-delay: 50ms; }
  .section:nth-child(3) { animation-delay: 100ms; }

  h3 {
    margin: 0 0 var(--spacing-md) 0;
    font-size: var(--font-size-base);
    color: var(--color-text-primary);
  }

  .item {
    padding: var(--spacing-sm) 0 var(--spacing-sm) var(--spacing-2xl);
    font-size: var(--font-size-sm);
    color: var(--color-text-tertiary);
    border-bottom: 1px solid var(--color-bg-hover);
  }

  .manifest-toggle {
    margin: var(--spacing-xl) 0;
  }

  .manifest-btn {
    background: transparent;
    color: var(--color-accent);
    border: 1px solid var(--color-accent);
    border-radius: var(--radius);
    padding: var(--spacing-sm) var(--spacing-xl);
    font-size: var(--font-size-sm);
    cursor: pointer;
    transition: background 100ms ease;
  }

  .manifest-btn:hover {
    background: rgba(14, 127, 212, 0.1);
  }

  .manifest {
    overflow-x: auto;
    margin-top: var(--spacing-md);
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: var(--font-size-sm);
  }

  th {
    text-align: left;
    padding: var(--spacing-md);
    background: var(--color-bg-hover);
    color: var(--color-text-secondary);
    border-bottom: 1px solid var(--color-border);
  }

  td {
    padding: var(--spacing-sm) var(--spacing-md);
    border-bottom: 1px solid var(--color-bg-hover);
    color: var(--color-text-tertiary);
  }

  .path {
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: var(--font-size-sm);
    color: var(--color-text-secondary);
    max-width: 250px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  tr.status-copied td:nth-child(5) { color: var(--color-success); }
  tr.status-skipped td:nth-child(5) { color: var(--color-warning); }
  tr.status-failed td:nth-child(5) { color: var(--color-danger); }

  footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--spacing-lg) var(--spacing-2xl);
    background: var(--color-bg-secondary);
    border-top: 1px solid var(--color-border);
  }

  .footer-info {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-xs);
  }

  .log-path {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
    font-size: var(--font-size-sm);
  }

  .log-label {
    color: var(--color-text-secondary);
  }

  .log-file {
    color: var(--color-text-secondary);
    font-family: 'Cascadia Code', 'Consolas', monospace;
  }

  .copy-btn {
    background: transparent;
    color: var(--color-text-tertiary);
    border: 1px solid var(--color-border-input);
    border-radius: var(--radius);
    padding: var(--spacing-xs) var(--spacing-md);
    font-size: var(--font-size-sm);
    cursor: pointer;
    transition: color 100ms ease, border-color 100ms ease;
  }

  .copy-btn:hover {
    color: var(--color-text-primary);
    border-color: var(--color-border-input);
  }

  .footer-actions {
    display: flex;
    gap: var(--spacing-md);
  }

  .run-again-btn {
    background: var(--color-bg-hover);
    color: var(--color-text-secondary);
    border: 1px solid var(--color-border);
    border-radius: var(--radius);
    padding: var(--spacing-md) var(--spacing-2xl);
    font-size: var(--font-size-base);
    font-weight: 600;
    cursor: pointer;
    transition: color 100ms ease, border-color 100ms ease;
  }

  .run-again-btn:hover {
    color: var(--color-text-primary);
    border-color: var(--color-border-input);
  }

  .close-btn {
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

  .close-btn:hover {
    background: var(--color-accent-hover);
  }
</style>
