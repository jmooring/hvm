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
	"slices"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/jmooring/hvm/pkg/archive"
	"github.com/jmooring/hvm/pkg/helpers"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
	"golang.org/x/oauth2"
)

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:     "use",
	Aliases: []string{"get"},
	Short:   "Select a version to use in the current directory",
	Long: `Displays a list of recent Hugo releases, prompting you to select a version
to use in the current directory. It then downloads, extracts, and caches the
release asset for your operating system and architecture and writes the version
tag to an .hvm file.

You may bypass the selection screen using --latest or --tag`,

	Run: func(cmd *cobra.Command, args []string) {
		useVersionInDotFile, err := cmd.Flags().GetBool("useVersionInDotFile")
		cobra.CheckErr(err)

		useLatest, err := cmd.Flags().GetBool("latest")
		cobra.CheckErr(err)

		useTag, err := cmd.Flags().GetString("tag")
		cobra.CheckErr(err)

		err = use(useVersionInDotFile, useLatest, useTag)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.Flags().Bool("useVersionInDotFile", false, "Use the version specified by the "+App.DotFileName+" file\nin the current directory")
	useCmd.Flags().Bool("latest", false, "Use the latest version")
	useCmd.Flags().String("tag", "", "Use specific tag version")
}

// A repository is a GitHub repository.
type repository struct {
	owner     string         // account owner of the GitHub repository
	name      string         // name of the GitHub repository without the .git extension
	tags      []string       // repository tags
	latestTag string         // latest repository tag
	client    *github.Client // a GitHub API client
}

// An asset is a GitHub asset for a given release, operating system, and architecture.
type asset struct {
	archiveDirPath  string // directory path of the downloaded archive
	archiveExt      string // extension of the downloaded archive: tar.gz or zip
	archiveFilePath string // file path of the downloaded archive
	archiveURL      string // download URL for this asset
	tag             string // user-selected tag associated with the release for this asset
	execName        string // name of the executable file
}

// use sets the version of the Hugo executable to use in the current directory.
func use(useVersionInDotFile bool, useLatest bool, useTag string) error {
	asset := newAsset()
	repo := newRepository()

	if useVersionInDotFile {
		version, err := getVersionFromDotFile(App.DotFilePath)
		if err != nil {
			return err
		}
		if version == "" {
			fmt.Fprintln(os.Stderr, "Error: the current directory does not contain an "+App.DotFileName+" file")
			os.Exit(1)
		}
		if !slices.Contains(repo.tags, version) {
			theFix := fmt.Sprintf("run \"%[1]s use\" to select a version, or \"%[1]s disable\" to remove the file", App.Name)
			fmt.Fprintf(os.Stderr, "Error: the version specified in the "+App.DotFileName+" file (%s) is not available in the repository: %s\n", version, theFix)
			os.Exit(1)
		}
		asset.tag = version
	} else if useLatest {
		err := repo.getLatestTag(asset)
		if err != nil {
			return err
		}
	} else if useTag != "" {
		err := repo.getSpecificTag(asset, useTag)
		if err != nil {
			return err
		}
	} else {
		msg := "Select a version to use in the current directory"
		err := repo.selectTag(asset, msg)
		if err != nil {
			return err
		}
		if asset.tag == "" {
			return nil // the user did not select a tag; do nothing
		}
	}

	exists, err := helpers.Exists(filepath.Join(App.CacheDirPath, asset.tag))
	if err != nil {
		return err
	}

	if exists {
		fmt.Printf("Using %s from cache.\n", asset.tag)
	} else {
		err := get(asset, repo)
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

// newAsset creates a new asset object, returning a pointer to same.
func newAsset() *asset {
	a := asset{
		execName: getExecName(),
	}

	return &a
}

// newRepository creates a new repository object, returning a pointer to same.
func newRepository() *repository {
	var client *github.Client

	if Config.GithubToken == "" {
		client = github.NewClient(nil)
	} else {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: Config.GithubToken},
		)
		tc := oauth2.NewClient(ctx, ts)

		client = github.NewClient(tc)
	}

	r := repository{
		client:    client,
		name:      App.RepositoryName,
		owner:     App.RepositoryOwner,
		latestTag: "",
	}

	err := r.fetchTags()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return &r
}

// get downloads and extracts the release asset.
func get(asset *asset, repo *repository) error {
	err := repo.fetchDownloadURL(asset)
	if err != nil {
		return err
	}

	asset.archiveDirPath, err = os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer func() {
		err := os.RemoveAll(asset.archiveDirPath)
		if err != nil {
			log.Fatal(err)
		}
	}()

	err = asset.downloadAsset()
	if err != nil {
		return err
	}

	err = archive.Extract(asset.archiveFilePath, asset.archiveDirPath, true)
	if err != nil {
		return err
	}

	err = helpers.CopyDirectoryContent(asset.archiveDirPath, filepath.Join(App.CacheDirPath, asset.tag))
	if err != nil {
		return err
	}

	return nil
}

// fetchTags fetches tags associated with recent releases.
func (r *repository) fetchTags() error {
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

	r.latestTag = tagNames[0]

	if Config.SortAscending {
		semver.Sort(tagNames)
	}
	r.tags = tagNames

	return nil
}

// getLatestTag returns the most recent tag from repository.
func (r *repository) getLatestTag(a *asset) error {
	if "" == r.latestTag {
		return fmt.Errorf("no latest release found")
	}
	a.tag = r.latestTag
	return nil
}

