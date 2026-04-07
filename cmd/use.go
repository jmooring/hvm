/*
Copyright 2023 Veriphor LLC

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
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jmooring/hvm/pkg/archive"
	"github.com/jmooring/hvm/pkg/dotfile"
	"github.com/jmooring/hvm/pkg/helpers"
	"github.com/jmooring/hvm/pkg/repository"
	"github.com/spf13/cobra"
)

// useCmd represents the use command.
var useCmd = &cobra.Command{
	Use:     "use [version] | [flags]",
	Aliases: []string{"get"},
	Short:   "Select or specify a version/edition for the current directory",
	Long: `Displays a list of recent Hugo releases, prompting you to select a
version/edition for the current directory. It then downloads, extracts,
and caches the release asset for your operating system and architecture and
writes the version/edition to an ` + app.DotFileName + ` file.

Bypass the selection menu by specifying a version or version/edition:

  hvm use v0.159.1
  hvm use v0.159.1/standard

  hvm use 0.159.1
  hvm use 0.159.1/standard

  hvm use latest
  hvm use latest/standard
`,
	Run: func(cmd *cobra.Command, args []string) {
		version := ""

		useVersionInDotFile, err := cmd.Flags().GetBool("useVersionInDotFile")
		cobra.CheckErr(err)

		if useVersionInDotFile {
			dm := dotfile.NewManager(app.DotFilePath, app.DotFileName, app.Name, config.DefaultEdition)
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
	useCmd.Flags().Bool("useVersionInDotFile", false, "Use the version specified by the "+app.DotFileName+" file\nfor the current directory")
}

// use sets the version/edition to use for the current directory.
func use(version string) error {
	asset, err := resolveAsset(version, "Select a version to use for the current directory", "Select an edition")
	if err != nil {
		return err
	}
	if asset == nil {
		return nil // user cancelled
	}

	exists, err := helpers.Exists(filepath.Join(app.CacheDirPath, asset.Tag, asset.Edition))
	if err != nil {
		return err
	}

	if exists {
		fmt.Printf("Using %s/%s from cache.\n", asset.Tag, asset.Edition)
	} else {
		err := downloadAndCache(asset)
		if err != nil {
			return err
		}
	}

	dm := dotfile.NewManager(app.DotFilePath, app.DotFileName, app.Name, config.DefaultEdition)
	err = dm.Write(asset.Tag + "/" + asset.Edition)
	if err != nil {
		return err
	}

	return nil
}

// downloadAndCache downloads and extracts the release asset.
// SetURLFromEditions must be called before this function to populate asset.ArchiveURL and asset.ArchiveExt.
func downloadAndCache(asset *repository.Asset) error {
	var err error
	asset.ArchiveDirPath, err = os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(asset.ArchiveDirPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temporary directory %s: %s\n", asset.ArchiveDirPath, err)
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

	err = helpers.CopyDirectoryContent(asset.ArchiveDirPath, filepath.Join(app.CacheDirPath, asset.Tag, asset.Edition))
	if err != nil {
		return err
	}

	return nil
}

// downloadAsset downloads the release asset.
func downloadAsset(a *repository.Asset) (retErr error) {
	// Create the file.
	a.ArchiveFilePath = filepath.Join(a.ArchiveDirPath, "hugo."+a.ArchiveExt)
	out, err := os.Create(a.ArchiveFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	// Get the data. Use a client with a connection/header timeout but no
	// overall timeout — Hugo release assets are large and streaming must
	// not be interrupted once it starts. Clone http.DefaultTransport so
	// that proxy settings, TLS configuration, and HTTP/2 support are
	// preserved. Use a comma-ok assertion in case DefaultTransport has
	// been replaced by a different RoundTripper; in that case fall back
	// to a transport with the same baseline settings as DefaultTransport
	// so that proxied/corporate environments still work.
	var transport *http.Transport
	if dt, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = dt.Clone()
	} else {
		transport = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}
	transport.ResponseHeaderTimeout = 30 * time.Second
	client := &http.Client{Transport: transport}
	resp, err := client.Get(a.ArchiveURL)
	if err != nil {
		return err
	}
	fmt.Printf("Downloading %s/%s... ", a.Tag, a.Edition)
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
