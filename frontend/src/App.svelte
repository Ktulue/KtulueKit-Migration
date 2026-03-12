<script>
  import { onMount } from 'svelte'
  import SelectionScreen from './screens/SelectionScreen.svelte'
  import ProgressScreen from './screens/ProgressScreen.svelte'
  import SummaryScreen from './screens/SummaryScreen.svelte'
  import FolderPicker from './components/FolderPicker.svelte'
  import { GetConfig, StartMigration, GetSourcePath } from '../wailsjs/go/main/App'
  import { EventsOn } from '../wailsjs/runtime/runtime'

  let screen = 'selection' // 'selection' | 'progress' | 'summary'
  let configView = null
  let configError = null
  let progressEvents = []
  let summaryResult = null
  let selectivePaths = {}
  let pickerItem = null
  let pickerSourcePath = ''
  let dryRun = false

  onMount(async () => {
    try {
      configView = await GetConfig()
    } catch (err) {
      configError = err
    }

    EventsOn('progress', (evt) => {
      progressEvents = [...progressEvents, evt]
    })

    EventsOn('complete', (result) => {
      summaryResult = result
      screen = 'summary'
    })
  })

  async function handleStartMigration(selectedIDs, userSelectivePaths, isDryRun) {
    dryRun = isDryRun
    progressEvents = []
    screen = 'progress'
    try {
      await StartMigration(selectedIDs, { ...selectivePaths, ...userSelectivePaths }, isDryRun)
    } catch (err) {
      summaryResult = { failed: [err.toString()], copied: [], skipped: [], manifest: [] }
      screen = 'summary'
    }
  }

  async function handleOpenPicker(item) {
    try {
      pickerSourcePath = await GetSourcePath(item.id)
      pickerItem = item
    } catch (err) {
      console.error('Could not resolve source path for picker:', err)
    }
  }

  function handlePickerConfirm(itemId, paths) {
    selectivePaths = { ...selectivePaths, [itemId]: paths }
    pickerItem = null
    pickerSourcePath = ''
  }

  function handlePickerCancel() {
    pickerItem = null
    pickerSourcePath = ''
  }

  function handleClose() {
    window.runtime.Quit()
  }
</script>

<main>
  {#if configError}
    <div class="error">
      <h2>Configuration Error</h2>
      <p>{configError}</p>
    </div>
  {:else if screen === 'selection' && configView}
    <SelectionScreen
      {configView}
      onStart={handleStartMigration}
      onOpenPicker={handleOpenPicker}
    />
  {:else if screen === 'progress'}
    <ProgressScreen events={progressEvents} {dryRun} />
  {:else if screen === 'summary' && summaryResult}
    <SummaryScreen result={summaryResult} onClose={handleClose} manifestPath={summaryResult.manifestPath || ''} />
  {/if}

  {#if pickerItem}
    <FolderPicker
      sourcePath={pickerSourcePath}
      itemId={pickerItem.id}
      onConfirm={handlePickerConfirm}
      onCancel={handlePickerCancel}
    />
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: #1a1a1a;
    color: #e0e0e0;
  }

  main {
    height: 100vh;
    overflow: hidden;
  }

  .error {
    padding: 2rem;
    text-align: center;
  }

  .error h2 {
    color: #e55;
  }
</style>
