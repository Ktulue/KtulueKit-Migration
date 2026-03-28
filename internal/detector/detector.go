package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Ktulue/KtulueKit-Migration/internal/config"
)

// DetectResult holds the outcome of destination detection for a single item.
type DetectResult struct {
	ItemID     string   `json:"itemId"`
	DestPath   string   `json:"destPath"`
	Method     string   `json:"method"`
	Confirmed  bool     `json:"confirmed"`
	Candidates []string `json:"candidates"`
}

// knownSegments maps path segments after Users\<name>\ to themselves.
var knownSegments = []string{
	filepath.Join("AppData", "LocalLow"),
	filepath.Join("AppData", "Local"),
	filepath.Join("AppData", "Roaming"),
	"Documents",
	"Desktop",
	"Videos",
	"Pictures",
	"Music",
	"Downloads",
}

// PatternMap attempts to remap a discovered source path to the equivalent
// local path by detecting known segments under Users\<name>\ and replacing
// the drive + username with the current user's profile path.
//
// Returns the remapped path and true if a known pattern was matched.
func PatternMap(sourcePath string, localProfile string) (string, bool) {
	normalized := filepath.FromSlash(sourcePath)
	lower := strings.ToLower(normalized)
	usersIdx := strings.Index(lower, strings.ToLower("Users"+string(filepath.Separator)))
	if usersIdx < 0 {
		return "", false
	}

	afterUsers := normalized[usersIdx+len("Users"+string(filepath.Separator)):]
	sepIdx := strings.Index(afterUsers, string(filepath.Separator))
	if sepIdx < 0 {
		return "", false
	}

	afterUsername := afterUsers[sepIdx+1:]
	if afterUsername == "" {
		return "", false
	}

	// Accept any path under Users\<name>\ — known segments get matched,
	// but dotfiles and other direct children are also accepted
	dest := filepath.Join(localProfile, afterUsername)
	return dest, true
}

// Detect resolves the destination path for items of a given app.
// sourcePaths maps itemID -> discovered source path.
func Detect(appName string, sourcePaths map[string]string, detection *config.Detection, localProfile string) []DetectResult {
	var results []DetectResult

	for itemID, srcPath := range sourcePaths {
		result := DetectResult{
			ItemID:     itemID,
			Candidates: []string{},
		}

		// Tier 1: pattern mapping
		if dest, ok := PatternMap(srcPath, localProfile); ok {
			result.DestPath = dest
			result.Method = "path-mapping"
			_, err := os.Stat(dest)
			result.Confirmed = err == nil
			results = append(results, result)
			continue
		}

		// Tier 2 placeholder — will be filled in Task 3
		// For now, just return empty result
		result.Method = ""
		results = append(results, result)
	}

	return results
}
