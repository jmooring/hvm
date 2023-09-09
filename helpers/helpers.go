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
	"strings"
)

// IsFile returns true if path is regular file, and returns an error if path
// does not exist.
func IsFile(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fi.Mode().IsRegular() {
		return true, nil
	}

	return false, nil
}

// IsDir returns true if path is a directory, and returns an error if path
// does not exist.
func IsDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fi.Mode().IsDir() {
		return true, nil
	}

	return false, nil
}

// Exists returns true if path exists, and false if it does not.
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

// IsEmpty returns true if the given file or directory is empty, and returns an
// error if path does not exist.
func IsEmpty(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	if fi.IsDir() {
		f, err := os.Open(path)
		if err != nil {
			return false, err
		}
		defer f.Close()

		list, err := f.Readdir(0)
		if err != nil {
			return false, err
		}
		return len(list) == 0, nil
	}

	return fi.Size() == 0, nil
}

// CopyFile copies a file from the src path to the dst path, overwriting an
// existing file if present.
func CopyFile(src string, dst string) error {
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

// CopyDirectoryContent copies the content of the src directory to the dst
// directory. Returns an error if the src directory does not exist.
func CopyDirectoryContent(src string, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", src)
	}

	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if path == src {
			return nil
		}

		segments := strings.Split(path, string(os.PathSeparator))
		target := filepath.Join(dst, filepath.Join(segments[1:]...))

		if d.IsDir() {
			err := os.MkdirAll(target, 0777)
			if err != nil {
				return err
			}
			return nil
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(target)
		if err != nil {
			return err
		}
		defer func() {
			if err := dst.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// RemoveDirectoryContent removes all content from dir, preserving the
// directory itself. Returns an error if dir does not exist, or if dir is not a
// directory.
func RemoveDirectoryContent(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if path != dir {
			err := os.RemoveAll(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
