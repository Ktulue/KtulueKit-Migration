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
