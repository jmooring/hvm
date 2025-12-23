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
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/jmooring/hvm/pkg/helpers"
)

func TestExtract(t *testing.T) {
	t.Parallel()

	const (
		pkgFileName   = "test.pkg"    // test extraction on darwin only
		tarGZFileName = "test.tar.gz" // test extraction on all operating systems
		zipFileName   = "test.zip"    // test extraction on all operating systems
	)

	var (
		osAll    = []string{"darwin", "linux", "windows"}
		osDarwin = []string{"darwin"}
		dstDir   = t.TempDir()
		srcDir   = t.TempDir()
	)

	err := helpers.CopyDirectoryContent("testdata", srcDir)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		src string
		dst string
		rm  bool
	}

	tests := []struct {
		name     string
		osMatrix []string
		args     args
		wantErr  bool
	}{
		{"pkgKeep", osDarwin, args{filepath.Join(srcDir, pkgFileName), filepath.Clean(dstDir), false}, false},
		{"pkgRemove", osDarwin, args{filepath.Join(srcDir, pkgFileName), filepath.Clean(dstDir), true}, false},
		{"tarGZKeep", osAll, args{filepath.Join(srcDir, tarGZFileName), filepath.Clean(dstDir), false}, false},
		{"tarGZRemove", osAll, args{filepath.Join(srcDir, tarGZFileName), filepath.Clean(dstDir), true}, false},
		{"zipKeep", osAll, args{filepath.Join(srcDir, zipFileName), filepath.Clean(dstDir), false}, false},
		{"zipRemove", osAll, args{filepath.Join(srcDir, zipFileName), filepath.Clean(dstDir), true}, false},
		{"unknownArchiveFormat", osAll, args{"unknown.archive.format", "", false}, true},
	}
	for _, tt := range tests {
		if !slices.Contains(tt.osMatrix, runtime.GOOS) {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := Extract(tt.args.src, tt.args.dst, tt.args.rm); (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

		// Check content of extracted files.
		// This is the archive structure. The extracted structure must match.
		// Each file contains its path name followed by a newline.
		//
		// d1/
		// ├── d2/
		// │   ├── f3.txt  <-- example: content = "d1/d2/f3.txt\n"
		// │   └── f4.txt
		// ├── f1.txt
		// └── f2.txt

		err = filepath.WalkDir(dstDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}

			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(f)
			if err != nil {
				return err
			}
			f.Close()

			actualContent := strings.TrimSpace(buf.String())
			expectedContent := filepath.ToSlash(strings.TrimPrefix(path, dstDir))
			if actualContent != expectedContent {
				t.Errorf("Extract() error: expected content = %s, actual content = %s", expectedContent, actualContent)
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		// Was the source file removed if the rm arg was true?
		if tt.args.rm {
			exists, err := helpers.Exists(tt.args.src)
			if err != nil {
				t.Fatal(err)
			}
			if exists {
				t.Errorf("Extract() error: %v was not deleted", tt.args.src)
			}
		}

		// Clear the destination directory before the next test.
		err = helpers.RemoveDirectoryContent(dstDir)
		if err != nil {
			t.Fatal(err)
		}
	}
}
