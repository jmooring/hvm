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
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// extractTarGZ extracts a gzipped tarball (src) to the dst directory.
func extractTarGZ(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		th, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case th == nil:
			continue
		}

		if strings.Contains(th.Name, "..") {
			return fmt.Errorf("detected unsafe file in archive (zip slip)")
		}

		target := filepath.Join(dst, th.Name)

		switch th.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0o777); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			err = copyFileFromTarGZ(target, th, tr)
			if err != nil {
				return err
			}
		}
	}
}

// copyFileFromTarGZ copies a file within a tar.gz archive to the target path.
func copyFileFromTarGZ(dst string, th *tar.Header, tr *tar.Reader) error {
	df, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, os.FileMode(th.Mode))
	if err != nil {
		return err
	}
	defer func() {
		if err := df.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if _, err := io.Copy(df, tr); err != nil {
		return err
	}

	return nil
}
