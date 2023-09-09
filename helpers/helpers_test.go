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
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

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
		{"C", args{filepath.Join(dPath, "does-not-exist.txt")}, false, false},
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
		{"C", args{filepath.Join(dPath, "does-not-exist.txt")}, false, false},
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

	dPathFi, _ := os.Stat(dPath)
	fPathFi, _ := os.Stat(fPath)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		want1   fs.FileInfo
		wantErr bool
	}{
		{"A", args{dPath}, true, dPathFi, false},
		{"B", args{fPath}, true, fPathFi, false},
		{"C", args{filepath.Join(dPath, "does-not-exist.txt")}, false, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Exists(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Exists() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Exists() got1 = %v, want %v", got1, tt.want1)
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
	os.WriteFile(fPathNotEmpty, []byte("not empty"), 0644)

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
		{"E", args{filepath.Join(t.TempDir(), "does-not-exist.txt")}, true, true},
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

	dPath := t.TempDir()
	fPath := filepath.Join(dPath, "source.txt")

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
		src string
		dst string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"A", args{fPath, filepath.Join(dPath, "destination.txt")}, false},
		{"B", args{dPath, filepath.Join(dPath, "destination.txt")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CopyFile(tt.args.src, tt.args.dst); (err != nil) != tt.wantErr {
				t.Errorf("CopyFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
