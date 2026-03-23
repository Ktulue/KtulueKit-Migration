<script>
  import { onMount } from 'svelte'
  import SelectionScreen from './screens/SelectionScreen.svelte'
  import ProgressScreen from './screens/ProgressScreen.svelte'
  import SummaryScreen from './screens/SummaryScreen.svelte'
  import FolderPicker from './components/FolderPicker.svelte'
  import { GetConfig, StartMigration, GetSourcePath } from '../wailsjs/go/main/App'
  import { EventsOn, Quit } from '../wailsjs/runtime/runtime'

  let screen = 'selection' // 'selection' | 'progress' | 'summary'
  let configView = null
  let configError = null
  let progressEvents = []
  let summaryResult = null
  let selectivePaths = {}
  let pickerItem = null
  let pickerSourcePath = ''
  let pickerConfirmedCallback = null
  let dryRun = false
  let pendingSourceRoot = ''
  let pendingDestRoot = ''

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

  async function handleStartMigration(selectedIDs, userSelectivePaths, isDryRun, sourceRoot, destRoot, sourcePathMap) {
    dryRun = isDryRun
    progressEvents = []
    screen = 'progress'
    try {
      await StartMigration(selectedIDs, { ...selectivePaths, ...userSelectivePaths }, isDryRun, sourceRoot || '', destRoot || '', sourcePathMap || {})
    } catch (err) {
      summaryResult = { failed: [err.toString()], copied: [], skipped: [], manifest: [] }
      screen = 'summary'
    }
  }

  async function handleOpenPicker(item, onConfirmed) {
    pickerConfirmedCallback = onConfirmed || null
    try {
      pickerSourcePath = await GetSourcePath(item.id)
      pickerItem = item
    } catch (err) {
      console.error('Could not resolve source path for picker:', err)
      pickerConfirmedCallback = null
    }
  }

  function handlePickerConfirm(itemId, paths) {
    selectivePaths = { ...selectivePaths, [itemId]: paths }
    if (pickerConfirmedCallback) {
      pickerConfirmedCallback(itemId)
      pickerConfirmedCallback = null
    }
    pickerItem = null
    pickerSourcePath = ''
  }

  function handlePickerCancel() {
    pickerConfirmedCallback = null
    pickerItem = null
    pickerSourcePath = ''
  }

  function handleProfileChange() {
    selectivePaths = {}
  }

  function handleRunAgain() {
    pendingSourceRoot = summaryResult?.sourceRootOverride ?? ''
    pendingDestRoot = summaryResult?.destRootOverride ?? ''
    summaryResult = null
    progressEvents = []
    screen = 'selection'
  }

  function handleClose() {
    Quit()
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
      initialSourceRoot={pendingSourceRoot}
      initialDestRoot={pendingDestRoot}
      onStart={handleStartMigration}
      onOpenPicker={handleOpenPicker}
      onProfileChange={handleProfileChange}
    />
  {:else if screen === 'progress'}
    <ProgressScreen events={progressEvents} {dryRun} />
  {:else if screen === 'summary' && summaryResult}
    <SummaryScreen result={summaryResult} onClose={handleClose} onRunAgain={handleRunAgain} manifestPath={summaryResult.manifestPath || ''} />
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
  @font-face {
    font-family: 'Nunito';
    src: url('./assets/fonts/nunito-400.woff2') format('woff2');
    font-weight: 400;
    font-style: normal;
    font-display: swap;
  }

  @font-face {
    font-family: 'Nunito';
    src: url('./assets/fonts/nunito-600.woff2') format('woff2');
    font-weight: 600;
    font-style: normal;
    font-display: swap;
  }

  @font-face {
    font-family: 'Nunito';
    src: url('./assets/fonts/nunito-700.woff2') format('woff2');
    font-weight: 700;
    font-style: normal;
    font-display: swap;
  }

  :root {
    /* Font */
    --font-primary: 'Nunito', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;

    /* Typography scale */
    --font-size-xs:   11px;
    --font-size-sm:   12px;
    --font-size-base: 15px;
    --font-size-lg:   16px;
    --font-size-xl:   18px;
    --font-size-2xl:  20px;

    /* Shape */
    --radius: 4px;

    /* Spacing (4px grid) */
    --spacing-xs:  4px;
    --spacing-sm:  6px;
    --spacing-md:  8px;
    --spacing-lg:  12px;
    --spacing-xl:  16px;
    --spacing-2xl: 20px;

    /* Colors — backgrounds */
    --color-bg-primary:   #1a1a1a;
    --color-bg-secondary: #111;
    --color-bg-hover:     #2a2a2a;

    /* Colors — borders */
    --color-border:       #333;
    --color-border-input: #555;

    /* Colors — text */
    --color-text-primary:   #e0e0e0;
    --color-text-secondary: #888;
    --color-text-tertiary:  #aaa;

    /* Colors — accent (blue) */
    --color-accent:          #0e7fd4;
    --color-accent-hover:    #1290e8;
    --color-accent-disabled: #444;

    /* Colors — danger */
    --color-danger:        #ff6b6b;
    --color-danger-action: #c0392b;

    /* Colors — success (Migration-specific) */
    --color-success:       #2ea043;
    --color-success-hover: #3ab854;

    /* Colors — warning (Migration-specific) */
    --color-warning:       #e6a817;
    --color-warning-hover: #f0b929;
  }

  :global(*, *::before, *::after) {
    box-sizing: border-box;
  }

  :global(body) {
    margin: 0;
    font-family: var(--font-primary);
    background: var(--color-bg-primary);
    color: var(--color-text-primary);
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
    color: var(--color-danger);
  }
</style>
