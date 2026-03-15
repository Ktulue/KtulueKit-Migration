<script>
  import ItemRow from './ItemRow.svelte'

  export let category
  export let selected
  export let onToggle
  export let onOpenPicker = () => {}

  let open = true

  $: allChecked = category.items.every(item => selected.has(item.id))
  $: someChecked = category.items.some(item => selected.has(item.id))

  function toggleAll(e) {
    e.stopPropagation()
    if (allChecked) {
      category.items.forEach(item => selected.delete(item.id))
    } else {
      category.items.forEach(item => {
        if (item.strategy === 'selective') return  // must use picker
        selected.add(item.id)
      })
    }
    onToggle()
  }
</script>

<div class="accordion">
  <div class="header" on:click={() => open = !open}>
    <span class="arrow" class:open>{'\u25b6'}</span>
    <span class="name">{category.name}</span>
    <span class="count">({category.items.length})</span>
    <button class="select-all" on:click={toggleAll}>
      {allChecked ? 'Deselect all' : 'Select all'}
    </button>
  </div>

  {#if open}
    <div class="items">
      {#each category.items as item}
        <ItemRow
          {item}
          checked={selected.has(item.id)}
          onChange={() => {
            if (selected.has(item.id)) {
              selected.delete(item.id)
            } else {
              selected.add(item.id)
            }
            onToggle()
          }}
          {onOpenPicker}
        />
      {/each}
    </div>
  {/if}
</div>

<style>
  .accordion {
    margin-bottom: var(--spacing-xs);
  }

  .header {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
    padding: var(--spacing-sm) var(--spacing-lg);
    background: var(--color-bg-hover);
    border-radius: var(--radius);
    cursor: pointer;
    user-select: none;
    transition: background 100ms ease;
  }

  .header:hover {
    background: var(--color-border-input);
  }

  .arrow {
    font-size: var(--font-size-sm);
    color: var(--color-text-secondary);
    transition: transform 150ms ease;
  }

  .arrow.open {
    transform: rotate(90deg);
  }

  .name {
    font-weight: 600;
    font-size: var(--font-size-base);
  }

  .count {
    color: var(--color-text-secondary);
    font-size: var(--font-size-sm);
  }

  .select-all {
    margin-left: auto;
    background: transparent;
    color: var(--color-text-secondary);
    border: 1px solid var(--color-border-input);
    border-radius: var(--radius);
    padding: 2px var(--spacing-sm);
    font-size: var(--font-size-sm);
    cursor: pointer;
    transition: color 100ms ease, border-color 100ms ease;
  }

  .select-all:hover {
    color: var(--color-text-primary);
    border-color: var(--color-text-secondary);
  }

  .items {
    padding-left: var(--spacing-2xl);
  }
</style>
