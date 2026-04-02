/*
Copyright © 2026 Joe Mooring <joe.mooring@veriphor.com>

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

// Package dotfile provides operations on the application dot file.
//
// The dot file stores a build identifier of the form "version/edition"
// (e.g. "v0.160.0/extended"). Files written by older versions of hvm contain
// only a version string; Read migrates these automatically.
package dotfile

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"

	"github.com/jmooring/hvm/pkg/repository"
	"golang.org/x/mod/semver"
)

// editionSupportBoundary is the last Hugo version for which the extended
// edition is guaranteed to exist. Legacy version-only dot files for versions
// at or before this boundary are migrated to "version/extended"; later
// versions use the configured default edition.
const editionSupportBoundary = "v0.160.0"

// Manager handles operations on the dot file.
type Manager struct {
	filePath       string
	fileName       string
	appName        string
	defaultEdition string // used when migrating files written by older hvm versions
}

// NewManager creates a new dotfile manager. defaultEdition is used when
// migrating dot files written by older versions of hvm that contain only a
// version string.
func NewManager(filePath, fileName, appName, defaultEdition string) *Manager {
	return &Manager{
		filePath:       filePath,
		fileName:       fileName,
		appName:        appName,
		defaultEdition: defaultEdition,
	}
}

// Read reads the build identifier ("version/edition") from the dot file in the
// current directory, or returns an empty string if the file does not exist.
//
// If the file contains only a version (written by an older version of hvm),
// Read migrates it to the version/edition format, rewrites the file, and
// prints a warning to stderr.
func (m *Manager) Read() (string, error) {
	exists, err := fileExists(m.filePath)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", nil
	}

	f, err := os.Open(m.filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(f)
	if err != nil {
		return "", err
	}

	theFix := fmt.Sprintf("run \"%s use\" to select a version, or \"%s disable\" to remove the file", m.appName, m.appName)

	dotFileContent := strings.TrimSpace(buf.String())
	if dotFileContent == "" {
		return "", fmt.Errorf("the %s file in the current directory is empty: %s", m.fileName, theFix)
	}

	// New format: version/edition (e.g. "v0.160.0/extended").
	if strings.Contains(dotFileContent, "/") {
		parts := strings.SplitN(dotFileContent, "/", 2)
		if !semver.IsValid(parts[0]) || !slices.Contains(repository.ValidEditions, parts[1]) {
			return "", fmt.Errorf("the %s file in the current directory has an invalid format: %s", m.fileName, theFix)
		}
		return dotFileContent, nil
	}

	// Legacy format: version only. Migrate.
	if !semver.IsValid(dotFileContent) {
		return "", fmt.Errorf("the %s file in the current directory has an invalid format: %s", m.fileName, theFix)
	}

	var edition string
	if semver.Compare(dotFileContent, editionSupportBoundary) <= 0 {
		edition = "extended"
	} else {
		edition = m.defaultEdition
	}

	migrated := dotFileContent + "/" + edition
	fmt.Fprintf(os.Stderr, "Warning: %s file migrated from %q to %q\n", m.fileName, dotFileContent, migrated)
	if err := m.Write(migrated); err != nil {
		return "", fmt.Errorf("failed to migrate %s file: %w", m.fileName, err)
	}

	return migrated, nil
}

// Write writes the version to the dot file for the current directory.
func (m *Manager) Write(version string) error {
	err := os.WriteFile(m.filePath, []byte(version), 0o644)
	if err != nil {
		return err
	}
	return nil
}

// fileExists checks if a file exists.
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}
