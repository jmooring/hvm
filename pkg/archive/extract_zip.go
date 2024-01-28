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

package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// extractZip extracts a zip file (src) to the dst directory.
func extractZip(src string, dst string) error {
	zrc, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zrc.Close()

	for _, f := range zrc.File {
		if strings.Contains(f.Name, "..") {
			return fmt.Errorf("detected unsafe file in archive (zip slip)")
		}

		target := filepath.Join(dst, f.Name)

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(target, 0o777)
			if err != nil {
				return err
			}
			continue
		}

		err := os.MkdirAll(filepath.Dir(target), 0o777)
		if err != nil {
			return err
		}

		copyFileFromZip(f, target)
	}

	return nil
}

// copyFileFromZip copies a file within a zip archive to the target path.
func copyFileFromZip(z *zip.File, dst string) error {
	zrc, err := z.Open()
	if err != nil {
		return err
	}
	defer zrc.Close()

	df, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, 0o777)
	if err != nil {
		return err
	}
	defer func() {
		if err := df.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	_, err = io.Copy(df, zrc)
	if err != nil {
		return err
	}

	return nil
}
