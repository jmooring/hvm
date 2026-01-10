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
package dotfile

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"golang.org/x/mod/semver"
)

// Manager handles operations on the dot file.
type Manager struct {
	filePath string
	fileName string
	appName  string
}

// NewManager creates a new dotfile manager.
func NewManager(filePath, fileName, appName string) *Manager {
	return &Manager{
		filePath: filePath,
		fileName: fileName,
		appName:  appName,
	}
}

// Read reads the version from the dot file in the current directory, or
// returns an empty string if the file does not exist.
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

	if !semver.IsValid(dotFileContent) {
		return "", fmt.Errorf("the %s file in the current directory has an invalid format: %s", m.fileName, theFix)
	}

	return dotFileContent, nil
}

// Write writes the version to the dot file in the current directory.
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
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
