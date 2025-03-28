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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/jmooring/hvm/pkg/helpers"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [version] | [flags]",
	Short: "Install a default version to use when version management is disabled",
	Long: `Displays a list of recent Hugo releases, prompting you to select a version
to use when version management is disabled in the current directory. It then
downloads, extracts, and caches the release asset for your operating system and
architecture, placing a copy in the cache "default" directory.

To use this version when version management is disabled in the current
directory, the cache "default" directory must be in your PATH. If it is not,
you will be prompted to add it when installation is complete.

Bypass the selection menu by specifying a version, or specify "latest" to use
the latest release.
`,
	Run: func(cmd *cobra.Command, args []string) {
		version := ""
		if len(args) > 0 {
			version = args[0]
		}
		err := install(version)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

// install sets the version of the Hugo executable to use when version
// management is disabled in the current directory.
func install(version string) error {
	asset := newAsset()
	repo := newRepository()

	if version == "" {
		msg := "Select a version to use when version management is disabled"
		err := repo.selectTag(asset, msg)
		if err != nil {
			return err
		}
		if asset.tag == "" {
			return nil // the user did not select a tag; do nothing
		}
	} else {
		err := repo.getTagFromString(asset, version)
		if err != nil {
			return err
		}
	}

	exists, err := helpers.Exists(filepath.Join(App.CacheDirPath, asset.tag))
	if err != nil {
		return err
	}

	if !exists {
		err := get(asset, repo)
		if err != nil {
			return err
		}
	}

	err = helpers.CopyFile(asset.getExecPath(), filepath.Join(App.CacheDirPath, App.DefaultDirName, asset.execName))
	if err != nil {
		return err
	}

	fmt.Println("Installation of", asset.tag, "complete.")

	if !slices.Contains(filepath.SplitList(os.Getenv("PATH")), App.DefaultDirPath) {
		fmt.Println()
		fmt.Printf("Please add %s to the PATH environment variable.\n", App.DefaultDirPath)
		fmt.Println("Open a new terminal after making the change.")

	}

	return nil
}
