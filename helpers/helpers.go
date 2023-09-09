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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// IsDir returns true if the given path is an existing directory.
func IsDir(path string) (bool, error) {
	exists, fi, err := Exists(path)
	if err != nil {
		return false, err
	}
	if exists && fi.Mode().IsDir() {
		return true, nil
	}

	return false, nil
}

// IsFile returns true if the given path is an existing regular file.
func IsFile(path string) (bool, error) {
	exists, fi, err := Exists(path)
	if err != nil {
		return false, err
	}
	if exists && fi.Mode().IsRegular() {
		return true, nil
	}

	return false, nil
}

// Exists returns true if the given file or directory exists.
func Exists(path string) (bool, fs.FileInfo, error) {
	fi, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}

	return true, fi, nil
}

// IsEmpty returns true if the given file or directory is empty. It returns an
// error if the file or directory does not exist.
func IsEmpty(path string) (bool, error) {
	exists, fi, err := Exists(path)
	if err != nil {
		return true, err
	}
	if !exists {
		return true, fmt.Errorf("%q does not exist", path)
	}
	if fi.IsDir() {
		f, err := os.Open(path)
		if err != nil {
			return false, err
		}
		defer f.Close()

		list, err := f.Readdir(-1)
		if err != nil {
			return false, err
		}
		return len(list) == 0, nil
	}

	return fi.Size() == 0, nil
}

// CopyFile copies a file from the src to the dst.
func CopyFile(src string, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("unable to copy: %s is not a file", src)
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	err = os.MkdirAll(filepath.Dir(dst), 0777)
	if err != nil {
		return err
	}

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := d.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	os.Chmod(dst, fi.Mode())

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}

	return nil
}
