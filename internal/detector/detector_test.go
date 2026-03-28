package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPatternMap_AppDataRoaming(t *testing.T) {
	localProfile := t.TempDir()
	expectedDest := filepath.Join(localProfile, "AppData", "Roaming", "obs-studio", "basic")
	os.MkdirAll(expectedDest, 0755)

	sourcePath := filepath.Join("E:", "Users", "Josh", "AppData", "Roaming", "obs-studio", "basic")

	result, ok := PatternMap(sourcePath, localProfile)
	if !ok {
		t.Fatal("PatternMap should match AppData\\Roaming pattern")
	}
	if result != expectedDest {
		t.Errorf("got %q, want %q", result, expectedDest)
	}
}

func TestPatternMap_AppDataLocal(t *testing.T) {
	localProfile := t.TempDir()
	expectedDest := filepath.Join(localProfile, "AppData", "Local", "SomeApp")
	os.MkdirAll(expectedDest, 0755)

	sourcePath := filepath.Join("D:", "Users", "OldUser", "AppData", "Local", "SomeApp")

	result, ok := PatternMap(sourcePath, localProfile)
	if !ok {
		t.Fatal("PatternMap should match AppData\\Local pattern")
	}
	if result != expectedDest {
		t.Errorf("got %q, want %q", result, expectedDest)
	}
}

func TestPatternMap_AppDataLocalLow(t *testing.T) {
	localProfile := t.TempDir()
	expectedDest := filepath.Join(localProfile, "AppData", "LocalLow", "BLAMCAM Interactive", "LurkBait")
	os.MkdirAll(expectedDest, 0755)

	sourcePath := filepath.Join("E:", "Users", "Josh", "AppData", "LocalLow", "BLAMCAM Interactive", "LurkBait")

	result, ok := PatternMap(sourcePath, localProfile)
	if !ok {
		t.Fatal("PatternMap should match AppData\\LocalLow pattern")
	}
	if result != expectedDest {
		t.Errorf("got %q, want %q", result, expectedDest)
	}
}

func TestPatternMap_UserProfile(t *testing.T) {
	localProfile := t.TempDir()
	expectedDest := filepath.Join(localProfile, ".ssh")
	os.MkdirAll(expectedDest, 0755)

	sourcePath := filepath.Join("E:", "Users", "Josh", ".ssh")

	result, ok := PatternMap(sourcePath, localProfile)
	if !ok {
		t.Fatal("PatternMap should match UserProfile pattern")
	}
	if result != expectedDest {
		t.Errorf("got %q, want %q", result, expectedDest)
	}
}

func TestPatternMap_Documents(t *testing.T) {
	localProfile := t.TempDir()
	expectedDest := filepath.Join(localProfile, "Documents", "MyStuff")
	os.MkdirAll(expectedDest, 0755)

	sourcePath := filepath.Join("E:", "Users", "Josh", "Documents", "MyStuff")

	result, ok := PatternMap(sourcePath, localProfile)
	if !ok {
		t.Fatal("PatternMap should match Documents pattern")
	}
	if result != expectedDest {
		t.Errorf("got %q, want %q", result, expectedDest)
	}
}

func TestPatternMap_NoMatch(t *testing.T) {
	localProfile := t.TempDir()
	sourcePath := filepath.Join("E:", "Program Files", "SomeApp", "data")

	_, ok := PatternMap(sourcePath, localProfile)
	if ok {
		t.Error("PatternMap should not match paths outside Users\\<name>\\...")
	}
}

func TestPatternMap_DestNotExist(t *testing.T) {
	localProfile := t.TempDir()
	sourcePath := filepath.Join("E:", "Users", "Josh", "AppData", "Roaming", "NonExistentApp")

	result, ok := PatternMap(sourcePath, localProfile)
	if !ok {
		t.Fatal("PatternMap should still match the pattern even if dest doesn't exist")
	}
	expectedDest := filepath.Join(localProfile, "AppData", "Roaming", "NonExistentApp")
	if result != expectedDest {
		t.Errorf("got %q, want %q", result, expectedDest)
	}
}

