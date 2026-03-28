<script>
  import ItemRow from './ItemRow.svelte'

  export let category
  export let selected
  export let onToggle
  export let onOpenPicker = () => {}
  export let discoveryMap = {}
  export let destMap = {}
  export let onAssist = () => {}
  export let onDetect = () => {}
  export let onDestOverride = () => {}

  let open = true

  $: allChecked = category.items.every(item => selected.has(item.id))
  $: someChecked = category.items.some(item => selected.has(item.id))

  function toggleAll(e) {
    e.stopPropagation()
    if (allChecked) {
      category.items.forEach(item => selected.delete(item.id))
    } else {
      category.items.forEach(item => {
        if (item.strategy === 'selective') return
        const disc = discoveryMap[item.id]
        if (disc && !disc.found) return
        selected.add(item.id)
      })
    }
    onToggle()
  }

  $: discoveredApps = [...new Set(
    category.items
      .filter(item => discoveryMap[item.id]?.found)
      .map(item => item.id.split(':')[0])
  )]
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
      {#each discoveredApps as appName}
        <div class="app-detect-row">
          <span class="app-detect-label">{appName}</span>
          <button class="detect-btn" on:click|stopPropagation={() => onDetect(appName)}>
            Detect
          </button>
        </div>
      {/each}
      {#each category.items as item}
        <ItemRow
          {item}
          checked={selected.has(item.id)}
          discoveryStatus={discoveryMap[item.id] || null}
          destResult={destMap[item.id] || null}
          {onAssist}
          {onDestOverride}
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

  .app-detect-row {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
    padding: var(--spacing-xs) 0;
  }
  .app-detect-label {
    font-size: var(--font-size-xs);
    color: var(--color-text-secondary);
    font-weight: 600;
  }
  .detect-btn {
    background: transparent;
    color: var(--color-accent);
    border: 1px solid var(--color-accent);
    border-radius: var(--radius);
    padding: 2px var(--spacing-sm);
    font-size: var(--font-size-sm);
    cursor: pointer;
    transition: color 100ms ease, border-color 100ms ease, background 100ms ease;
  }
  .detect-btn:hover {
    background: rgba(14, 127, 212, 0.1);
  }
</style>
