<script>
  import { fade } from 'svelte/transition'

  export let evt

  const statusIcons = {
    copying:  '⏳',
    copied:   '✅',
    skipped:  '⏭️',
    failed:   '❌',
  }

  let showDetail = false

  $: icon = statusIcons[evt.status] || '❓'
  $: isCopying = evt.status === 'copying'
</script>

<div class="progress-item" class:copying={isCopying} in:fade={{ duration: 150 }}>
  <div class="main-line">
    <span class="counter">[{evt.index}/{evt.total}]</span>
    <span class="icon">{icon}</span>
    <span class="name">{evt.app} — {evt.label}</span>
    <span class="elapsed">{evt.elapsed}</span>
  </div>

  {#if evt.detail}
    <button class="detail-toggle" on:click={() => showDetail = !showDetail}>
      {showDetail ? 'Hide' : 'Details'}
    </button>
    {#if showDetail}
      <div class="detail">{evt.detail}</div>
    {/if}
  {/if}
</div>

<style>
  .progress-item {
    padding: var(--spacing-md) 0;
    border-bottom: 1px solid var(--color-bg-hover);
  }

  .progress-item.copying {
    opacity: 0.7;
  }

  .main-line {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
    font-size: var(--font-size-base);
  }

  .counter {
    color: var(--color-text-secondary);
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: var(--font-size-sm);
    min-width: 50px;
  }

  .icon {
    font-size: var(--font-size-base);
  }

  .name {
    flex: 1;
    color: var(--color-text-primary);
  }

  .elapsed {
    color: var(--color-text-secondary);
    font-size: var(--font-size-sm);
  }

  .detail-toggle {
    background: none;
    border: none;
    color: var(--color-accent);
    font-size: var(--font-size-sm);
    cursor: pointer;
    padding: 2px 0;
    margin-top: 2px;
    transition: color 100ms ease;
  }

  .detail {
    background: var(--color-bg-secondary);
    padding: var(--spacing-md);
    margin-top: var(--spacing-xs);
    border-radius: var(--radius);
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: var(--font-size-sm);
    color: var(--color-text-secondary);
    max-height: 200px;
    overflow-y: auto;
    white-space: pre-wrap;
  }
</style>
