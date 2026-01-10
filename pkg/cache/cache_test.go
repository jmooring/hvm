package cache

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestExecName(t *testing.T) {
	want := "hugo"
	if runtime.GOOS == "windows" {
		want = "hugo.exe"
	}
	if got := ExecName(); got != want {
		t.Fatalf("ExecName(): want %q got %q", want, got)
	}
}

func TestSize(t *testing.T) {
	base := t.TempDir()

	// Files to include (not under exclude dir)
	write(t, filepath.Join(base, "v1", "hugo"), 20)
	write(t, filepath.Join(base, "v1", "extra.bin"), 5)
	write(t, filepath.Join(base, "rootfile"), 3)

	// Files to exclude (under exclude dir)
	write(t, filepath.Join(base, "default", "hugo"), 10)
	write(t, filepath.Join(base, "default", "nested", "file"), 7)

	// Path that starts with exclude prefix should be excluded
	write(t, filepath.Join(base, "default123", "file"), 9)

	size, err := Size(base, "default")
	if err != nil {
		t.Fatalf("Size() error: %v", err)
	}

	// Included: 20 + 5 + 3 = 28
	// Excluded: 10 + 7 + 9
	if want := int64(28); size != want {
		t.Fatalf("Size(): want %d got %d", want, size)
	}
}

func TestSize_EmptyDir(t *testing.T) {
	base := t.TempDir()
	size, err := Size(base, "default")
	if err != nil {
		t.Fatalf("Size() error: %v", err)
	}
	if size != 0 {
		t.Fatalf("Size(): want 0 got %d", size)
	}
}

func write(t *testing.T, path string, bytes int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	b := make([]byte, bytes)
	for i := range b {
		b[i] = byte(i%251 + 1)
	}
	if err := os.WriteFile(path, b, 0o666); err != nil {
		t.Fatalf("write: %v", err)
	}
}
