package main

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/Ktulue/KtulueKit-Migration/internal/config"
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
		Categories: categories,
		Profiles:   profiles,
	}, nil
}

// StartMigration validates selections and kicks off the migration in a goroutine.
// Progress events are emitted to the frontend via Wails runtime events.
func (a *App) StartMigration(selectedIDs []string) error {
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

		rep := reporter.New("logs")

		r := runner.New(cfg, rep)
		r.SetSelectedIDs(selectedIDs)
		r.SetOnProgress(func(evt runner.ProgressEvent) {
			runtime.EventsEmit(a.ctx, "progress", evt)
		})

		result := r.Run()

		runtime.EventsEmit(a.ctx, "complete", SummaryResult{
			Copied:   rep.NamesBy(reporter.StatusCopied),
			Skipped:  rep.NamesBy(reporter.StatusSkipped),
			Failed:   rep.NamesBy(reporter.StatusFailed),
			Bytes:    result.TotalBytes,
			Elapsed:  result.Elapsed.String(),
			LogPath:  rep.LogPath(),
			Manifest: buildManifest(result),
		})
	}()

	return nil
}

// buildManifest converts runner results into ManifestEntry slices for the summary.
func buildManifest(result runner.RunResult) []ManifestEntry {
	var entries []ManifestEntry
	for _, r := range result.Items {
		entries = append(entries, ManifestEntry{
			App:         r.App,
			Label:       r.Label,
			SourcePath:  r.SourcePath,
			TargetPath:  r.TargetPath,
			Status:      r.Status,
			BytesCopied: r.BytesCopied,
		})
	}
	return entries
}
