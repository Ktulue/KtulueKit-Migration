package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/Ktulue/KtulueKit-Migration/internal/config"
	"github.com/Ktulue/KtulueKit-Migration/internal/detector"
	"github.com/Ktulue/KtulueKit-Migration/internal/discovery"
	"github.com/Ktulue/KtulueKit-Migration/internal/mapper"
	"github.com/Ktulue/KtulueKit-Migration/internal/reporter"
	"github.com/Ktulue/KtulueKit-Migration/internal/runner"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct bound to the Wails frontend.
type App struct {
	ctx        context.Context
	configPath string
	mu         sync.Mutex
	running    bool
}

// NewApp creates a new App instance with the given config path.
func NewApp(configPath string) *App {
	return &App{
		configPath: configPath,
	}
}

// startup is called when the Wails app starts. Saves the runtime context.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetConfig loads the migration config and transforms it into a display model
// for the frontend. Items are grouped by category and sorted alphabetically.
func (a *App) GetConfig() (*ConfigView, error) {
	cfg, err := config.Load(a.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Build category map from apps
	catMap := make(map[string][]ItemView)
	for _, app := range cfg.Apps {
		for _, item := range app.Items {
			iv := ItemView{
				ID:          app.Name + ":" + item.Label,
				Name:        app.Name + " — " + item.Label,
				Description: item.Description,
				Notes:       item.Notes,
				Strategy:    item.Strategy,
			}
			catMap[app.Category] = append(catMap[app.Category], iv)
		}
	}

	// Sort items within each category
	for cat := range catMap {
		items := catMap[cat]
		sort.Slice(items, func(i, j int) bool {
			return items[i].Name < items[j].Name
		})
	}

	// Build ordered category list
	var categories []CategoryView
	for _, cat := range categoryOrder {
		if items, ok := catMap[cat]; ok {
			categories = append(categories, CategoryView{Name: cat, Items: items})
		}
	}
	// Catch any categories not in the predefined order
	for cat, items := range catMap {
		found := false
		for _, ordered := range categoryOrder {
			if cat == ordered {
				found = true
				break
			}
		}
		if !found {
			categories = append(categories, CategoryView{Name: cat, Items: items})
		}
	}

	// Build profiles
	var profiles []ProfileView
	for _, p := range cfg.Profiles {
		profiles = append(profiles, ProfileView{Name: p.Name, IDs: p.IDs})
	}

	return &ConfigView{
		BackupRoot: cfg.BackupRoot,
		Categories: categories,
		Profiles:   profiles,
	}, nil
}

// StartMigration validates selections and kicks off the migration in a goroutine.
// Progress events are emitted to the frontend via Wails runtime events.
func (a *App) StartMigration(selectedIDs []string, selectivePaths map[string][]string, dryRun bool, sourceRootOverride string, destRootOverride string, sourcePathMap map[string]string, destPathMap map[string]string) error {
	if len(selectedIDs) == 0 {
		return fmt.Errorf("no items selected")
	}

	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("migration already in progress")
	}
	a.running = true
	a.mu.Unlock()

	// Capture overrides into locals before the goroutine to avoid closure issues.
	srcOverride := sourceRootOverride
	dstOverride := destRootOverride
	srcPaths := sourcePathMap
	dstPaths := destPathMap

	go func() {
		defer func() {
			a.mu.Lock()
			a.running = false
			a.mu.Unlock()
		}()

		cfg, err := config.Load(a.configPath)
		if err != nil {
			runtime.EventsEmit(a.ctx, "complete", SummaryResult{
				Failed: []string{"config load: " + err.Error()},
			})
			return
		}

		var rep *reporter.Reporter
		if dryRun {
			rep = reporter.NewNull()
		} else {
			rep = reporter.New("logs")
		}

		// Shallow copy is safe here: BackupRoot is a string (value type);
		// Apps slice header is copied but the underlying array is only read, never mutated.
		cfgCopy := *cfg
		if srcOverride != "" {
			cfgCopy.BackupRoot = srcOverride
		}
		r := runner.New(&cfgCopy, rep)
		r.SetSelectedIDs(selectedIDs)
		r.SetSelectivePaths(selectivePaths)
		r.SetDryRun(dryRun)
		r.SetDestRootOverride(dstOverride)
		r.SetSourcePathMap(srcPaths)
		r.SetDestPathMap(dstPaths)
		r.SetOnProgress(func(evt runner.ProgressEvent) {
			runtime.EventsEmit(a.ctx, "progress", evt)
		})

		result := r.Run()

		var manifestPath string
		if !dryRun {
			ts := rep.Timestamp().Format("2006-01-02_15-04-05")
			manifestPath = filepath.Join("logs", fmt.Sprintf("manifest_%s.json", ts))
			if err := rep.WriteManifest(manifestPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not write manifest: %v\n", err)
				manifestPath = ""
			}
		}

		runtime.EventsEmit(a.ctx, "complete", SummaryResult{
			Copied:             rep.NamesBy(reporter.StatusCopied),
			Skipped:            rep.NamesBy(reporter.StatusSkipped),
			Failed:             rep.NamesBy(reporter.StatusFailed),
			Bytes:              result.TotalBytes,
			Elapsed:            result.Elapsed.String(),
			LogPath:            rep.LogPath(),
			ManifestPath:       manifestPath,
			SourceRootOverride: srcOverride,
			DestRootOverride:   dstOverride,
			Manifest:           buildManifest(result),
		})
	}()

	return nil
}

