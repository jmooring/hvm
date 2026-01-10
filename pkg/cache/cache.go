/*
Copyright © 2026 Joe Mooring <joe.mooring@veriphor.com>

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
	"io/fs"
	"os"
	"runtime"
	"strings"
)

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
		if !d.IsDir() && !strings.HasPrefix(path, excludeDir) {
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
