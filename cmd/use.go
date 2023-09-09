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
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/jmooring/hvm/archive"
	"github.com/jmooring/hvm/helpers"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:     "use",
	Aliases: []string{"get"},
	Short:   "Select a version to use in the current directory",
	Long: `Displays a list of recent Hugo releases, prompting you to select a version
to use in the current directory. It then downloads, extracts, and caches the
release asset for your operating system and architecture.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := use()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}

// A repository is a GitHub repository.
type repository struct {
	owner  string         // account owner of the GitHub repository
	name   string         // name of the GitHub repository without the .git extension
	tags   []string       // repository tags in semver ascending order
	client *github.Client // a GitHub API client
}

// An asset is a GitHub asset for a given release, operating system, and architecture.
type asset struct {
	arch        string // client architecture: amd64, arm64, etc.
	archiveExt  string // extension of the downloaded asset archive: tar.gz or zip
	archivePath string // path to the downloaded asset archive
	os          string // client operating system: linux, darwin, windows, etc.
	tag         string // the user-selected tag associated with the release for this asset
	url         string // download URL for this asset
}

// use sets the version of the Hugo executable to use in the current directory.
func use() error {
	repo, err := newRepository("gohugoio", "hugo")
	if err != nil {
		return err
	}

	asset, err := newAsset(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	err = asset.fetchTags(repo)
	if err != nil {
		return err
	}

	msg := "Select a version to use in the current directory"
	err = asset.selectTag(repo, msg)
	if err != nil {
		return err
	}
	if asset.tag == "" {
		return nil // the user did not select a tag; do nothing
	}

	exists, _, err := helpers.Exists(filepath.Join(cacheDir, asset.tag))
	if err != nil {
		return err
	}

	if exists {
		fmt.Printf("Using %s from cache.\n", asset.tag)
	} else {
		err = asset.fetchURL(repo)
		if err != nil {
			return err
		}

		err = asset.downloadAsset()
		if err != nil {
			return err
		}

		err = archive.Extract(asset.archivePath, filepath.Dir(asset.archivePath), true)
		if err != nil {
			return err
		}
	}

	err = asset.createDotFile()
	if err != nil {
		return err
	}

	return nil
}

// newRepository creates a new repository object, returning a pointer to same.
func newRepository(owner string, name string) (*repository, error) {
	r := repository{
		owner:  owner,
		name:   name,
		client: github.NewClient(nil),
	}

	return &r, nil
}

// newAsset creates a new asset object, returning a pointer to same.
func newAsset(os string, arch string) (*asset, error) {
	a := asset{
		os:   os,
		arch: arch,
	}

	return &a, nil
}

// fetchTags fetches tags associated with recent releases.
func (a *asset) fetchTags(r *repository) error {
	opts := &github.ListOptions{PerPage: 100}

	var tags []*github.RepositoryTag
	for {
		tagsThisPage, resp, err := r.client.Repositories.ListTags(context.Background(), r.owner, r.name, opts)
		if err != nil {
			return err
		}
		tags = append(tags, tagsThisPage...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if len(tags) == 0 {
		return fmt.Errorf("this repository has no tags")
	}

	var tagNames []string
	for _, tag := range tags {
		if semver.Compare(tag.GetName(), "v0.54.0") >= 0 {
			tagNames = append(tagNames, tag.GetName())
		}
	}

	semver.Sort(tagNames)

	n := Config.NumTagsToDisplay
	if n < 0 {
		n = 9999
	}
	if n <= len(tagNames) {
		tagNames = tagNames[len(tagNames)-n:]
	}
	r.tags = tagNames

	return nil
}

// selectTag prompts the user to select a tag from a list of recent tags.
func (a *asset) selectTag(r *repository, msg string) error {
	// List tags.
	fmt.Println()

	for i, tag := range r.tags {
		exists, _, err := helpers.Exists(filepath.Join(cacheDir, tag))
		if err != nil {
			return err
		}
		var flag string
		if exists {
			flag = "[cached]"
		}
		// The tags array is zero-based, but choices are one-based.
		fmt.Printf("%-25s", fmt.Sprintf("%3d) %s %s", i+1, tag, flag))
		if (i+1)%3 == 0 {
			fmt.Print("\n")

		} else if i == len(r.tags)-1 {
			fmt.Print("\n")
		}
	}

	// Prompt user.
	fmt.Println()
	fmt.Printf("%s: ", msg)

	// Validate response.
	for a.tag == "" {
		var rs string
		fmt.Scanln(&rs)

		// User pressed return; do nothing.
		if len(rs) == 0 {
			fmt.Println("Canceled.")
			return nil
		}

		ri, err := strconv.Atoi(strings.TrimSpace(rs))
		if err != nil || ri < 1 || ri > len(r.tags) {
			fmt.Printf("Please enter a number between 1 and %d, or press Enter to cancel: ", len(r.tags))
		} else {
			a.tag = r.tags[ri-1]
		}
	}

	return nil
}

// fetchURL fetches the download URL for the user-selected tag.
func (a *asset) fetchURL(r *repository) error {
	// The archive file name was different with v0.102.3 and earlier.
	version := a.tag[1:]
	pattern := ""
	if semver.Compare(a.tag, "v0.102.3") == 1 {
		// v0.103.0 and later
		switch a.os {
		case "darwin":
			pattern = `hugo_extended_` + version + `_darwin-universal.tar.gz`
		case "windows":
			pattern = `hugo_extended_` + version + `_windows-` + a.arch + `.zip`
		case "linux":
			pattern = `hugo_extended_` + version + `_linux-` + a.arch + `.tar.gz`
		default:
			return fmt.Errorf("unsupported operating system")
		}
	} else {
		// v0.102.3 and earlier
		switch a.os {
		case "darwin":
			pattern = `hugo_extended_` + version + `_macOS-64bit.tar.gz`
		case "windows":
			pattern = `hugo_extended_` + version + `_Windows-64bit.zip`
		case "linux":
			pattern = `hugo_extended_` + version + `_Linux-64bit.tar.gz`
			if a.arch == "arm64" {
				pattern = `hugo_extended_` + version + `_Linux-ARM64.deb`
			}
		default:
			return fmt.Errorf("unsupported operating system")
		}
	}
	regexp := regexp.MustCompile(pattern)

	release, _, err := r.client.Repositories.GetReleaseByTag(context.Background(), r.owner, r.name, a.tag)
	if err != nil {
		return err
	}

	for _, asset := range release.Assets {
		downloadURL := asset.GetBrowserDownloadURL()
		if regexp.MatchString(downloadURL) {
			a.url = downloadURL
			switch {
			case strings.HasSuffix(a.url, ".tar.gz"):
				a.archiveExt = "tar.gz"
			case strings.HasSuffix(a.url, ".zip"):
				a.archiveExt = "zip"
			default:
				return fmt.Errorf("unable to determine archive extension")
			}
			return nil
		}
	}

	return fmt.Errorf("unable to find download for %s %s/%s", a.tag, a.os, a.arch)
}

func (a *asset) downloadAsset() error {
	// Create the download directory.
	err := os.MkdirAll(filepath.Join(cacheDir, a.tag), 0777)
	if err != nil {
		return err
	}

	// Create the file.
	a.archivePath = filepath.Join(cacheDir, a.tag, "hugo."+a.archiveExt)
	out, err := os.Create(a.archivePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Get the data.
	resp, err := http.Get(a.url)
	if err != nil {
		return err
	}
	fmt.Printf("Downloading %s... ", a.tag)
	defer resp.Body.Close()

	// Check server response.
	if resp.StatusCode != http.StatusOK {
		if err != nil {
			return fmt.Errorf("bad status: %s", resp.Status)
		}
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("done.\n")

	return nil
}

// createDotFile creates an .hvm file in the current directory
func (a *asset) createDotFile() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(wd, dotFileName), []byte(a.getExecPath()+"\n"), 0644)
	if err != nil {
		return err
	}

	return nil
}

// getExecPath returns the path of the Hugo executable file.
func (a *asset) getExecPath() string {
	return filepath.Join(cacheDir, a.tag, a.getExecName())
}

// getExecName returns the name of the Hugo executable file.
func (a *asset) getExecName() string {
	execName := "hugo"
	if a.os == "windows" {
		execName = "hugo.exe"
	}

	return execName
}
