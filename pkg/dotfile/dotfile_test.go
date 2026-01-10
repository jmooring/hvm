package dotfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager("/tmp/.hvm", ".hvm", "hvm")
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestRead_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hvm")
	m := NewManager(path, ".hvm", "hvm")
	got, err := m.Read()
	if err != nil {
		t.Fatalf("Read() unexpected error: %v", err)
	}
	if got != "" {
		t.Fatalf("Read(): want empty string got %q", got)
	}
}

func TestRead_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hvm")
	if err := os.WriteFile(path, []byte("\n\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	m := NewManager(path, ".hvm", "hvm")
	_, err := m.Read()
	if err == nil {
		t.Fatal("Read() expected error for empty file")
	}
	if want := "empty"; !contains(err.Error(), want) {
		t.Fatalf("Read() error: want to contain %q got %q", want, err.Error())
	}
	if want := "run \"hvm use\""; !contains(err.Error(), want) {
		t.Fatalf("Read() fix message missing: want to contain %q got %q", want, err.Error())
	}
}

func TestRead_InvalidFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hvm")
	if err := os.WriteFile(path, []byte("1.2.3"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	m := NewManager(path, ".hvm", "hvm")
	_, err := m.Read()
	if err == nil {
		t.Fatal("Read() expected error for invalid format")
	}
	if want := "invalid format"; !contains(err.Error(), want) {
		t.Fatalf("Read() error: want to contain %q got %q", want, err.Error())
	}
}

func TestRead_ValidSemver(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hvm")
	if err := os.WriteFile(path, []byte("  v1.2.3  \n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	m := NewManager(path, ".hvm", "hvm")
	got, err := m.Read()
	if err != nil {
		t.Fatalf("Read() unexpected error: %v", err)
	}
	if want := "v1.2.3"; got != want {
		t.Fatalf("Read(): want %q got %q", want, got)
	}
}

func TestWriteAndRead_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hvm")
	m := NewManager(path, ".hvm", "hvm")
	if err := m.Write("v0.54.0"); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	got, err := m.Read()
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if want := "v0.54.0"; got != want {
		t.Fatalf("round-trip: want %q got %q", want, got)
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	// not exist
	exists, err := fileExists(path)
	if err != nil {
		t.Fatalf("fileExists error: %v", err)
	}
	if exists {
		t.Fatal("fileExists: want false got true")
	}
	// create
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	exists, err = fileExists(path)
	if err != nil {
		t.Fatalf("fileExists error: %v", err)
	}
	if !exists {
		t.Fatal("fileExists: want true got false")
	}
}

// contains is a helper to check substring existence.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) == 0 || indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	// small inline search to avoid importing strings
	if len(sub) == 0 {
		return 0
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if s[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
