<script>
  export let evt

  const statusIcons = {
    copying:  '\u23f3',
    copied:   '\u2705',
    skipped:  '\u23ed\ufe0f',
    failed:   '\u274c',
  }

  let showDetail = false

  $: icon = statusIcons[evt.status] || '\u2753'
  $: isCopying = evt.status === 'copying'
</script>

<div class="progress-item" class:copying={isCopying}>
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
    padding: 0.5rem 0;
    border-bottom: 1px solid #222;
  }

  .progress-item.copying {
    opacity: 0.7;
  }

  .main-line {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.9rem;
  }

  .counter {
    color: #666;
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: 0.8rem;
    min-width: 50px;
  }

  .icon {
    font-size: 1rem;
  }

  .name {
    flex: 1;
    color: #ddd;
  }

  .elapsed {
    color: #666;
    font-size: 0.8rem;
  }

  .detail-toggle {
    background: none;
    border: none;
    color: #2ea043;
    font-size: 0.75rem;
    cursor: pointer;
    padding: 0.2rem 0;
    margin-top: 0.2rem;
  }

  .detail {
    background: #111;
    padding: 0.5rem;
    margin-top: 0.3rem;
    border-radius: 4px;
    font-family: 'Cascadia Code', 'Consolas', monospace;
    font-size: 0.75rem;
    color: #999;
    max-height: 200px;
    overflow-y: auto;
    white-space: pre-wrap;
  }
</style>
