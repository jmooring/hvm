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
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jmooring/hvm/pkg/archive"
	"github.com/jmooring/hvm/pkg/cache"
	"github.com/jmooring/hvm/pkg/dotfile"
	gh "github.com/jmooring/hvm/pkg/github"
	"github.com/jmooring/hvm/pkg/helpers"
	"github.com/jmooring/hvm/pkg/repository"
	"github.com/spf13/cobra"
)

// useCmd represents the use command.
var useCmd = &cobra.Command{
	Use:     "use [version] | [flags]",
	Aliases: []string{"get"},
	Short:   "Select or specify a version to use in the current directory",
	Long: `Displays a list of recent Hugo releases, prompting you to select a version
to use in the current directory. It then downloads, extracts, and caches the
release asset for your operating system and architecture and writes the version
tag to an ` + app.DotFileName + ` file.

Bypass the selection menu by specifying a version, or specify "latest" to use
the latest release.
`,
	Run: func(cmd *cobra.Command, args []string) {
		version := ""

		useVersionInDotFile, err := cmd.Flags().GetBool("useVersionInDotFile")
		cobra.CheckErr(err)

		if useVersionInDotFile {
			dm := dotfile.NewManager(app.DotFilePath, app.DotFileName, app.Name)
			version, err = dm.Read()
			cobra.CheckErr(err)
			if version == "" {
				cobra.CheckErr(fmt.Errorf("the current directory does not contain an %s file", app.DotFileName))
			}
		} else if len(args) > 0 {
			version = args[0]
		}

		err = use(version)
		cobra.CheckErr(err)
	},
}

// init registers the use command with the root command.
func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.Flags().Bool("useVersionInDotFile", false, "Use the version specified by the "+app.DotFileName+" file\nin the current directory")
}

// use sets the version of the Hugo executable to use in the current directory.
func use(version string) error {
	asset := repository.NewAsset(cache.ExecName())

	client := gh.NewClient(config.GitHubToken)
	repo, err := repository.NewRepository(app.ManagedApp.RepositoryOwner, app.ManagedApp.RepositoryName, client)
	if err != nil {
		return err
	}

	if version == "" {
		msg := "Select a version to use in the current directory"
		err := repo.SelectTag(asset, msg, config.SortAscending, config.NumTagsToDisplay, func(tag string) (bool, error) {
			return helpers.Exists(filepath.Join(app.CacheDirPath, tag))
		})
		if err != nil {
			return err
		}
		if asset.Tag == "" {
			return nil // the user did not select a tag; do nothing
		}
	} else {
		err := repo.GetTagFromString(asset, version)
		if err != nil {
			return err
		}
	}

	exists, err := helpers.Exists(filepath.Join(app.CacheDirPath, asset.Tag))
	if err != nil {
		return err
	}

	if exists {
		fmt.Printf("Using %s from cache.\n", asset.Tag)
	} else {
		err := downloadAndCache(asset, repo)
		if err != nil {
			return err
		}
	}

	dm := dotfile.NewManager(app.DotFilePath, app.DotFileName, app.Name)
	err = dm.Write(asset.Tag)
	if err != nil {
		return err
	}

	return nil
}

// downloadAndCache downloads and extracts the release asset.
func downloadAndCache(asset *repository.Asset, repo *repository.Repository) error {
	err := repo.FetchDownloadURL(asset)
	if err != nil {
		return err
	}

	asset.ArchiveDirPath, err = os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer func() {
		err := os.RemoveAll(asset.ArchiveDirPath)
		if err != nil {
			log.Fatal(err)
		}
	}()

	err = downloadAsset(asset)
	if err != nil {
		return err
	}

	err = archive.Extract(asset.ArchiveFilePath, asset.ArchiveDirPath, true)
	if err != nil {
		return err
	}

	err = helpers.CopyDirectoryContent(asset.ArchiveDirPath, filepath.Join(app.CacheDirPath, asset.Tag))
	if err != nil {
		return err
	}

	return nil
}

// downloadAsset downloads the release asset.
func downloadAsset(a *repository.Asset) error {
	// Create the file.
	a.ArchiveFilePath = filepath.Join(a.ArchiveDirPath, "hugo."+a.ArchiveExt)
	out, err := os.Create(a.ArchiveFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Get the data.
	resp, err := http.Get(a.ArchiveURL)
	if err != nil {
		return err
	}
	fmt.Printf("Downloading %s... ", a.Tag)
	defer resp.Body.Close()

	// Check server response.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file.
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("done.\n")

	return nil
}
