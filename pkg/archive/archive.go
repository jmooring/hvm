/*
Copyright © 2023 Joe Mooring <joe.mooring@veriphor.com>

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

// Package archive implements routines for extracting archive files including
// gzipped tarballs and zip files.
package archive

import (
	"fmt"
	"os"
	"strings"
)

// Extract extracts an archive (src) to the dst directory. If rm is is true,
// removes src when complete. Supports gzipped tarballs and zip files.
func Extract(src, dst string, rm bool) error {
	switch {
	case strings.HasSuffix(strings.ToLower(src), ".tar.gz"):
		err := extractTarGZ(src, dst)
		if err != nil {
			return err
		}
	case strings.HasSuffix(strings.ToLower(src), ".zip"):
		err := extractZip(src, dst)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown archive format")
	}

	if rm {
		err := os.Remove(src)
		if err != nil {
			return err
		}
	}

	return nil
}