func TestSearchForApp_FindsTarget(t *testing.T) {
	root := t.TempDir()
	progFiles := filepath.Join(root, "Program Files")
	appDir := filepath.Join(progFiles, "LurkBait")
	os.MkdirAll(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "LurkBait.exe"), []byte("exe"), 0644)

	candidates := SearchForAppInRoots([]string{progFiles}, "LurkBait", "LurkBait.exe")
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0] != appDir {
		t.Errorf("got %q, want %q", candidates[0], appDir)
	}
}

func TestSearchForApp_NoExeConfirmation(t *testing.T) {
	root := t.TempDir()
	progFiles := filepath.Join(root, "Program Files")
	appDir := filepath.Join(progFiles, "SomeApp")
	os.MkdirAll(appDir, 0755)

	candidates := SearchForAppInRoots([]string{progFiles}, "SomeApp", "")
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0] != appDir {
		t.Errorf("got %q, want %q", candidates[0], appDir)
	}
}

func TestSearchForApp_ExeMismatch(t *testing.T) {
	root := t.TempDir()
	progFiles := filepath.Join(root, "Program Files")
	appDir := filepath.Join(progFiles, "LurkBait")
	os.MkdirAll(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "WrongApp.exe"), []byte("exe"), 0644)

	candidates := SearchForAppInRoots([]string{progFiles}, "LurkBait", "LurkBait.exe")
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates (exe mismatch), got %d", len(candidates))
	}
}

func TestSearchForApp_MultipleCandidates(t *testing.T) {
	root := t.TempDir()
	dir1 := filepath.Join(root, "ProgramFiles")
	dir2 := filepath.Join(root, "SteamLibrary")
	os.MkdirAll(filepath.Join(dir1, "LurkBait"), 0755)
	os.MkdirAll(filepath.Join(dir2, "LurkBait"), 0755)

	candidates := SearchForAppInRoots([]string{dir1, dir2}, "LurkBait", "")
	if len(candidates) != 2 {
		t.Errorf("expected 2 candidates, got %d", len(candidates))
	}
}

func TestSearchForApp_NotFound(t *testing.T) {
	root := t.TempDir()
	progFiles := filepath.Join(root, "Program Files")
	os.MkdirAll(progFiles, 0755)

	candidates := SearchForAppInRoots([]string{progFiles}, "NonExistent", "")
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates, got %d", len(candidates))
	}
}

func TestDetect_Tier1Wins(t *testing.T) {
	localProfile := t.TempDir()
	expectedDest := filepath.Join(localProfile, "AppData", "Roaming", "obs-studio", "basic")
	os.MkdirAll(expectedDest, 0755)

	sourcePaths := map[string]string{
		"OBS Studio:scenes & profiles": filepath.Join("E:", "Users", "Josh", "AppData", "Roaming", "obs-studio", "basic"),
	}

	results := Detect("OBS Studio", sourcePaths, nil, localProfile)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Method != "path-mapping" {
		t.Errorf("Method = %q, want path-mapping", r.Method)
	}
	if r.DestPath != expectedDest {
		t.Errorf("DestPath = %q, want %q", r.DestPath, expectedDest)
	}
	if !r.Confirmed {
		t.Error("expected Confirmed = true")
	}
}

func TestDetect_Tier1Unconfirmed(t *testing.T) {
	localProfile := t.TempDir()

	sourcePaths := map[string]string{
		"TestApp:data": filepath.Join("E:", "Users", "Josh", "AppData", "Roaming", "NonExistentApp"),
	}

	results := Detect("TestApp", sourcePaths, nil, localProfile)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Method != "path-mapping" {
		t.Errorf("Method = %q, want path-mapping", r.Method)
	}
	if r.Confirmed {
		t.Error("expected Confirmed = false (dest doesn't exist)")
	}
}

func TestDetect_NothingFound(t *testing.T) {
	localProfile := t.TempDir()

	sourcePaths := map[string]string{
		"UnknownApp:data": filepath.Join("E:", "RandomPath", "data"),
	}

	results := Detect("UnknownApp", sourcePaths, nil, localProfile)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.DestPath != "" {
		t.Errorf("expected empty DestPath, got %q", r.DestPath)
	}
	if r.Confirmed {
		t.Error("expected Confirmed = false")
	}
}
