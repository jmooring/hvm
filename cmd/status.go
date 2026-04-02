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

	dm := dotfile.NewManager(app.DotFilePath, app.DotFileName, app.Name, config.DefaultEdition)
	buildID, err := dm.Read()
	if err != nil {
		return err
	}

	// buildID is "version/edition" (e.g. "v0.160.0/extended") or empty.
	var version, edition string
	if buildID != "" {
		parts := strings.SplitN(buildID, "/", 2)
		version = parts[0]
		if len(parts) == 2 {
			edition = parts[1]
		}
	}

	execPath := filepath.Join(app.CacheDirPath, version, edition, cache.ExecName())
	execPathExists, err := helpers.Exists(execPath)
	if err != nil {
		return err
	}

	// Prints exec path, else exit code 1.
	if printExecPath {
		if buildID == "" {
			os.Exit(1)
		} else {
			fmt.Println(execPath)
			os.Exit(0)
		}
	}

	// Prints the exec path if the exec is cached, else exit code 1.
	if printExecPathCached {
		if buildID == "" {
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

	if buildID == "" {
		fmt.Println("Version management is disabled for the current directory.")
	} else {
		if execPathExists {
			fmt.Printf("The current directory is configured to use Hugo %s.\n", buildID)
		} else {
			fmt.Printf("The %s file in the current directory refers to a Hugo\n", app.DotFileName)
			fmt.Printf("version (%s) that is not cached.\n", buildID)
			fmt.Println()
			if promptYesNo("Would you like to get it now?", true) {
				err = use(buildID)
				if err != nil {
					theFix := fmt.Sprintf("run \"%[1]s use\" to select a version, or \"%[1]s disable\" to remove the file", app.Name)
					return fmt.Errorf("unable to get %s (%s): %w: %s", app.DotFileName, buildID, err, theFix)
				}
			} else {
				err = disable()
				if err != nil {
					return err
				}
			}
			fmt.Println()
		}
	}

	// Get tag/edition entries; ignore app.DefaultDirName.
	// Each entry is a tag directory whose subdirectories are editions.
	sd, err := os.ReadDir(app.CacheDirPath)
	if err != nil {
		return err
	}

	var buildIDs []string
	for _, d := range sd {
		if !d.IsDir() || d.Name() == app.DefaultDirName {
			continue
		}
		tag := d.Name()
		// List edition subdirectories within this tag directory.
		editionDirs, err := os.ReadDir(filepath.Join(app.CacheDirPath, tag))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not read cache directory %s: %s\n", tag, err)
			continue
		}
		for _, ed := range editionDirs {
			if ed.IsDir() {
				buildIDs = append(buildIDs, tag+"/"+ed.Name())
			}
		}
	}

	if len(buildIDs) == 0 {
		fmt.Println("The cache is empty.")
		return nil
	}

	fmt.Println("Cached versions of the Hugo executable:")
	fmt.Println()
	slices.SortStableFunc(buildIDs, func(a, b string) int {
		aTag, aEd, _ := strings.Cut(a, "/")
		bTag, bEd, _ := strings.Cut(b, "/")
		if c := semver.Compare(aTag, bTag); c != 0 {
			return c
		}
		return strings.Compare(aEd, bEd)
	})
	if !config.SortAscending {
		slices.Reverse(buildIDs)
	}
	for _, id := range buildIDs {
		fmt.Println(id)
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
