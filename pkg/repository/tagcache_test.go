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

package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmooring/hvm/pkg/cache"
)

func TestSaveAndLoadTagCache(t *testing.T) {
	dir := t.TempDir()
	tags := []string{"v0.153.0", "v0.152.0", "v0.151.0"}

	if err := saveTagCache(dir, tags); err != nil {
		t.Fatalf("saveTagCache error: %v", err)
	}

	got, err := loadTagCache(dir)
	if err != nil {
		t.Fatalf("loadTagCache error: %v", err)
	}

	if len(got) != len(tags) {
		t.Fatalf("loadTagCache: want %d tags got %d", len(tags), len(got))
	}
	for i, tag := range tags {
		if got[i] != tag {
			t.Errorf("loadTagCache[%d]: want %q got %q", i, tag, got[i])
		}
	}
}

func TestLoadTagCache_NotExist(t *testing.T) {
	dir := t.TempDir()
	got, err := loadTagCache(dir)
	if err != nil {
		t.Fatalf("loadTagCache on missing file should not error: %v", err)
	}
	if got != nil {
		t.Fatalf("loadTagCache on missing file: want nil got %v", got)
	}
}

func TestLoadTagCache_Corrupt(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, cache.TagListFileName), []byte("not json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := loadTagCache(dir)
	if err == nil {
		t.Fatal("loadTagCache on corrupt file should error")
	}
}
