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
package helpers

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsFile(t *testing.T) {
	t.Parallel()

	dPath := t.TempDir()
	fPath := filepath.Join(dPath, "empty-file.txt")

	f, err := os.Create(fPath)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"A", args{dPath}, false, false},
		{"B", args{fPath}, true, false},
		{"C", args{filepath.Join(dPath, "does-not-exist.txt")}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDir(t *testing.T) {
	t.Parallel()

	dPath := t.TempDir()
	fPath := filepath.Join(dPath, "exists.txt")

	f, err := os.Create(fPath)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"A", args{dPath}, true, false},
		{"B", args{fPath}, false, false},
		{"C", args{filepath.Join(dPath, "does-not-exist.txt")}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsDir(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExists(t *testing.T) {
	t.Parallel()

	dPath := t.TempDir()
	fPath := filepath.Join(dPath, "empty-file.txt")

	f, err := os.Create(fPath)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"A", args{dPath}, true, false},
		{"B", args{fPath}, true, false},
		{"C", args{filepath.Join(dPath, "does-not-exist.txt")}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Exists(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Exists() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	dPathEmpty := t.TempDir()
	fPathEmpty := filepath.Join(t.TempDir(), "empty-file.txt")

	f, err := os.Create(fPathEmpty)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()

	dPathNotEmpty := t.TempDir()
	fPathNotEmpty := filepath.Join(dPathNotEmpty, "not-empty-file.txt")

	err = os.WriteFile(fPathNotEmpty, []byte("not empty"), 0644)
	if err != nil {
		t.Error(err)
	}

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"A", args{dPathEmpty}, true, false},
		{"B", args{fPathEmpty}, true, false},
		{"C", args{dPathNotEmpty}, false, false},
		{"D", args{fPathNotEmpty}, false, false},
		{"E", args{filepath.Join(t.TempDir(), "does-not-exist.txt")}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsEmpty(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsEmpty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	t.Parallel()

	srcIsDir := "testdata"
	srcIsFile := "testdata/f1.txt"
	srcDoesNotExist := "does-not-exist"

	dst := filepath.Join(t.TempDir(), "destination.txt")

	type args struct {
		src string
		dst string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"A", args{srcIsDir, dst}, true},
		{"B", args{srcDoesNotExist, dst}, true},
		{"C", args{srcIsFile, dst}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CopyFile(tt.args.src, tt.args.dst); (err != nil) != tt.wantErr {
				t.Errorf("CopyFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Test C generated a file; now compare content.
	fi, err := os.Open(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer fi.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(fi)
	if err != nil {
		t.Fatal(err)
	}

	want := "f1.txt"
	got := strings.TrimSpace(buf.String())
	if want != got {
		t.Errorf("file content is incorrect: want %q got %q", want, got)
	}
}

func TestCopyDirectoryContent(t *testing.T) {
	t.Parallel()

	srcIsDir := "testdata"
	srcIsFile := "testdata/f1.txt"
	srcDoesNotExist := "does-not-exist"

	dst := t.TempDir()

	type args struct {
		src string
		dst string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"A", args{srcDoesNotExist, dst}, true},
		{"B", args{srcIsFile, dst}, true},
		{"C", args{srcIsDir, dst}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CopyDirectoryContent(tt.args.src, tt.args.dst); (err != nil) != tt.wantErr {
				t.Errorf("CopyDirectoryContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Test C copied a directory; now compare content. Expected:
	//
	// dst/
	// ├── d1/
	// │   ├── f3.txt  <-- example: content = "d1/f3.txt\n"
	// │   └── f4.txt
	// ├── f1.txt
	// └── f2.txt

	fi, err := os.Open(filepath.Join(dst, "d1", "f3.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer fi.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(fi)
	if err != nil {
		t.Fatal(err)
	}

	want := "d1/f3.txt"
	got := strings.TrimSpace(buf.String())
	if want != got {
		t.Errorf("file content is incorrect: want %q got %q", want, got)
	}

}

func TestRemoveDirectoryContent(t *testing.T) {
	t.Parallel()

	doesNotExist := "does-not-exist"
	isFile := "testdata/f1.txt"

	// Create a directory and stuff it with content.
	dir := t.TempDir()
	err := CopyDirectoryContent("testdata", dir)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"A", args{doesNotExist}, true},
		{"B", args{isFile}, true},
		{"C", args{dir}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RemoveDirectoryContent(tt.args.dir); (err != nil) != tt.wantErr {
				t.Errorf("RemoveDirectoryContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Test C cleared the directory; verify.
	empty, err := IsEmpty(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !empty {
		t.Error("the directory was not emptied")
	}
}
