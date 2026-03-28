package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// writeTemp writes content to a temp file and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return filepath.ToSlash(f.Name())
}

const baseConfig = `{
  "version": "1.0",
  "metadata": { "name": "Test", "author": "Tester" },
  "backup_root": "D:\\Backup",
  "apps": []
}`

func TestLoad_WithDetection(t *testing.T) {
	const cfg = `{
  "version": "1.0",
  "metadata": { "name": "Test", "author": "Tester" },
  "backup_root": "D:\\Backup",
  "apps": [
    {
      "name": "LurkBait",
      "category": "Games",
      "detection": {
        "registry": "HKCU\\Software\\LurkBait\\InstallPath",
        "searchPaths": ["Program Files", "Program Files (x86)", "SteamLibrary/steamapps/common"],
        "searchTarget": "LurkBait",
        "executable": "LurkBait.exe"
      },
      "items": [
        { "label": "Save Data", "source": "LurkBait/saves", "target": "%APPDATA%/LurkBait/saves" }
      ]
    }
  ]
}`

	path := writeTemp(t, cfg)
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(loaded.Apps))
	}

	app := loaded.Apps[0]
	if app.Detection == nil {
		t.Fatal("expected Detection to be non-nil")
	}

	d := app.Detection
	if d.Registry != `HKCU\Software\LurkBait\InstallPath` {
		t.Errorf("Registry: got %q, want %q", d.Registry, `HKCU\Software\LurkBait\InstallPath`)
	}
	if len(d.SearchPaths) != 3 {
		t.Errorf("SearchPaths: got %d elements, want 3", len(d.SearchPaths))
	} else {
		if d.SearchPaths[0] != "Program Files" {
			t.Errorf("SearchPaths[0]: got %q, want %q", d.SearchPaths[0], "Program Files")
		}
		if d.SearchPaths[1] != "Program Files (x86)" {
			t.Errorf("SearchPaths[1]: got %q, want %q", d.SearchPaths[1], "Program Files (x86)")
		}
		if d.SearchPaths[2] != "SteamLibrary/steamapps/common" {
			t.Errorf("SearchPaths[2]: got %q, want %q", d.SearchPaths[2], "SteamLibrary/steamapps/common")
		}
	}
	if d.SearchTarget != "LurkBait" {
		t.Errorf("SearchTarget: got %q, want %q", d.SearchTarget, "LurkBait")
	}
	if d.Executable != "LurkBait.exe" {
		t.Errorf("Executable: got %q, want %q", d.Executable, "LurkBait.exe")
	}
}

func TestLoad_WithoutDetection(t *testing.T) {
	const cfg = `{
  "version": "1.0",
  "metadata": { "name": "Test", "author": "Tester" },
  "backup_root": "D:\\Backup",
  "apps": [
    {
      "name": "OBS Studio",
      "category": "Streaming",
      "items": [
        { "label": "Scenes", "source": "OBS/basic", "target": "%APPDATA%/obs-studio/basic" }
      ]
    }
  ]
}`

	path := writeTemp(t, cfg)
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(loaded.Apps))
	}

	if loaded.Apps[0].Detection != nil {
		t.Errorf("expected Detection to be nil, got %+v", loaded.Apps[0].Detection)
	}
}

// Ensure Detection round-trips through JSON cleanly (omitempty — nil should not appear in output).
func TestDetection_OmitEmptyOnMarshal(t *testing.T) {
	app := App{
		Name:     "NoDetect",
		Category: "Utilities",
		Items:    []Item{{Label: "cfg", Source: "src", Target: "tgt"}},
	}

	b, err := json.Marshal(app)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	if _, ok := m["detection"]; ok {
		t.Error("expected 'detection' key to be absent when Detection is nil (omitempty)")
	}
}
