<script>
  import { createEventDispatcher } from 'svelte'

  export let result = null  // PreflightResult | null

  const dispatch = createEventDispatcher()

  let expanded = true
  let runAnyway = false

  $: if (result) { runAnyway = false; expanded = true }  // reset on new result

  $: showRunAnyway = result &&
    result.sourceRootOK &&
    result.destRootOK &&
    result.hasItemWarnings

  function handleRunAnyway() {
    dispatch('runAnyway', runAnyway)
  }
</script>

{#if result}
  <div class="preflight-panel">
    <div class="summary-row">
      <span class="summary-text">
        {#if !result.sourceRootOK}
          <span class="err">⚠ Source root not found</span>
        {:else if !result.destRootOK}
          <span class="err">⚠ Destination root not found</span>
        {:else}
          Pre-flight: <strong>{result.readyCount}/{result.totalCount}</strong> ready
          {#if result.hasItemWarnings}
            <span class="warn"> — {result.totalCount - result.readyCount} source path{result.totalCount - result.readyCount !== 1 ? 's' : ''} not found</span>
          {/if}
        {/if}
      </span>

      {#if result.items && result.items.some(i => !i.found)}
        <button class="toggle-btn" on:click={() => expanded = !expanded}>
          {expanded ? '▲' : '▼'}
        </button>
      {/if}

      {#if showRunAnyway}
        <label class="run-anyway-label">
          <input type="checkbox" bind:checked={runAnyway} on:change={handleRunAnyway} />
          Run anyway
        </label>
      {/if}
    </div>

    {#if expanded && result.items && result.items.length > 0}
      <ul class="item-list">
        {#each result.items.filter(i => !i.found) as item}
          <li class="item-row item-missing">
            <span class="item-icon">↳</span>
            <span class="item-label">{item.label}</span>
            <span class="item-path">[not found at {item.path}]</span>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}

<style>
  .preflight-panel {
    background: var(--color-bg-secondary);
    border-bottom: 1px solid var(--color-border);
    padding: var(--spacing-sm) var(--spacing-2xl);
    font-size: var(--font-size-sm);
  }

  .summary-row {
    display: flex;
    align-items: center;
    gap: var(--spacing-lg);
  }

  .summary-text { color: var(--color-text-secondary); flex: 1; }
  .err  { color: var(--color-danger); }
  .warn { color: var(--color-warning); }

  .toggle-btn {
    background: transparent;
    border: none;
    color: var(--color-border-input);
    cursor: pointer;
    font-size: var(--font-size-xs);
    padding: 0 var(--spacing-xs);
  }

  .run-anyway-label {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    color: var(--color-warning);
    cursor: pointer;
    font-size: var(--font-size-sm);
  }

  .run-anyway-label input { accent-color: var(--color-warning); }

  .item-list {
    list-style: none;
    margin: var(--spacing-xs) 0 0;
    padding: 0 0 0 var(--spacing-xl);
  }

  .item-row {
    display: flex;
    gap: var(--spacing-sm);
    padding: var(--spacing-xs) 0;
    color: var(--color-warning);
  }

  .item-icon { color: var(--color-border-input); }

  .item-label { flex-shrink: 0; }

  .item-path {
    color: var(--color-text-secondary);
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: var(--font-size-xs);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
