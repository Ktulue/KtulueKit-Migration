package reporter_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ktulue/KtulueKit-Migration/internal/reporter"
)

func TestNewNull_NoFileCreated(t *testing.T) {
	tmp := t.TempDir()
	r := reporter.NewNull()
	// All methods must not panic and must not create files
	r.Add(reporter.Result{App: "Test", Label: "item", Status: reporter.StatusCopied})
	r.Summary()
	r.Close()

	entries, _ := os.ReadDir(tmp)
	if len(entries) != 0 {
		t.Error("NewNull should not create any files")
	}
	if r.LogPath() != "" {
		t.Errorf("expected empty log path, got %q", r.LogPath())
	}
}

func TestWriteManifest(t *testing.T) {
	tmp := t.TempDir()
	r := reporter.New(tmp)
	r.Add(reporter.Result{
		App: "OBS Studio", Label: "scenes",
		SourcePath: "D:/backup/obs", TargetPath: "C:/AppData/obs",
		Status: reporter.StatusCopied, BytesCopied: 1024,
	})
	r.Add(reporter.Result{
		App: "Personal Files", Label: "Documents",
		SourcePath: "D:/backup/docs", TargetPath: "C:/Users/docs",
		Status: reporter.StatusCopied, BytesCopied: 512,
		SelectedPaths: []string{"D:/backup/docs/Notes.txt"},
	})
	r.Close()

	manifestPath := filepath.Join(tmp, "manifest.json")
	if err := r.WriteManifest(manifestPath); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	var m struct {
		Version string `json:"version"`
		Items   []struct {
			App           string   `json:"app"`
			SelectedPaths []string `json:"selectedPaths"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m.Version != "1.0" {
		t.Errorf("expected version 1.0, got %q", m.Version)
	}
	if len(m.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(m.Items))
	}
	if len(m.Items[1].SelectedPaths) != 1 {
		t.Error("expected selectedPaths on Documents item")
	}
}
