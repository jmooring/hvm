package dotfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager("/tmp/.hvm", ".hvm", "hvm", "standard")
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestRead_FileDoesNotExist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hvm")
	m := NewManager(path, ".hvm", "hvm", "standard")
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
	m := NewManager(path, ".hvm", "hvm", "standard")
	_, err := m.Read()
	if err == nil {
		t.Fatal("Read() expected error for empty file")
	}
	if want := "empty"; !strings.Contains(err.Error(), want) {
		t.Fatalf("Read() error: want to contain %q got %q", want, err.Error())
	}
	if want := "run \"hvm use\""; !strings.Contains(err.Error(), want) {
		t.Fatalf("Read() fix message missing: want to contain %q got %q", want, err.Error())
	}
}

func TestRead_InvalidFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hvm")
	if err := os.WriteFile(path, []byte("1.2.3"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	m := NewManager(path, ".hvm", "hvm", "standard")
	_, err := m.Read()
	if err == nil {
		t.Fatal("Read() expected error for invalid format")
	}
	if want := "invalid format"; !strings.Contains(err.Error(), want) {
		t.Fatalf("Read() error: want to contain %q got %q", want, err.Error())
	}
}

func TestRead_EmptyEdition(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hvm")
	if err := os.WriteFile(path, []byte("v1.2.3/"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	m := NewManager(path, ".hvm", "hvm", "standard")
	_, err := m.Read()
	if err == nil {
		t.Fatal("Read() expected error for empty edition")
	}
	if want := "invalid format"; !strings.Contains(err.Error(), want) {
		t.Fatalf("Read() error: want to contain %q got %q", want, err.Error())
	}
}

// TestRead_LegacyVersionMigrated verifies that a version-only file (written by
// an older hvm) is automatically migrated to version/edition format.
func TestRead_LegacyVersionMigrated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hvm")
	// v0.100.0 is <= editionSupportBoundary (v0.160.0), so must migrate to extended.
	if err := os.WriteFile(path, []byte("v0.100.0"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	m := NewManager(path, ".hvm", "hvm", "standard")
	got, err := m.Read()
	if err != nil {
		t.Fatalf("Read() unexpected error: %v", err)
	}
	if want := "v0.100.0/extended"; got != want {
		t.Fatalf("Read(): want %q got %q", want, got)
	}
	// File on disk must also be updated.
	contents, _ := os.ReadFile(path)
	if string(contents) != "v0.100.0/extended" {
		t.Fatalf("file not updated: got %q", string(contents))
	}
}

// TestRead_LegacyVersionAboveBoundaryMigrated verifies that a version-only file
// above the edition support boundary is migrated using defaultEdition.
func TestRead_LegacyVersionAboveBoundaryMigrated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hvm")
	// v0.200.0 is > editionSupportBoundary, so uses defaultEdition.
	if err := os.WriteFile(path, []byte("v0.200.0"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	m := NewManager(path, ".hvm", "hvm", "extended_withdeploy")
	got, err := m.Read()
	if err != nil {
		t.Fatalf("Read() unexpected error: %v", err)
	}
	if want := "v0.200.0/extended_withdeploy"; got != want {
		t.Fatalf("Read(): want %q got %q", want, got)
	}
}

func TestRead_ValidBuildID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hvm")
	if err := os.WriteFile(path, []byte("  v1.2.3/extended  \n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	m := NewManager(path, ".hvm", "hvm", "standard")
	got, err := m.Read()
	if err != nil {
		t.Fatalf("Read() unexpected error: %v", err)
	}
	if want := "v1.2.3/extended"; got != want {
		t.Fatalf("Read(): want %q got %q", want, got)
	}
}

func TestWriteAndRead_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".hvm")
	m := NewManager(path, ".hvm", "hvm", "standard")
	if err := m.Write("v0.54.0/extended"); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	got, err := m.Read()
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if want := "v0.54.0/extended"; got != want {
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
