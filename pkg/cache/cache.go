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

// Package cache provides cache-related operations.
package cache

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SchemaVersion is the current cache directory schema version.
const SchemaVersion = 1

// SchemaFileName is the name of the cache schema version file stored at the
// root of the cache directory.
const SchemaFileName = "schema.json"

// TagListFileName is the name of the cached release tag list file stored at
// the root of the cache directory.
const TagListFileName = "releases.json"

// Schema represents the cache schema version file stored at the root of the
// cache directory.
type Schema struct {
	SchemaVersion int `json:"schemaVersion"`
}

// EnsureSchema verifies that the cache schema file exists and matches the
// current schema version. If not, all versioned cache directories (except
// defaultDirName) are removed and a new schema file is written.
// Returns the number of directories removed; 0 means the schema was already
// current or the cache contained no versioned directories.
func EnsureSchema(cacheDirPath, defaultDirName string) (int, error) {
	schemaPath := filepath.Join(cacheDirPath, SchemaFileName)
	data, err := os.ReadFile(schemaPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return 0, err
	}
	if err == nil {
		var s Schema
		if json.Unmarshal(data, &s) == nil && s.SchemaVersion == SchemaVersion {
			return 0, nil
		}
	}

	// Schema absent or version mismatch: remove versioned cache directories.
	entries, err := os.ReadDir(cacheDirPath)
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, e := range entries {
		if !e.IsDir() || e.Name() == defaultDirName {
			continue
		}
		if err := os.RemoveAll(filepath.Join(cacheDirPath, e.Name())); err != nil {
			return removed, err
		}
		removed++
	}

	data, err = json.Marshal(Schema{SchemaVersion: SchemaVersion})
	if err != nil {
		return removed, err
	}
	if err := os.WriteFile(schemaPath, data, 0o644); err != nil {
		return removed, err
	}

	return removed, nil
}

// ExecName returns the name of the executable file based on the OS.
func ExecName() string {
	if runtime.GOOS == "windows" {
		return "hugo.exe"
	}
	return "hugo"
}

// Size returns the size of the cache directory, in bytes, excluding
// the specified exclude directory.
func Size(cachePath, excludeDir string) (int64, error) {
	var size int64 = 0
	err := fs.WalkDir(os.DirFS(cachePath), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && !strings.HasPrefix(path, excludeDir) && path != SchemaFileName && path != TagListFileName {
			fi, err := d.Info()
			if err != nil {
				return err
			}
			size += fi.Size()
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return size, nil
}