// buildManifest converts runner results into ManifestEntry slices for the summary.
func buildManifest(result runner.RunResult) []ManifestEntry {
	var entries []ManifestEntry
	for _, r := range result.Items {
		sp := r.SelectedPaths
		if sp == nil {
			sp = []string{}
		}
		entries = append(entries, ManifestEntry{
			App:           r.App,
			Label:         r.Label,
			SourcePath:    r.SourcePath,
			TargetPath:    r.TargetPath,
			Status:        r.Status,
			BytesCopied:   r.BytesCopied,
			SelectedPaths: sp,
		})
	}
	return entries
}

// ListFolder returns the immediate contents of the directory at path.
// Used by the frontend FolderPicker component.
func (a *App) ListFolder(path string) ([]FolderEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("reading folder: %w", err)
	}
	result := make([]FolderEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		var size int64
		if !e.IsDir() {
			size = info.Size()
		}
		result = append(result, FolderEntry{
			Name:  e.Name(),
			Path:  filepath.Join(path, e.Name()),
			IsDir: e.IsDir(),
			Size:  size,
		})
	}
	return result, nil
}

// ValidateBackupRoot checks whether the configured backup root directory exists.
// Returns true if it exists and is accessible, false otherwise.
func (a *App) ValidateBackupRoot() (bool, error) {
	cfg, err := config.Load(a.configPath)
	if err != nil {
		return false, err
	}
	resolved := mapper.ResolvePath(cfg.BackupRoot)
	info, err := os.Stat(resolved)
	if err != nil {
		return false, nil // not found — not an error, just absent
	}
	return info.IsDir(), nil
}

// BrowseForFolder opens a native OS directory picker dialog and returns
// the selected folder path, or an empty string if cancelled.
func (a *App) BrowseForFolder(startPath string) (string, error) {
	selected, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		DefaultDirectory: startPath,
		Title:            "Select Folder",
	})
	if err != nil {
		return "", err
	}
	return selected, nil
}

