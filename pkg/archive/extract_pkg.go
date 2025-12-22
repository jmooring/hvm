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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/jmooring/hvm/pkg/helpers"
)

// extractPkg unpacks the contents of a macOS .pkg file (src) to the dst
// directory. It specifically extracts the files contained within the
// package's Payload.
func extractPkg(src, dst string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("extraction of pkg file is limited to darwin")
	}

	exists, err := helpers.Exists(src)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("unable to find %s", src)
	}

	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}

	expansionDir := filepath.Join(tempDir, "expanded")

	cmd := exec.Command("pkgutil", "--expand-full", src, expansionDir)
	co, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", co)
	}

	err = helpers.CopyDirectoryContent(filepath.Join(expansionDir, "Payload"), dst)
	if err != nil {
		return err
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		return err
	}

	return nil
}
