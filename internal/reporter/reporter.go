// Package reporter manages migration result tracking and log file output.
// Mirrors the pattern from KtulueKit-W11's reporter package.
package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Status constants for migration outcomes.
const (
	StatusCopied  = "copied"
	StatusSkipped = "skipped"
	StatusFailed  = "failed"
)

// Result records the outcome of a single migration item.
type Result struct {
	App           string
	Label         string
	SourcePath    string
	TargetPath    string
	Status        string
	BytesCopied   int64
	Detail        string
	SelectedPaths []string
}

// Reporter collects migration results and maintains a log file.
type Reporter struct {
	results   []Result
	logFile   *os.File
	logPath   string
	timestamp time.Time
}

// New creates a Reporter and opens a timestamped log file in the given directory.
func New(logDir string) *Reporter {
	_ = os.MkdirAll(logDir, 0755)

	ts := time.Now()
	logPath := filepath.Join(logDir, fmt.Sprintf("migration_%s.log", ts.Format("2006-01-02_15-04-05")))

	f, err := os.Create(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not create log file: %v\n", err)
	}

	r := &Reporter{
		logFile:   f,
		logPath:   logPath,
		timestamp: ts,
	}

	r.LogLine("KtulueKit Migration — Run started at %s", ts.UTC().Format(time.RFC3339))
	r.LogLine("---")

	return r
}

// NewNull returns a Reporter that discards all output. Used for dry-run mode.
func NewNull() *Reporter {
	return &Reporter{timestamp: time.Now()}
}

// Add records a result and writes it to both stdout and the log file.
func (r *Reporter) Add(res Result) {
	r.results = append(r.results, res)

	icon := statusIcon(res.Status)
	line := fmt.Sprintf("%s %s — %s: %s", icon, res.App, res.Label, res.Status)
	if res.Detail != "" {
		line += fmt.Sprintf(" (%s)", res.Detail)
	}
	if res.BytesCopied > 0 {
		line += fmt.Sprintf(" [%s]", formatBytes(res.BytesCopied))
	}

	fmt.Println(line)
	r.LogLine(line)
	r.LogLine("  Source: %s", res.SourcePath)
	r.LogLine("  Target: %s", res.TargetPath)
}

// Summary prints a categorized final report.
func (r *Reporter) Summary() {
	r.LogLine("---")
	r.LogLine("MIGRATION SUMMARY")

	for _, status := range []string{StatusCopied, StatusSkipped, StatusFailed} {
		names := r.NamesBy(status)
		if len(names) == 0 {
			continue
		}
		r.LogLine("\n%s %s (%d):", statusIcon(status), status, len(names))
		for _, name := range names {
			r.LogLine("  - %s", name)
		}
	}
}

// NamesBy returns the display names of all results matching the given status.
func (r *Reporter) NamesBy(status string) []string {
	var names []string
	for _, res := range r.results {
		if res.Status == status {
			names = append(names, res.App+" — "+res.Label)
		}
	}
	return names
}

// Results returns all recorded results.
func (r *Reporter) Results() []Result {
	return r.results
}

// HasFailures returns true if any items failed.
func (r *Reporter) HasFailures() bool {
	for _, res := range r.results {
		if res.Status == StatusFailed {
			return true
		}
	}
	return false
}

// LogPath returns the path to the current log file.
func (r *Reporter) LogPath() string {
	return r.logPath
}

// LogLine writes a formatted line to the log file only.
func (r *Reporter) LogLine(format string, args ...interface{}) {
	if r.logFile == nil {
		return
	}
	fmt.Fprintf(r.logFile, format+"\n", args...)
}

// Close flushes and closes the log file.
func (r *Reporter) Close() {
	if r.logFile != nil {
		r.logFile.Close()
	}
}

// manifestItem is the JSON structure for a single manifest entry.
type manifestItem struct {
	App           string   `json:"app"`
	Label         string   `json:"label"`
	SourcePath    string   `json:"sourcePath"`
	TargetPath    string   `json:"targetPath"`
	Status        string   `json:"status"`
	BytesCopied   int64    `json:"bytesCopied"`
	SelectedPaths []string `json:"selectedPaths"`
}

// manifest is the top-level JSON structure written for KtulueKit-Cleanup.
type manifest struct {
	Version string         `json:"version"`
	RunAt   string         `json:"runAt"`
	Items   []manifestItem `json:"items"`
}

// WriteManifest serializes all results to a JSON file at the given path.
// This is the contract consumed by KtulueKit-Cleanup.
func (r *Reporter) WriteManifest(path string) error {
	items := make([]manifestItem, 0, len(r.results))
	for _, res := range r.results {
		sp := res.SelectedPaths
		if sp == nil {
			sp = []string{}
		}
		items = append(items, manifestItem{
			App:           res.App,
			Label:         res.Label,
			SourcePath:    res.SourcePath,
			TargetPath:    res.TargetPath,
			Status:        res.Status,
			BytesCopied:   res.BytesCopied,
			SelectedPaths: sp,
		})
	}

	data, err := json.MarshalIndent(manifest{
		Version: "1.0",
		RunAt:   r.timestamp.UTC().Format(time.RFC3339),
		Items:   items,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling manifest: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func statusIcon(status string) string {
	switch status {
	case StatusCopied:
		return "\u2705" // green check
	case StatusSkipped:
		return "\u23ed\ufe0f" // skip
	case StatusFailed:
		return "\u274c" // red X
	default:
		return "\u2753" // question mark
	}
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
