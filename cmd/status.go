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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jmooring/hvm/pkg/cache"
	"github.com/jmooring/hvm/pkg/dotfile"
	"github.com/jmooring/hvm/pkg/helpers"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

// statusCmd represents the status command.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the status",
	Long: `Display a list of cached assets, the size of the cache, and the cache
location. The "default" directory created by the "install" command is excluded.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := status(cmd)
		cobra.CheckErr(err)
	},
}

// init registers the status command with the root command.
func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("printExecPath", false, `Based on the version specified in the `+app.DotFileName+` file,
print the path to Hugo executable, otherwise
return exit code 1. This path is not verified, and the
executable may not exist.`)
	statusCmd.Flags().Bool("printExecPathCached", false, `Based on the version specified in the `+app.DotFileName+` file,
print the path to Hugo executable if cached, otherwise
return exit code 1.`)
	statusCmd.MarkFlagsMutuallyExclusive("printExecPath", "printExecPathCached")
}

// status displays a list of cached assets, the size of the cache, and the
// cache location. The "default" directory created by the "install" command is
// excluded.
//
//gocyclo:ignore
func status(cmd *cobra.Command) error {
	printExecPath, err := cmd.Flags().GetBool("printExecPath")
	if err != nil {
		return err
	}

	printExecPathCached, err := cmd.Flags().GetBool("printExecPathCached")
	if err != nil {
		return err
	}

	dm := dotfile.NewManager(app.DotFilePath, app.DotFileName, app.Name)
	version, err := dm.Read()
	if err != nil {
		return err
	}

	execPath := filepath.Join(app.CacheDirPath, version, cache.ExecName())
	execPathExists, err := helpers.Exists(execPath)
	if err != nil {
		return err
	}

	// Prints exec path, else exit code 1.
	if printExecPath {
		if version == "" {
			os.Exit(1)
		} else {
			fmt.Println(execPath)
			os.Exit(0)
		}
	}

	// Prints the exec path if the exec is cached, else exit code 1.
	if printExecPathCached {
		if version == "" {
			os.Exit(1)
		} else {
			if execPathExists {
				fmt.Println(execPath)
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		}
	}

	if version == "" {
		fmt.Println("Version management is disabled in the current directory.")
	} else {
		if execPathExists {
			fmt.Printf("The current directory is configured to use Hugo %s.\n", version)
		} else {
			var r string
			fmt.Printf("The %s file in the current directory refers to a Hugo\n", app.DotFileName)
			fmt.Printf("version (%s) that is not cached.\n", version)
			fmt.Println()
			for {
				fmt.Printf("Would you like to get it now? (Y/n): ")
				fmt.Scanln(&r)
				if r == "" || strings.EqualFold(string(r[0]), "y") {
					err = use(version)
					if err != nil {
						theFix := fmt.Sprintf("run \"%[1]s use\" to select a version, or \"%[1]s disable\" to remove the file", app.Name)
						return fmt.Errorf("the version specified in the %s file (%s) is not available in the repository: %s", app.DotFileName, version, theFix)
					}
					fmt.Println()
					break
				}
				if strings.EqualFold(string(r[0]), "n") {
					err = disable()
					if err != nil {
						return err
					}
					fmt.Println()
					break
				}
			}
		}
	}

	// Get tags; ignore app.DefaultDirName.
	sd, err := os.ReadDir(app.CacheDirPath)
	if err != nil {
		return err
	}

	var tags []string
	for _, d := range sd {
		if d.IsDir() && d.Name() != app.DefaultDirName {
			tags = append(tags, d.Name())
		}
	}

	if len(tags) == 0 {
		fmt.Println("The cache is empty.")
		return nil
	}

	fmt.Println("Cached versions of the Hugo executable:")
	fmt.Println()
	semver.Sort(tags)
	if !config.SortAscending {
		slices.Reverse(tags)
	}
	for _, tag := range tags {
		fmt.Println(tag)
	}
	fmt.Println()

	size, err := cache.Size(app.CacheDirPath, app.DefaultDirName)
	if err != nil {
		return err
	}
	fmt.Println("Cache size:", size/1000000, "MB")
	fmt.Println("Cache directory:", app.CacheDirPath)

	return nil
}

// readBytes reads and trims whitespace from a bytes buffer.
func readBytes(b *bytes.Buffer) (string, error) {
	content := strings.TrimSpace(b.String())
	return content, nil
}
