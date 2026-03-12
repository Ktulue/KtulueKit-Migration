// Package runner orchestrates the migration process, iterating through
// selected items and coordinating the copier, mapper, and reporter.
package runner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/Ktulue/KtulueKit-Migration/internal/config"
	"github.com/Ktulue/KtulueKit-Migration/internal/copier"
	"github.com/Ktulue/KtulueKit-Migration/internal/mapper"
	"github.com/Ktulue/KtulueKit-Migration/internal/reporter"
)

// ProgressEvent is emitted to the frontend to report real-time progress.
type ProgressEvent struct {
	Index   int    `json:"index"`
	Total   int    `json:"total"`
	App     string `json:"app"`
	Label   string `json:"label"`
	Status  string `json:"status"`
	Detail  string `json:"detail"`
	Elapsed string `json:"elapsed"`
}

// RunResultItem records the outcome of a single item copy.
type RunResultItem struct {
	App           string
	Label         string
	SourcePath    string
	TargetPath    string
	Status        string
	BytesCopied   int64
	SelectedPaths []string
}

// RunResult captures the full migration outcome.
type RunResult struct {
	Items      []RunResultItem
	TotalBytes int64
	Elapsed    time.Duration
}

// Runner orchestrates the migration process.
type Runner struct {
	cfg              *config.Config
	rep              *reporter.Reporter
	selectedIDs      map[string]bool
	selectivePaths   map[string][]string
	dryRun           bool
	onProgress       func(ProgressEvent)
	destRootOverride string
}

// New creates a Runner with the given config and reporter.
func New(cfg *config.Config, rep *reporter.Reporter) *Runner {
	return &Runner{
		cfg: cfg,
		rep: rep,
	}
}

// SetSelectedIDs filters the migration to only the given item IDs.
func (r *Runner) SetSelectedIDs(ids []string) {
	r.selectedIDs = make(map[string]bool, len(ids))
	for _, id := range ids {
		r.selectedIDs[id] = true
	}
}

// SetSelectivePaths sets per-item path selections for selective strategy items.
func (r *Runner) SetSelectivePaths(paths map[string][]string) {
	r.selectivePaths = paths
}

// SetDryRun enables dry-run mode — paths are resolved but no files are written.
func (r *Runner) SetDryRun(dryRun bool) {
	r.dryRun = dryRun
}

// SetOnProgress registers a callback for progress events (used by GUI mode).
func (r *Runner) SetOnProgress(fn func(ProgressEvent)) {
	r.onProgress = fn
}

// SetDestRootOverride sets a runtime destination root override applied to all target paths.
func (r *Runner) SetDestRootOverride(override string) {
	r.destRootOverride = override
}

