package discovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Ktulue/KtulueKit-Migration/internal/config"
)

// DiscoveredItem represents a single config item and whether it was found on the scanned drive.
type DiscoveredItem struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	SourcePath string   `json:"sourcePath"`
	Found      bool     `json:"found"`
	Partial    []string `json:"partial"`
}

// Result holds the outcome of a drive scan.
type Result struct {
	Items      []DiscoveredItem `json:"items"`
	FoundCount int              `json:"foundCount"`
	TotalCount int              `json:"totalCount"`
}

var systemProfiles = map[string]bool{
	"default":      true,
	"public":       true,
	"all users":    true,
	"default user": true,
}

var envVarPattern = regexp.MustCompile(`%([^%]+)%`)

// Scan examines a cloned Windows drive at drivePath, resolving each config item's
// target path against every real user profile under Users\. It picks the profile
// with the most hits and returns the results.
func Scan(ctx context.Context, drivePath string, cfg *config.Config) (*Result, error) {
	type itemInfo struct {
		id     string
		label  string
		target string
	}
	var allItems []itemInfo
	for _, app := range cfg.Apps {
		for _, item := range app.Items {
			allItems = append(allItems, itemInfo{
				id:     app.Name + ":" + item.Label,
				label:  app.Name + " — " + item.Label,
				target: item.Target,
			})
		}
	}

	totalCount := len(allItems)

	usersDir := filepath.Join(drivePath, "Users")
	profiles, err := listRealProfiles(usersDir)
	if err != nil || len(profiles) == 0 {
		items := make([]DiscoveredItem, totalCount)
		for i, info := range allItems {
			items[i] = DiscoveredItem{
				ID:      info.id,
				Label:   info.label,
				Found:   false,
				Partial: []string{},
			}
		}
		return &Result{Items: items, FoundCount: 0, TotalCount: totalCount}, nil
	}

	type profileScore struct {
		name  string
		path  string
		count int
		items []DiscoveredItem
	}

	var best profileScore
	for _, prof := range profiles {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		profPath := filepath.Join(usersDir, prof)
		envMap := buildEnvMap(drivePath, profPath)
		var items []DiscoveredItem
		found := 0

		for _, info := range allItems {
			resolved := resolveTargetWithMap(info.target, envMap)
			exists := pathExists(resolved)
			if exists {
				found++
			}
			items = append(items, DiscoveredItem{
				ID:         info.id,
				Label:      info.label,
				SourcePath: resolved,
				Found:      exists,
				Partial:    []string{},
			})
		}

		if found > best.count {
			best = profileScore{name: prof, path: profPath, count: found, items: items}
		}
	}

	// Clear SourcePath for items that were not found.
	for i := range best.items {
		if !best.items[i].Found {
			best.items[i].SourcePath = ""
		}
	}

	return &Result{
		Items:      best.items,
		FoundCount: best.count,
		TotalCount: totalCount,
	}, nil
}

// listRealProfiles returns directory names under usersDir, excluding system profiles.
func listRealProfiles(usersDir string) ([]string, error) {
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading Users directory: %w", err)
	}

	var profiles []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if systemProfiles[strings.ToLower(e.Name())] {
			continue
		}
		profiles = append(profiles, e.Name())
	}
	return profiles, nil
}

// buildEnvMap creates a mapping of Windows environment variable names to paths
// rooted under the given profile directory on the cloned drive.
func buildEnvMap(drivePath, profilePath string) map[string]string {
	return map[string]string{
		"APPDATA":      filepath.Join(profilePath, "AppData", "Roaming"),
		"LOCALAPPDATA": filepath.Join(profilePath, "AppData", "Local"),
		"USERPROFILE":  profilePath,
	}
}

// resolveTargetWithMap replaces %VAR% tokens in target with values from envMap.
func resolveTargetWithMap(target string, envMap map[string]string) string {
	resolved := envVarPattern.ReplaceAllStringFunc(target, func(match string) string {
		varName := strings.Trim(match, "%")
		if val, ok := envMap[strings.ToUpper(varName)]; ok {
			return val
		}
		return match
	})
	return filepath.FromSlash(resolved)
}

// pathExists returns true if the given path exists on disk.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
