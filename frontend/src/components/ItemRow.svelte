<script>
  export let item
  export let checked
  export let onChange
  export let onOpenPicker = () => {}

  let showTooltip = false
  $: tooltipText = item.description || item.notes || ''
  $: isSelective = item.strategy === 'selective'
</script>

<div class="item-row">
  <label>
    <input
      type="checkbox"
      {checked}
      on:click|preventDefault={(e) => {
        if (isSelective && !checked) {
          onOpenPicker(item)
        } else {
          onChange()
        }
      }}
    />
    <span>{item.name}</span>
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
  label {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
    cursor: pointer;
    flex: 1;
    font-size: var(--font-size-sm);
  }
  input[type="checkbox"] { accent-color: var(--color-accent); }
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
