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

// Package helpers implements utility routines for working with the filesystem.
package helpers

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CopyDirectoryContent replaces dst with the contents of src, removing dst
// first if it exists and creating it fresh. Returns an error if src does not
// exist or if src is not a directory.
func CopyDirectoryContent(src, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", src)
	}

	err = os.RemoveAll(dst)
	if err != nil {
		return err
	}

	return os.CopyFS(dst, os.DirFS(src))
}

// CopyFile copies a file from src to dst, overwriting an existing file if
// present. Returns an error if src does not exist, or if src is a directory.
func CopyFile(src, dst string) (retErr error) {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("%s is not a file", src)
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	err = os.MkdirAll(filepath.Dir(dst), 0o777)
	if err != nil {
		return err
	}

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := d.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	if err := os.Chmod(dst, fi.Mode()); err != nil {
		return err
	}

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}

	return nil
}

// Exists reports whether path exists.
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// IsString reports whether i is a string.
func IsString(i any) bool {
	_, ok := i.(string)

	return ok
}

// JoinWithConjunction joins a slice of strings into a human-readable list,
// separating items with commas and prefixing the last item with conjunction
// (e.g., "and" or "or"). Returns an empty string for an empty slice, the
// single value for a one-element slice, and two values joined by the
// conjunction without commas. Returns an error if conjunction is blank.
func JoinWithConjunction(items []string, conjunction string) (string, error) {
	if conjunction == "" {
		return "", fmt.Errorf("conjunction must not be blank")
	}
	switch len(items) {
	case 0:
		return "", nil
	case 1:
		return items[0], nil
	case 2:
		return items[0] + " " + conjunction + " " + items[1], nil
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", " + conjunction + " " + items[len(items)-1], nil
	}
}
