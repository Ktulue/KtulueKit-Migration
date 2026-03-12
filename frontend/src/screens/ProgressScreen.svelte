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
    padding: 12px 20px;
    background: #111;
    border-bottom: 1px solid #333;
  }

  h2 {
    margin: 0 0 12px 0;
    font-size: 18px;
    color: #e0e0e0;
  }

  .progress-bar-container {
    width: 100%;
    height: 6px;
    background: #333;
    border-radius: 3px;
    overflow: hidden;
  }

  .progress-bar {
    height: 100%;
    background: #2ea043;
    border-radius: 3px;
    transition: width 0.3s ease;
  }

  .progress-text {
    display: block;
    margin-top: 6px;
    font-size: 12px;
    color: #999;
  }

  .dry-run-banner {
    background: #2a2000;
    color: #d4a017;
    padding: 6px 20px;
    font-size: 12px;
    text-align: center;
    border-bottom: 1px solid #3a3000;
  }

  .feed {
    flex: 1;
    overflow-y: auto;
    padding: 12px 20px;
  }
</style>
