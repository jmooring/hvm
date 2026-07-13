/*
Copyright 2023 Veriphor LLC

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

// TestExists tests the Exists function.
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
		{"a", args{dPath}, true, false},
		{"b", args{fPath}, true, false},
		{"c", args{filepath.Join(dPath, "does-not-exist.txt")}, false, false},
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

// TestCopyFile tests the CopyFile function.
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
		{"a", args{srcIsDir, dst}, true},
		{"b", args{srcDoesNotExist, dst}, true},
		{"c", args{srcIsFile, dst}, false},
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

// TestCopyDirectoryContent tests the CopyDirectoryContent function.
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
		{"a", args{srcDoesNotExist, dst}, true},
		{"b", args{srcIsFile, dst}, true},
		{"c", args{srcIsDir, dst}, false},
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

// TestJoinWithConjunction tests the JoinWithConjunction function.
func TestJoinWithConjunction(t *testing.T) {
	t.Parallel()

	type args struct {
		items       []string
		conjunction string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"blank conjunction", args{[]string{"a"}, ""}, "", true},
		{"zero values", args{[]string{}, "or"}, "", false},
		{"one value", args{[]string{"a"}, "or"}, "a", false},
		{"two values", args{[]string{"a", "b"}, "or"}, "a or b", false},
		{"three values", args{[]string{"a", "b", "c"}, "or"}, "a, b, or c", false},
		{"four values", args{[]string{"a", "b", "c", "d"}, "and"}, "a, b, c, and d", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JoinWithConjunction(tt.args.items, tt.args.conjunction)
			if (err != nil) != tt.wantErr {
				t.Errorf("JoinWithConjunction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JoinWithConjunction() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestIsString tests the IsString function.
func TestIsString(t *testing.T) {
	type args struct {
		i any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"a", args{nil}, false},
		{"b", args{true}, false},
		{"c", args{false}, false},
		{"d", args{""}, true},
		{"e", args{"a"}, true},
		{"f", args{"1"}, true},
		{"g", args{1}, false},
		{"h", args{0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsString(tt.args.i); got != tt.want {
				t.Errorf("IsString() = %v, want %v", got, tt.want)
			}
		})
	}
}
