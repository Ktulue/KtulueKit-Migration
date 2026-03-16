<script>
  import { afterUpdate } from 'svelte'
  import ProgressItem from '../components/ProgressItem.svelte'

  export let events = []
  export let dryRun = false

  let feedEl

  afterUpdate(() => {
    if (feedEl) {
      feedEl.scrollTop = feedEl.scrollHeight
    }
  })

  $: latestIndex = events.length > 0 ? events[events.length - 1].index : 0
  $: total = events.length > 0 ? events[events.length - 1].total : 1
  $: percent = Math.round((latestIndex / total) * 100)
</script>

<div class="progress-screen">
  <header>
    <h2>Migrating...</h2>
    <div class="progress-bar-container">
      <div class="progress-bar" style="width: {percent}%"></div>
    </div>
    <span class="progress-text">{latestIndex} / {total}</span>
  </header>

  {#if dryRun}
    <div class="dry-run-banner">Dry run — no files will be copied</div>
  {/if}

  <div class="feed" bind:this={feedEl}>
    {#each events as evt}
      <ProgressItem {evt} />
    {/each}
  </div>
</div>

<style>
  .progress-screen {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  header {
    padding: var(--spacing-lg) var(--spacing-2xl);
    background: var(--color-bg-secondary);
    border-bottom: 1px solid var(--color-border);
  }

  h2 {
    margin: 0 0 12px 0;
    font-size: var(--font-size-xl);
    color: var(--color-text-primary);
  }

  .progress-bar-container {
    width: 100%;
    height: 6px;
    background: var(--color-border);
    border-radius: 3px;
    overflow: hidden;
  }

  .progress-bar {
    height: 100%;
    background: var(--color-accent);
    border-radius: 3px;
    transition: width 0.3s ease;
  }

  .progress-text {
    display: block;
    margin-top: 6px;
    font-size: var(--font-size-sm);
    color: var(--color-text-secondary);
  }

  .dry-run-banner {
    background: rgba(230, 168, 23, 0.08);
    color: var(--color-warning);
    padding: var(--spacing-sm) var(--spacing-2xl);
    font-size: var(--font-size-sm);
    text-align: center;
    border-bottom: 1px solid var(--color-border);
  }

  .feed {
    flex: 1;
    overflow-y: auto;
    padding: var(--spacing-lg) var(--spacing-2xl);
  }
</style>
