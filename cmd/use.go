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
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmooring/hvm/archive"
	"github.com/jmooring/hvm/dotfile"
	"github.com/jmooring/hvm/helpers"
	"github.com/jmooring/hvm/repository"
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

	client := newHTTPClient()

	digest, err := downloadAsset(asset, client)
	if err != nil {
		return err
	}

	if asset.ChecksumsURL != "" {
		archiveFilename := path.Base(asset.ArchiveURL)
		expected, err := fetchExpectedChecksum(client, asset.ChecksumsURL, archiveFilename)
		if err != nil {
			return err
		}
		if digest != expected {
			return fmt.Errorf("checksum mismatch for %s: got %s, expected %s", archiveFilename, digest, expected)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Warning: no checksums file found for %s; skipping integrity check\n", asset.Tag)
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

// newHTTPClient returns an HTTP client with a connection/header timeout but no
// overall timeout — Hugo release assets are large and streaming must not be
// interrupted once it starts. It clones http.DefaultTransport so that proxy
// settings, TLS configuration, and HTTP/2 support are preserved.
func newHTTPClient() *http.Client {
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
	return &http.Client{Transport: transport}
}

// downloadAsset downloads the release asset and returns its SHA-256 hex digest.
func downloadAsset(a *repository.Asset, client *http.Client) (digest string, retErr error) {
	// Create the file.
	a.ArchiveFilePath = filepath.Join(a.ArchiveDirPath, "hugo."+a.ArchiveExt)
	out, err := os.Create(a.ArchiveFilePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := out.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	resp, err := client.Get(a.ArchiveURL)
	if err != nil {
		return "", err
	}
	fmt.Printf("Downloading %s/%s... ", a.Tag, a.Edition)
	defer resp.Body.Close()

	// Check server response.
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file, computing SHA-256 as we go.
	h := sha256.New()
	_, err = io.Copy(out, io.TeeReader(resp.Body, h))
	if err != nil {
		return "", err
	}
	fmt.Printf("done.\n")

	return hex.EncodeToString(h.Sum(nil)), nil
}

// fetchExpectedChecksum downloads the checksums file and returns the expected
// SHA-256 hex digest for the named archive file.
func fetchExpectedChecksum(client *http.Client, checksumsURL, archiveFilename string) (string, error) {
	resp, err := client.Get(checksumsURL)
	if err != nil {
		return "", fmt.Errorf("downloading checksums file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading checksums file: bad status: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[1] == archiveFilename {
			return fields[0], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading checksums file: %w", err)
	}

	return "", fmt.Errorf("no checksum found for %s in checksums file", archiveFilename)
}
