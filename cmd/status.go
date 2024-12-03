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
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jmooring/hvm/pkg/helpers"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

// statusCmd represents the status command
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

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("printExecPath", false, `Based on the version specified in the `+App.DotFileName+` file,
print the path to Hugo executable, otherwise
return exit code 1. This path is not verified, and the
executable may not exist.`)
	statusCmd.Flags().Bool("printExecPathCached", false, `Based on the version specified in the `+App.DotFileName+` file,
print the path to Hugo executable if cached, otherwise
return exit code 1.`)
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

	version, err := getVersionFromDotFile(App.DotFilePath)
	if err != nil {
		return err
	}

	execPath := filepath.Join(App.CacheDirPath, version, getExecName())
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
			fmt.Printf("The %s file in the current directory refers to a Hugo\n", App.DotFileName)
			fmt.Printf("version (%s) that is not cached.\n", version)
			fmt.Println()
			for {
				fmt.Printf("Would you like to get it now? (Y/n): ")
				fmt.Scanln(&r)
				if len(r) == 0 || strings.ToLower(string(r[0])) == "y" {
					err = use(true, false, "")
					if err != nil {
						return err
					}
					fmt.Println()
					break
				}
				if strings.ToLower(string(r[0])) == "n" {
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

	// Get tags; ignore App.DefaultDirName.
	sd, err := os.ReadDir(App.CacheDirPath)
	if err != nil {
		return err
	}

	var tags []string
	for _, d := range sd {
		if d.IsDir() && d.Name() != App.DefaultDirName {
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
	if !Config.SortAscending {
		slices.Reverse(tags)
	}
	for _, tag := range tags {
		fmt.Println(tag)
	}
	fmt.Println()

	size, err := getCacheSize()
	if err != nil {
		return err
	}
	fmt.Println("Cache size:", size/1000000, "MB")
	fmt.Println("Cache directory:", App.CacheDirPath)

	return nil
}

// getCacheSize returns the size of the cache directory, in bytes, excluding
// the App.DefaultDirName subdirectory.
func getCacheSize() (int64, error) {
	var size int64 = 0
	err := fs.WalkDir(os.DirFS(App.CacheDirPath), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && !strings.HasPrefix(path, App.DefaultDirName) {
			fi, err := d.Info()
			if err != nil {
				return err
			}
			size = size + fi.Size()
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return size, nil
}

// getVersionFromDotFile returns the semver string from the app dot file in the
// current directory, or an empty string if the file does not exist.
func getVersionFromDotFile(path string) (string, error) {
	exists, err := helpers.Exists(path)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", nil
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(f)
	if err != nil {
		return "", err
	}

	theFix := fmt.Sprintf("run \"%[1]s use\" to select a version, or \"%[1]s disable\" to remove the file", App.Name)

	dotHvmContent := strings.TrimSpace(buf.String())
	if dotHvmContent == "" {
		return "", fmt.Errorf("the %s file in the current directory is empty: %s", App.DotFileName, theFix)
	}

	if !semver.IsValid(dotHvmContent) {
		return "", fmt.Errorf("the %s file in the current directory has an invalid format: %s", App.DotFileName, theFix)
	}

	return (dotHvmContent), nil
}
