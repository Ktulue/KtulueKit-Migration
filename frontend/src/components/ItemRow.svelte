<script>
  export let item
  export let checked
  export let onChange

  let showTooltip = false

  $: tooltipText = item.description || item.notes || ''
</script>

<div class="item-row">
  <label>
    <input type="checkbox" {checked} on:change={onChange} />
    <span>{item.name}</span>
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
    padding: 0.4rem 0;
    gap: 0.5rem;
  }

  label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    flex: 1;
    font-size: 0.9rem;
  }

  input[type="checkbox"] {
    accent-color: #2ea043;
  }

  .tooltip-trigger {
    position: relative;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: #444;
    color: #999;
    font-size: 0.7rem;
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
    padding: 0.5rem 0.75rem;
    border-radius: 4px;
    font-size: 0.8rem;
    white-space: normal;
    width: 250px;
    z-index: 10;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.4);
  }
</style>
