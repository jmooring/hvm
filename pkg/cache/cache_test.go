/*
Copyright 2026 Veriphor LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"encoding/json"
	"errors"
	"io/fs"
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

func writeSchema(t *testing.T, dir string, version int) {
	t.Helper()
	data, err := json.Marshal(Schema{SchemaVersion: version})
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, SchemaFileName), data, 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}
}

func readSchema(t *testing.T, dir string) Schema {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, SchemaFileName))
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	return s
}

func TestEnsureSchema_CurrentVersion(t *testing.T) {
	base := t.TempDir()
	writeSchema(t, base, SchemaVersion)
	// A versioned dir that must NOT be removed.
	if err := os.MkdirAll(filepath.Join(base, "v1.0.0"), 0o777); err != nil {
		t.Fatal(err)
	}

	n, err := EnsureSchema(base, "default")
	if err != nil {
		t.Fatalf("EnsureSchema() error: %v", err)
	}
	if n != 0 {
		t.Fatalf("EnsureSchema(): want 0 removed, got %d", n)
	}
	if _, err := os.Stat(filepath.Join(base, "v1.0.0")); errors.Is(err, fs.ErrNotExist) {
		t.Fatal("EnsureSchema(): v1.0.0 should not have been removed")
	}
}

func TestEnsureSchema_NoSchema_EmptyCache(t *testing.T) {
	base := t.TempDir()

	n, err := EnsureSchema(base, "default")
	if err != nil {
		t.Fatalf("EnsureSchema() error: %v", err)
	}
	if n != 0 {
		t.Fatalf("EnsureSchema(): want 0 removed, got %d", n)
	}
	if s := readSchema(t, base); s.SchemaVersion != SchemaVersion {
		t.Fatalf("schema version: want %d got %d", SchemaVersion, s.SchemaVersion)
	}
}

func TestEnsureSchema_NoSchema_OldDirs(t *testing.T) {
	base := t.TempDir()
	write(t, filepath.Join(base, "v0.153.0", "hugo"), 10)
	write(t, filepath.Join(base, "v0.159.0", "hugo"), 10)
	write(t, filepath.Join(base, "default", "hugo"), 10)

	n, err := EnsureSchema(base, "default")
	if err != nil {
		t.Fatalf("EnsureSchema() error: %v", err)
	}
	if n != 2 {
		t.Fatalf("EnsureSchema(): want 2 removed, got %d", n)
	}
	for _, name := range []string{"v0.153.0", "v0.159.0"} {
		if _, err := os.Stat(filepath.Join(base, name)); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("EnsureSchema(): %s should have been removed", name)
		}
	}
	if _, err := os.Stat(filepath.Join(base, "default")); errors.Is(err, fs.ErrNotExist) {
		t.Fatal("EnsureSchema(): default/ should have been preserved")
	}
	if s := readSchema(t, base); s.SchemaVersion != SchemaVersion {
		t.Fatalf("schema version: want %d got %d", SchemaVersion, s.SchemaVersion)
	}
}

func TestEnsureSchema_WrongVersion(t *testing.T) {
	base := t.TempDir()
	writeSchema(t, base, 0)
	write(t, filepath.Join(base, "v0.153.0", "hugo"), 10)

	n, err := EnsureSchema(base, "default")
	if err != nil {
		t.Fatalf("EnsureSchema() error: %v", err)
	}
	if n != 1 {
		t.Fatalf("EnsureSchema(): want 1 removed, got %d", n)
	}
	if _, err := os.Stat(filepath.Join(base, "v0.153.0")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("EnsureSchema(): v0.153.0 should have been removed")
	}
	if s := readSchema(t, base); s.SchemaVersion != SchemaVersion {
		t.Fatalf("schema version after migration: want %d got %d", SchemaVersion, s.SchemaVersion)
	}
}

func TestEnsureSchema_UnreadableSchemaFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test not applicable on Windows")
	}
	base := t.TempDir()
	schemaPath := filepath.Join(base, SchemaFileName)
	writeSchema(t, base, SchemaVersion)
	// Make the schema file unreadable.
	if err := os.Chmod(schemaPath, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(schemaPath, 0o644) })

	_, err := EnsureSchema(base, "default")
	if err == nil {
		t.Fatal("EnsureSchema(): expected error for unreadable schema file")
	}
}
