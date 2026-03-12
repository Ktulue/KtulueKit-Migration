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
    margin-bottom: 0.25rem;
  }

  .header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.6rem 0.75rem;
    background: #2a2a2a;
    border-radius: 4px;
    cursor: pointer;
    user-select: none;
  }

  .header:hover {
    background: #333;
  }

  .arrow {
    font-size: 0.7rem;
    color: #888;
    transition: transform 0.15s;
  }

  .arrow.open {
    transform: rotate(90deg);
  }

  .name {
    font-weight: 600;
    font-size: 0.95rem;
  }

  .count {
    color: #888;
    font-size: 0.85rem;
  }

  .select-all {
    margin-left: auto;
    background: transparent;
    color: #999;
    border: 1px solid #555;
    border-radius: 4px;
    padding: 0.15rem 0.5rem;
    font-size: 0.75rem;
    cursor: pointer;
  }

  .select-all:hover {
    color: #e0e0e0;
    border-color: #888;
  }

  .items {
    padding-left: 1.5rem;
  }
</style>