// Run executes the migration. It iterates through all selected items,
// resolves paths, copies files, and reports results.
func (r *Runner) Run() RunResult {
	start := time.Now()
	var result RunResult

	// Build the flat list of work items from selected IDs
	type workItem struct {
		app  config.App
		item config.Item
		id   string
	}
	var work []workItem
	for _, app := range r.cfg.Apps {
		for _, item := range app.Items {
			id := app.Name + ":" + item.Label
			if r.selectedIDs != nil && !r.selectedIDs[id] {
				continue
			}
			work = append(work, workItem{app: app, item: item, id: id})
		}
	}

	total := len(work)

	for i, w := range work {
		elapsed := time.Since(start)

		// Emit "copying" progress event
		r.emitProgress(ProgressEvent{
			Index:   i + 1,
			Total:   total,
			App:     w.app.Name,
			Label:   w.item.Label,
			Status:  "copying",
			Elapsed: elapsed.Round(time.Second).String(),
		})

		// Resolve source and target paths
		sourcePath := mapper.BuildSourcePath(r.cfg.BackupRoot, w.item.Source)

		// resolvedTarget is the fully env-var-expanded absolute path from the config.
		// If env-var expansion succeeded, it will match X:\ on Windows.
		// If it doesn't (e.g. unexpanded %UNKNOWN_VAR%), we treat it as invalid.
		resolvedTarget := mapper.BuildTargetPath(w.item.Target)

		// Guard: if a dest override is active and the resolved target has no drive prefix,
		// log as failed and skip — do not copy to a garbage path.
		if r.destRootOverride != "" {
			if len(resolvedTarget) < 3 || resolvedTarget[1] != ':' || resolvedTarget[2] != '\\' {
				r.reportItemFull(w.app.Name, w.item.Label, sourcePath, resolvedTarget, reporter.StatusFailed, 0, "target path has no drive prefix", nil)
				result.Items = append(result.Items, RunResultItem{
					App: w.app.Name, Label: w.item.Label,
					SourcePath: sourcePath, TargetPath: resolvedTarget,
					Status: reporter.StatusFailed,
				})
				r.emitProgress(ProgressEvent{
					Index: i + 1, Total: total,
					App: w.app.Name, Label: w.item.Label,
					Status: "failed", Detail: "target path has no drive prefix",
					Elapsed: time.Since(start).Round(time.Second).String(),
				})
				continue
			}
		}
		targetPath := mapper.ApplyDestOverride(resolvedTarget, r.destRootOverride)

		// Validate source exists
		if err := mapper.ValidateSourceExists(sourcePath, w.app.Name, w.item.Label); err != nil {
			r.reportItemFull(w.app.Name, w.item.Label, sourcePath, targetPath, reporter.StatusSkipped, 0, err.Error(), nil)
			result.Items = append(result.Items, RunResultItem{
				App: w.app.Name, Label: w.item.Label,
				SourcePath: sourcePath, TargetPath: targetPath,
				Status: reporter.StatusSkipped,
			})

			r.emitProgress(ProgressEvent{
				Index: i + 1, Total: total,
				App: w.app.Name, Label: w.item.Label,
				Status: "skipped", Detail: err.Error(),
				Elapsed: time.Since(start).Round(time.Second).String(),
			})
			continue
		}

		// Determine copy strategy
		strategy := w.item.Strategy
		if strategy == "" {
			strategy = "mirror"
		}

		var bytesCopied int64
		var copyErr error
		var selectedPaths []string

		if r.dryRun {
			// Dry-run: estimate size without copying
			bytesCopied, copyErr = estimateSize(sourcePath)
		} else if strategy == "selective" {
			itemID := w.app.Name + ":" + w.item.Label
			selectedPaths = r.selectivePaths[itemID]
			if len(selectedPaths) == 0 {
				// No paths chosen in picker — skip rather than silently copying nothing
				r.reportItemFull(w.app.Name, w.item.Label, sourcePath, targetPath, reporter.StatusSkipped, 0, "no paths selected in folder picker", nil)
				result.Items = append(result.Items, RunResultItem{
					App: w.app.Name, Label: w.item.Label,
					SourcePath: sourcePath, TargetPath: targetPath,
					Status: reporter.StatusSkipped,
				})
				r.emitProgress(ProgressEvent{
					Index: i + 1, Total: total,
					App: w.app.Name, Label: w.item.Label,
					Status: "skipped", Detail: "no paths selected in folder picker",
					Elapsed: time.Since(start).Round(time.Second).String(),
				})
				continue
			}
			for _, p := range selectedPaths {
				n, err := copier.CopyPath(p, filepath.Join(targetPath, filepath.Base(p)))
				bytesCopied += n
				if err != nil {
					copyErr = err
					break
				}
			}
		} else {
			// mirror or file strategy
			info, statErr := os.Stat(sourcePath)
			if statErr != nil {
				copyErr = statErr
			} else if info.IsDir() {
				bytesCopied, copyErr = copier.MirrorDir(sourcePath, targetPath)
			} else {
				bytesCopied, copyErr = copier.CopyFile(sourcePath, targetPath)
			}
		}

		if copyErr != nil {
			r.reportItemFull(w.app.Name, w.item.Label, sourcePath, targetPath, reporter.StatusFailed, 0, copyErr.Error(), nil)
			result.Items = append(result.Items, RunResultItem{
				App: w.app.Name, Label: w.item.Label,
				SourcePath: sourcePath, TargetPath: targetPath,
				Status: reporter.StatusFailed,
			})

			r.emitProgress(ProgressEvent{
				Index: i + 1, Total: total,
				App: w.app.Name, Label: w.item.Label,
				Status: "failed", Detail: copyErr.Error(),
				Elapsed: time.Since(start).Round(time.Second).String(),
			})
			continue
		}

		// Success
		result.TotalBytes += bytesCopied
		r.reportItemFull(w.app.Name, w.item.Label, sourcePath, targetPath, reporter.StatusCopied, bytesCopied, "", selectedPaths)
		result.Items = append(result.Items, RunResultItem{
			App: w.app.Name, Label: w.item.Label,
			SourcePath: sourcePath, TargetPath: targetPath,
			Status: reporter.StatusCopied, BytesCopied: bytesCopied,
			SelectedPaths: selectedPaths,
		})

		r.emitProgress(ProgressEvent{
			Index: i + 1, Total: total,
			App: w.app.Name, Label: w.item.Label,
			Status:  "copied",
			Detail:  fmt.Sprintf("%s copied", formatBytes(bytesCopied)),
			Elapsed: time.Since(start).Round(time.Second).String(),
		})
	}

	result.Elapsed = time.Since(start)
	r.rep.Summary()
	r.rep.Close()

	return result
}

func (r *Runner) emitProgress(evt ProgressEvent) {
	if r.onProgress != nil {
		r.onProgress(evt)
	}
}

func (r *Runner) reportItemFull(app, label, source, target, status string, bytes int64, detail string, selectedPaths []string) {
	r.rep.Add(reporter.Result{
		App:           app,
		Label:         label,
		SourcePath:    source,
		TargetPath:    target,
		Status:        status,
		BytesCopied:   bytes,
		Detail:        detail,
		SelectedPaths: selectedPaths,
	})
}

func estimateSize(path string) (int64, error) {
	var total int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	return total, err
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
