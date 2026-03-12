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
    padding: 6px 0;
    gap: 8px;
  }
  label {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    flex: 1;
    font-size: 13px;
  }
  input[type="checkbox"] { accent-color: #2ea043; }
  .picker-btn {
    background: transparent;
    color: #2ea043;
    border: 1px solid #2ea043;
    border-radius: 3px;
    padding: 1px 8px;
    font-size: 11px;
    cursor: pointer;
  }
  .picker-btn:hover { background: rgba(46,160,67,0.1); }
  .tooltip-trigger {
    position: relative;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: #444;
    color: #999;
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
    background: #333;
    color: #ddd;
    padding: 8px 12px;
    border-radius: 4px;
    font-size: 12px;
    white-space: normal;
    width: 250px;
    z-index: 10;
    box-shadow: 0 2px 8px rgba(0,0,0,0.4);
  }
</style>
