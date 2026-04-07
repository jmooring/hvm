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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jmooring/hvm/pkg/cache"
)

// errCorruptCache is returned by loadTagCache when the cache file exists but
// contains invalid JSON. It is distinct from read/permission errors so that
// callers can decide whether to delete the file.
var errCorruptCache = errors.New("corrupt tag cache")

type tagCache struct {
	Tags []string `json:"tags"` // newest first
}

// saveTagCache writes the tag list to the cache file.
func saveTagCache(cacheDirPath string, tags []string) error {
	data, err := json.Marshal(tagCache{Tags: tags})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cacheDirPath, cache.TagListFileName), data, 0o644)
}

// loadTagCache reads the tag list from the cache file.
// Returns nil slice (not an error) if the file does not exist.
func loadTagCache(cacheDirPath string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(cacheDirPath, cache.TagListFileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var tc tagCache
	if err := json.Unmarshal(data, &tc); err != nil {
		return nil, fmt.Errorf("%w: %w", errCorruptCache, err)
	}
	return tc.Tags, nil
}