// selectTag prompts the user to select a tag from a list of recent tags.
func (r *repository) selectTag(a *asset, msg string) error {
	// List tags.
	fmt.Println()

	n := Config.NumTagsToDisplay
	if n < 0 {
		n = 9999
	}

	tags := r.tags
	if n <= len(r.tags) {
		if Config.SortAscending {
			tags = tags[len(tags)-n:]
		} else {
			tags = tags[:n]
		}
	}

	for i, tag := range tags {
		exists, err := helpers.Exists(filepath.Join(App.CacheDirPath, tag))
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
		} else if i == len(tags)-1 {
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
		if err != nil || ri < 1 || ri > len(tags) {
			fmt.Printf("Please enter a number between 1 and %d, or press Enter to cancel: ", len(tags))
		} else {
			a.tag = tags[ri-1]
		}
	}

	return nil
}

// fetchDownloadURL fetches the download URL for the user-selected tag.
func (r *repository) fetchDownloadURL(a *asset) error {
	// The archive file name was different with v0.102.3 and earlier.
	version := a.tag[1:]
	pattern := ""
	if semver.Compare(a.tag, "v0.102.3") == 1 {
		// v0.103.0 and later
		switch runtime.GOOS {
		case "darwin":
			pattern = `hugo_extended_` + version + `_darwin-universal.tar.gz`
		case "windows":
			pattern = `hugo_extended_` + version + `_windows-` + runtime.GOARCH + `.zip`
		case "linux":
			pattern = `hugo_extended_` + version + `_linux-` + runtime.GOARCH + `.tar.gz`
		default:
			return fmt.Errorf("unsupported operating system")
		}
	} else {
		// v0.102.3 and earlier
		switch runtime.GOOS {
		case "darwin":
			pattern = `hugo_extended_` + version + `_macOS-64bit.tar.gz`
		case "windows":
			pattern = `hugo_extended_` + version + `_Windows-64bit.zip`
		case "linux":
			pattern = `hugo_extended_` + version + `_Linux-64bit.tar.gz`
			if runtime.GOARCH == "arm64" {
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
		archiveURL := asset.GetBrowserDownloadURL()
		if regexp.MatchString(archiveURL) {
			a.archiveURL = archiveURL
			switch {
			case strings.HasSuffix(a.archiveURL, ".tar.gz"):
				a.archiveExt = "tar.gz"
			case strings.HasSuffix(a.archiveURL, ".zip"):
				a.archiveExt = "zip"
			default:
				return fmt.Errorf("unable to determine archive extension")
			}
			return nil
		}
	}

	return fmt.Errorf("unable to find download for %s %s/%s", a.tag, runtime.GOOS, runtime.GOARCH)
}

// downloadAsset downloads the release asset.
func (a *asset) downloadAsset() error {
	// Create the file.
	a.archiveFilePath = filepath.Join(a.archiveDirPath, "hugo."+a.archiveExt)
	out, err := os.Create(a.archiveFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Get the data.
	resp, err := http.Get(a.archiveURL)
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

	// Write the body to file.
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("done.\n")

	return nil
}

// getExecName returns the name of the Hugo executable file.
func getExecName() string {
	if runtime.GOOS == "windows" {
		return "hugo.exe"
	}
	return "hugo"
}

// getExecPath returns the path of the Hugo executable file.
func (a *asset) getExecPath() string {
	return filepath.Join(App.CacheDirPath, a.tag, a.execName)
}

// createDotFile creates an application dot file in the current directory.
func (a *asset) createDotFile() error {
	err := os.WriteFile(App.DotFilePath, []byte(a.tag), 0o644)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) getSpecificTag(a *asset, version string) error {
	// fast return for simple cases
	switch version {
	case "":
		return fmt.Errorf("invalid tag: %s", version)
	case "v": // latest tag is always valid
		a.tag = r.latestTag
		return nil
	}
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if strings.HasPrefix(version, "v.") { // empty major
		version = strings.Replace(version, "v.", semver.Major(r.latestTag)+".", 1)
	}
	if "" == semver.Canonical(version) {
		return fmt.Errorf("invalid tag: %s", version)
	}
	if version == semver.Canonical(version) { // major.minor.patch
		a.tag = version
	} else {
		type CompareSemVer func(semver string) bool
		var compareSemVer CompareSemVer
		if version == semver.Major(version) { // major only
			compareSemVer = func(t string) bool {
				return semver.Compare(semver.Major(version), semver.Major(t)) >= 0
			}
		} else if version == semver.MajorMinor(version) { // major.minor
			compareSemVer = func(t string) bool {
				return semver.Major(version) == semver.Major(t) && semver.Compare(semver.MajorMinor(version), semver.MajorMinor(t)) >= 0
			}
		}
		if Config.SortAscending {
			for i := len(r.tags) - 1; i >= 0; i-- {
				tag := r.tags[i]
				if compareSemVer(tag) {
					a.tag = tag
					return nil
				}
			}
		} else {
			for _, tag := range r.tags {
				if compareSemVer(tag) {
					a.tag = tag
					return nil
				}
			}
		}
	}
	if !slices.Contains(r.tags, a.tag) {
		return fmt.Errorf("tag not found: %s", version)
	}
	return nil
}
