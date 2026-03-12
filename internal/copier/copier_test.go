package copier_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Ktulue/KtulueKit-Migration/internal/copier"
)

func TestCopyPath_File(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	dst := filepath.Join(tmp, "dst", "src.txt")
	if err := os.WriteFile(src, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	n, err := copier.CopyPath(src, dst)
	if err != nil {
		t.Fatalf("CopyPath file: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes, got %d", n)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "hello" {
		t.Errorf("expected 'hello', got %q", string(got))
	}
}

func TestCopyPath_Dir(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "srcdir")
	_ = os.MkdirAll(srcDir, 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("aaa"), 0644) // 3 bytes
	_ = os.WriteFile(filepath.Join(srcDir, "b.txt"), []byte("bb"), 0644)  // 2 bytes
	dstDir := filepath.Join(tmp, "dstdir")

	n, err := copier.CopyPath(srcDir, dstDir)
	if err != nil {
		t.Fatalf("CopyPath dir: %v", err)
	}
	if n != 5 { // 3 + 2 bytes
		t.Errorf("expected 5 bytes, got %d", n)
	}
	if _, err := os.Stat(filepath.Join(dstDir, "a.txt")); err != nil {
		t.Error("a.txt not found in dst")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "b.txt")); err != nil {
		t.Error("b.txt not found in dst")
	}
}
