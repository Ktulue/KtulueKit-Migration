<script>
  export let item
  export let checked
  export let onChange
  export let onOpenPicker = () => {}
  export let discoveryStatus = null
  export let onAssist = () => {}

  let showTooltip = false
  $: tooltipText = item.description || item.notes || ''
  $: isSelective = item.strategy === 'selective'
  $: discovered = discoveryStatus !== null
  $: discoveredFound = discoveryStatus?.found ?? false
  $: discoveredNotFound = discovered && !discoveredFound
</script>

<div class="item-row" class:dimmed={discoveredNotFound}>
  <label>
    <span
      class="checkbox"
      class:checked={checked}
      class:disabled={discoveredNotFound}
      on:click|preventDefault={() => {
        if (discoveredNotFound) return
        if (isSelective && !checked) {
          onOpenPicker(item)
        } else {
          onChange()
        }
      }}
      role="checkbox"
      aria-checked={checked}
      tabindex="0"
      on:keydown={(e) => { if (e.key === ' ' || e.key === 'Enter') { e.preventDefault(); onChange() }}}
    >
      {#if checked}
        <svg width="12" height="12" viewBox="0 0 12 12">
          <polyline points="2,6 5,9 10,3" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      {/if}
    </span>
    <span class="item-name">{item.name}</span>
    {#if discoveredFound}
      <span class="discovery-badge found" title={discoveryStatus.sourcePath}>found</span>
    {:else if discoveredNotFound}
      <span class="discovery-badge not-found">not found</span>
      <button class="assist-btn" on:click|stopPropagation={() => onAssist(item)}>Locate</button>
    {/if}
    {#if isSelective && checked}
      <button class="picker-btn" on:click|stopPropagation={() => onOpenPicker(item)}>
        Edit selection
      </button>
    {/if}
  </label>
  {#if tooltipText}
    <div
      class="tooltip-trigger"
      on:mouseenter={() => showTooltip = true}
      on:mouseleave={() => showTooltip = false}
    >
      ?
      {#if showTooltip}
        <div class="tooltip">{tooltipText}</div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .item-row {
    display: flex;
    align-items: center;
    padding: var(--spacing-sm) 0;
    gap: var(--spacing-md);
  }
  .item-row.dimmed {
    opacity: 0.45;
  }
  label {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
    cursor: pointer;
    flex: 1;
    font-size: var(--font-size-sm);
  }
  .checkbox {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    border: 2px solid var(--color-text-secondary);
    border-radius: 3px;
    flex-shrink: 0;
    cursor: pointer;
    transition: background 100ms ease, border-color 100ms ease;
    color: #fff;
  }
  .checkbox.checked {
    background: var(--color-accent);
    border-color: var(--color-accent);
  }
  .checkbox.disabled {
    border-color: var(--color-border-input);
    cursor: not-allowed;
  }
  .discovery-badge {
    font-size: var(--font-size-xs);
    padding: 1px 6px;
    border-radius: var(--radius);
    font-weight: 600;
  }
  .discovery-badge.found {
    color: var(--color-success);
    background: rgba(46, 160, 67, 0.12);
  }
  .discovery-badge.not-found {
    color: var(--color-text-secondary);
    background: rgba(136, 136, 136, 0.12);
  }
  .assist-btn {
    background: transparent;
    color: var(--color-warning);
    border: 1px solid var(--color-warning);
    border-radius: var(--radius);
    padding: 1px var(--spacing-md);
    font-size: var(--font-size-xs);
    cursor: pointer;
    transition: color 100ms ease, border-color 100ms ease, background 100ms ease;
  }
  .assist-btn:hover { background: rgba(230, 168, 23, 0.1); }
  .picker-btn {
    background: transparent;
    color: var(--color-accent);
    border: 1px solid var(--color-accent);
    border-radius: var(--radius);
    padding: 1px var(--spacing-md);
    font-size: var(--font-size-xs);
    cursor: pointer;
    transition: color 100ms ease, border-color 100ms ease, background 100ms ease;
  }
  .picker-btn:hover { background: rgba(14, 127, 212, 0.1); }
  .tooltip-trigger {
    position: relative;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--color-bg-hover);
    color: var(--color-text-secondary);
    font-size: 11px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: help;
    flex-shrink: 0;
  }
  .tooltip {
    position: absolute;
    bottom: calc(100% + 8px);
    right: 0;
    background: var(--color-bg-hover);
    color: var(--color-text-primary);
    padding: var(--spacing-md) var(--spacing-lg);
    border-radius: var(--radius);
    font-size: var(--font-size-sm);
    white-space: normal;
    width: 250px;
    z-index: 10;
    box-shadow: 0 2px 8px rgba(0,0,0,0.4);
  }
</style>