// ScanDrive scans a drive path for app data matching the config items.
// Used by the frontend to auto-discover source paths on a cloned drive.
func (a *App) ScanDrive(drivePath string) (*discovery.Result, error) {
	cfg, err := config.Load(a.configPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return discovery.Scan(a.ctx, drivePath, cfg)
}

// PreflightCheck validates the source root, destination root, and all selected
// item source paths before allowing the user to start migration.
func (a *App) PreflightCheck(selectedIDs []string, sourceRoot string, destRoot string, sourcePathMap map[string]string, destPathMap map[string]string) (PreflightResult, error) {
	cfg, err := config.Load(a.configPath)
	if err != nil {
		return PreflightResult{}, fmt.Errorf("loading config: %w", err)
	}

	result := PreflightResult{}

	// Check 1: source root
	if info, err := os.Stat(sourceRoot); err == nil && info.IsDir() {
		result.SourceRootOK = true
	}

	// Check 2: destination root.
	// DestRootOK = true means migration can proceed:
	//   - blank destRoot → no override, always OK
	//   - destRoot exists as dir → OK (already present)
	//   - destRoot doesn't exist but parent does → OK (will be created at run time; drive is mounted)
	// DestRootOK = false means hard block: the drive/parent is not accessible.
	if destRoot == "" {
		result.DestRootOK = true
	} else {
		if info, err := os.Stat(destRoot); err == nil && info.IsDir() {
			result.DestRootOK = true
		} else {
			parent := filepath.Dir(strings.TrimRight(destRoot, `\/`))
			if info, err := os.Stat(parent); err == nil && info.IsDir() {
				result.DestRootOK = true
			}
		}
	}

	// Build selected ID set
	selectedSet := make(map[string]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		selectedSet[id] = true
	}

	// Check 3: per-item source paths
	for _, app := range cfg.Apps {
		for _, item := range app.Items {
			id := app.Name + ":" + item.Label
			if !selectedSet[id] {
				continue
			}
			result.TotalCount++

			sourcePath := ""
			if sourcePathMap != nil {
				sourcePath = sourcePathMap[id]
			}
			if sourcePath == "" {
				sourcePath = mapper.BuildSourcePath(sourceRoot, item.Source)
			}
			_, statErr := os.Stat(sourcePath)
			found := statErr == nil

			destPath := ""
			destOK := true
			if destPathMap != nil && destPathMap[id] != "" {
				destPath = destPathMap[id]
				parent := filepath.Dir(strings.TrimRight(destPath, `\/`))
				if _, err := os.Stat(parent); err != nil {
					destOK = false
				}
			}

			label := app.Name + " — " + item.Label
			if !found {
				result.HasItemWarnings = true
			} else {
				result.ReadyCount++
			}

			result.Items = append(result.Items, PreflightItem{
				ID:       id,
				Label:    label,
				Path:     sourcePath,
				Found:    found,
				DestPath: destPath,
				DestOK:   destOK,
			})
		}
	}

	return result, nil
}

// GetSourcePath resolves the backup source path for a given item ID.
// Used by the frontend FolderPicker to list folder contents.
func (a *App) GetSourcePath(itemID string) (string, error) {
	cfg, err := config.Load(a.configPath)
	if err != nil {
		return "", err
	}
	for _, app := range cfg.Apps {
		for _, item := range app.Items {
			if app.Name+":"+item.Label == itemID {
				return mapper.BuildSourcePath(cfg.BackupRoot, item.Source), nil
			}
		}
	}
	return "", fmt.Errorf("item not found: %s", itemID)
}

// DetectDestination runs destination detection for a single app's items.
// sourcePathMap contains the discovered source paths (itemID -> absolute path).
// Returns detection results per item.
func (a *App) DetectDestination(appName string, sourcePathMap map[string]string) ([]DetectResultView, error) {
	cfg, err := config.Load(a.configPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// Find the app in config
	var detection *config.Detection
	for _, app := range cfg.Apps {
		if app.Name == appName {
			detection = app.Detection
			break
		}
	}

	// Get current user's profile path
	localProfile := os.Getenv("USERPROFILE")
	if localProfile == "" {
		return nil, fmt.Errorf("USERPROFILE environment variable not set")
	}

	results := detector.Detect(appName, sourcePathMap, detection, localProfile)

	var views []DetectResultView
	for _, r := range results {
		views = append(views, DetectResultView{
			ItemID:     r.ItemID,
			DestPath:   r.DestPath,
			Method:     r.Method,
			Confirmed:  r.Confirmed,
			Candidates: r.Candidates,
		})
	}

	return views, nil
}
