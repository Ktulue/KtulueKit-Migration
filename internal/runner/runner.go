// Package runner orchestrates the migration process, iterating through
// selected items and coordinating the copier, mapper, and reporter.
package runner

import (
	"fmt"
	"os"
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
	App         string
	Label       string
	SourcePath  string
	TargetPath  string
	Status      string
	BytesCopied int64
}

// RunResult captures the full migration outcome.
type RunResult struct {
	Items      []RunResultItem
	TotalBytes int64
	Elapsed    time.Duration
}

// Runner orchestrates the migration process.
type Runner struct {
	cfg         *config.Config
	rep         *reporter.Reporter
	selectedIDs map[string]bool
	onProgress  func(ProgressEvent)
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

// SetOnProgress registers a callback for progress events (used by GUI mode).
func (r *Runner) SetOnProgress(fn func(ProgressEvent)) {
	r.onProgress = fn
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
		targetPath := mapper.BuildTargetPath(w.item.Target)

		// Validate source exists
		if err := mapper.ValidateSourceExists(sourcePath, w.app.Name, w.item.Label); err != nil {
			r.reportItem(w.app.Name, w.item.Label, sourcePath, targetPath, reporter.StatusSkipped, 0, err.Error())
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

		// Determine if source is a file or directory and copy accordingly
		var bytesCopied int64
		var copyErr error

		info, _ := os.Stat(sourcePath)
		if info.IsDir() {
			bytesCopied, copyErr = copier.MirrorDir(sourcePath, targetPath)
		} else {
			bytesCopied, copyErr = copier.CopyFile(sourcePath, targetPath)
		}

		if copyErr != nil {
			r.reportItem(w.app.Name, w.item.Label, sourcePath, targetPath, reporter.StatusFailed, 0, copyErr.Error())
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
		r.reportItem(w.app.Name, w.item.Label, sourcePath, targetPath, reporter.StatusCopied, bytesCopied, "")
		result.Items = append(result.Items, RunResultItem{
			App: w.app.Name, Label: w.item.Label,
			SourcePath: sourcePath, TargetPath: targetPath,
			Status: reporter.StatusCopied, BytesCopied: bytesCopied,
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

func (r *Runner) reportItem(app, label, source, target, status string, bytes int64, detail string) {
	r.rep.Add(reporter.Result{
		App:         app,
		Label:       label,
		SourcePath:  source,
		TargetPath:  target,
		Status:      status,
		BytesCopied: bytes,
		Detail:      detail,
	})
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
