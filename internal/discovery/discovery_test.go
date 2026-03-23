package discovery

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ktulue/KtulueKit-Migration/internal/config"
)

func setupFakeDrive(t *testing.T) (string, *config.Config) {
	t.Helper()
	root := t.TempDir()

	user := filepath.Join(root, "Users", "TestUser")
	dirs := []string{
		filepath.Join(user, "AppData", "Roaming", "obs-studio", "basic"),
		filepath.Join(user, "AppData", "Local", "BraveSoftware", "Brave-Browser", "User Data", "Default"),
		filepath.Join(user, ".ssh"),
		filepath.Join(user, "Documents"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{
		Version:    "1.0",
		BackupRoot: "D:\\Backup",
		Apps: []config.App{
			{
				Name:     "OBS Studio",
				Category: "Streaming",
				Items: []config.Item{
					{Label: "scenes & profiles", Source: "obs-studio/basic", Target: "%APPDATA%/obs-studio/basic"},
				},
			},
			{
				Name:     "Brave Browser",
				Category: "Browser & Identity",
				Items: []config.Item{
					{Label: "user profile", Source: "brave/User Data/Default", Target: "%LOCALAPPDATA%/BraveSoftware/Brave-Browser/User Data/Default"},
				},
			},
			{
				Name:     "SSH",
				Category: "Dev Tools",
				Items: []config.Item{
					{Label: "keys & config", Source: "ssh", Target: "%USERPROFILE%/.ssh"},
				},
			},
			{
				Name:     "Missing App",
				Category: "Other",
				Items: []config.Item{
					{Label: "not here", Source: "missing", Target: "%APPDATA%/DoesNotExist"},
				},
			},
		},
	}

	return root, cfg
}

func TestScan_FindsItemsInProfile(t *testing.T) {
	root, cfg := setupFakeDrive(t)
	result, err := Scan(context.Background(), root, cfg)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if result.TotalCount != 4 {
		t.Errorf("TotalCount = %d, want 4", result.TotalCount)
	}
	if result.FoundCount != 3 {
		t.Errorf("FoundCount = %d, want 3", result.FoundCount)
	}

	itemMap := make(map[string]DiscoveredItem)
	for _, item := range result.Items {
		itemMap[item.ID] = item
	}

	obs := itemMap["OBS Studio:scenes & profiles"]
	if !obs.Found {
		t.Error("OBS Studio should be found")
	}
	if obs.SourcePath == "" {
		t.Error("OBS Studio SourcePath should not be empty")
	}

	missing := itemMap["Missing App:not here"]
	if missing.Found {
		t.Error("Missing App should not be found")
	}
}

func TestScan_NoUsersDir(t *testing.T) {
	root := t.TempDir()
	cfg := &config.Config{
		Version:    "1.0",
		BackupRoot: "D:\\Backup",
		Apps: []config.App{
			{
				Name:     "Test",
				Category: "Test",
				Items: []config.Item{
					{Label: "item", Source: "test", Target: "%APPDATA%/test"},
				},
			},
		},
	}

	result, err := Scan(context.Background(), root, cfg)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if result.FoundCount != 0 {
		t.Errorf("FoundCount = %d, want 0", result.FoundCount)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", result.TotalCount)
	}
}

func TestScan_PicksBestProfile(t *testing.T) {
	root := t.TempDir()

	userA := filepath.Join(root, "Users", "ProfileA")
	os.MkdirAll(filepath.Join(userA, "AppData", "Roaming", "obs-studio", "basic"), 0755)

	userB := filepath.Join(root, "Users", "ProfileB")
	os.MkdirAll(filepath.Join(userB, "AppData", "Roaming", "obs-studio", "basic"), 0755)
	os.MkdirAll(filepath.Join(userB, ".ssh"), 0755)

	cfg := &config.Config{
		Version:    "1.0",
		BackupRoot: "D:\\Backup",
		Apps: []config.App{
			{Name: "OBS Studio", Category: "Streaming", Items: []config.Item{
				{Label: "scenes", Source: "obs", Target: "%APPDATA%/obs-studio/basic"},
			}},
			{Name: "SSH", Category: "Dev", Items: []config.Item{
				{Label: "keys", Source: "ssh", Target: "%USERPROFILE%/.ssh"},
			}},
		},
	}

	result, err := Scan(context.Background(), root, cfg)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}

	if result.FoundCount != 2 {
		t.Errorf("FoundCount = %d, want 2 (ProfileB should win)", result.FoundCount)
	}
}

func TestScan_FiltersSystemProfiles(t *testing.T) {
	root := t.TempDir()

	for _, name := range []string{"Default", "Public", "All Users", "Default User"} {
		os.MkdirAll(filepath.Join(root, "Users", name, "AppData", "Roaming", "obs-studio", "basic"), 0755)
	}

	cfg := &config.Config{
		Version:    "1.0",
		BackupRoot: "D:\\Backup",
		Apps: []config.App{
			{Name: "OBS", Category: "S", Items: []config.Item{
				{Label: "x", Source: "obs", Target: "%APPDATA%/obs-studio/basic"},
			}},
		},
	}

	result, err := Scan(context.Background(), root, cfg)
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if result.FoundCount != 0 {
		t.Errorf("FoundCount = %d, want 0 (system profiles should be skipped)", result.FoundCount)
	}
}
