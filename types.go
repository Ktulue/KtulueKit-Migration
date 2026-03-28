package main

// ConfigView is the display model sent to the frontend.
type ConfigView struct {
	BackupRoot string        `json:"backupRoot"`
	Categories []CategoryView `json:"categories"`
	Profiles   []ProfileView  `json:"profiles"`
}

// categoryOrder defines the display sequence for migration categories.
var categoryOrder = []string{
	"Streaming",
	"Dev Tools",
	"Creative Suite",
	"Utilities",
	"Browser & Identity",
	"Communication",
	"Media Assets",
	"Shell & Terminal",
	"Games",
	"Personal Files",
}

// CategoryView represents a named group of migration items.
type CategoryView struct {
	Name  string     `json:"name"`
	Items []ItemView `json:"items"`
}

// ItemView is a single selectable migration item shown in the UI.
type ItemView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
	Strategy    string `json:"strategy"`
}

// FolderEntry represents a single file or directory inside a listed folder.
type FolderEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

// ProfileView is a named preset that selects a subset of migration items.
type ProfileView struct {
	Name string   `json:"name"`
	IDs  []string `json:"ids"`
}

// SummaryResult captures the outcome of a migration run.
type SummaryResult struct {
	Copied                []string        `json:"copied"`
	Skipped               []string        `json:"skipped"`
	Failed                []string        `json:"failed"`
	Bytes                 int64           `json:"bytes"`
	Elapsed               string          `json:"elapsed"`
	LogPath               string          `json:"logPath"`
	ManifestPath          string          `json:"manifestPath"`
	SourceRootOverride    string          `json:"sourceRootOverride,omitempty"`
	DestRootOverride      string          `json:"destRootOverride,omitempty"`
	Manifest              []ManifestEntry `json:"manifest"`
}

// ManifestEntry records a single copy operation for the user to review.
type ManifestEntry struct {
	App           string   `json:"app"`
	Label         string   `json:"label"`
	SourcePath    string   `json:"sourcePath"`
	TargetPath    string   `json:"targetPath"`
	Status        string   `json:"status"`
	BytesCopied   int64    `json:"bytesCopied"`
	SelectedPaths []string `json:"selectedPaths"`
}

// PreflightResult is returned by the PreflightCheck backend method.
type PreflightResult struct {
	SourceRootOK    bool            `json:"sourceRootOK"`
	DestRootOK      bool            `json:"destRootOK"`
	HasItemWarnings bool            `json:"hasItemWarnings"`
	Items           []PreflightItem `json:"items"`
	ReadyCount      int             `json:"readyCount"`
	TotalCount      int             `json:"totalCount"`
}

// PreflightItem records the check result for a single selected item.
type PreflightItem struct {
	ID       string `json:"id"`              // app.Name + ":" + item.Label
	Label    string `json:"label"`           // app.Name + " — " + item.Label
	Path     string `json:"path"`            // resolved source path actually checked
	Found    bool   `json:"found"`
	DestPath string `json:"destPath,omitempty"` // override destination path (if set)
	DestOK   bool   `json:"destOK"`             // false means parent dir not accessible
}

// DetectResultView is the display model for a single item's destination detection result.
type DetectResultView struct {
	ItemID     string   `json:"itemId"`
	DestPath   string   `json:"destPath"`
	Method     string   `json:"method"`
	Confirmed  bool     `json:"confirmed"`
	Candidates []string `json:"candidates"`
}
