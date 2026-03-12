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
    background: #181818;
    border-bottom: 1px solid #333;
    padding: 6px 20px;
    font-size: 12px;
  }

  .summary-row {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .summary-text { color: #999; flex: 1; }
  .err  { color: #e55; }
  .warn { color: #d4a017; }

  .toggle-btn {
    background: transparent;
    border: none;
    color: #555;
    cursor: pointer;
    font-size: 10px;
    padding: 0 4px;
  }

  .run-anyway-label {
    display: flex;
    align-items: center;
    gap: 5px;
    color: #d4a017;
    cursor: pointer;
    font-size: 12px;
  }

  .run-anyway-label input { accent-color: #d4a017; }

  .item-list {
    list-style: none;
    margin: 4px 0 0;
    padding: 0 0 0 16px;
  }

  .item-row {
    display: flex;
    gap: 6px;
    padding: 2px 0;
    color: #d4a017;
  }

  .item-icon { color: #555; }

  .item-label { flex-shrink: 0; }

  .item-path {
    color: #666;
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: 11px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
