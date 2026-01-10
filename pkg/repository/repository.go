/*
Copyright © 2026 Joe Mooring <joe.mooring@veriphor.com>

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

// Package repository provides types and operations for managing GitHub releases.
package repository

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/mod/semver"
)

// Repository represents a GitHub repository.
type Repository struct {
	owner     string         // Owner of the GitHub repository
	name      string         // Name of the GitHub repository
	tags      []string       // Repository tags
	latestTag string         // Latest repository tag
	client    *github.Client // A GitHub API client
}

// Asset represents a GitHub release asset for a given release, OS, and architecture.
type Asset struct {
	ArchiveDirPath  string // Directory path of the downloaded archive
	ArchiveExt      string // Extension of the downloaded archive: pkg, tar.gz, or zip
	ArchiveFilePath string // File path of the downloaded archive
	ArchiveURL      string // Download URL for this asset
	Tag             string // User-selected tag associated with the release for this asset
	ExecName        string // Name of the executable file
}

// NewRepository creates a new Repository instance and fetches tags.
func NewRepository(owner, name string, client *github.Client) (*Repository, error) {
	r := &Repository{
		owner:     owner,
		name:      name,
		client:    client,
		latestTag: "",
	}

	err := r.FetchTags()
	if err != nil {
		return nil, err
	}

	return r, nil
}

// NewAsset creates a new Asset instance.
func NewAsset(execName string) *Asset {
	return &Asset{
		ExecName: execName,
	}
}

// FetchTags fetches tags associated with recent releases.
func (r *Repository) FetchTags() error {
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
		// Releases prior to v0.54.0 were not semantically versioned.
		if semver.Compare(tag.GetName(), "v0.54.0") >= 0 {
			tagNames = append(tagNames, tag.GetName())
		}
	}

	r.latestTag = tagNames[0]
	r.tags = tagNames

	return nil
}

// GetLatestTag returns the most recent tag from the repository.
func (r *Repository) GetLatestTag(a *Asset) error {
	if r.latestTag == "" {
		return fmt.Errorf("no latest release found")
	}
	a.Tag = r.latestTag
	return nil
}

// GetTagFromString parses and validates a version string and sets it on the asset.
func (r *Repository) GetTagFromString(a *Asset, version string) error {
	inputVersion := version
	// fast return for simple cases
	if version == "latest" {
		return r.GetLatestTag(a)
	}
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if semver.Compare(version, "") == 0 {
		return fmt.Errorf("invalid tag: %s", inputVersion)
	}
	if !slices.Contains(r.tags, version) {
		return fmt.Errorf("tag \"%s\" not found in repository", inputVersion)
	}
	a.Tag = version
	return nil
}

// SelectTag prompts the user to select a tag from a list of recent tags.
// sortAscending determines the sort order, numTagsToDisplay limits the count.
// excludeExistsFn is used to check if a tag is already cached.
func (r *Repository) SelectTag(a *Asset, msg string, sortAscending bool, numTagsToDisplay int, excludeExistsFn func(string) (bool, error)) error {
	tags := r.tags

	// Make a copy for sorting operations
	displayTags := make([]string, len(tags))
	copy(displayTags, tags)

	// Apply sorting if requested
	if sortAscending {
		semver.Sort(displayTags)
	}

	// Apply limit
	n := numTagsToDisplay
	if n < 0 {
		n = 9999
	}

	if n <= len(displayTags) {
		if sortAscending {
			displayTags = displayTags[len(displayTags)-n:]
		} else {
			displayTags = displayTags[:n]
		}
	}

	// Display tags
	fmt.Println()
	for i, tag := range displayTags {
		exists, err := excludeExistsFn(tag)
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
		} else if i == len(displayTags)-1 {
			fmt.Print("\n")
		}
	}

	// Prompt and collect user input
	fmt.Println()
	fmt.Printf("%s: ", msg)

	for a.Tag == "" {
		var rs string
		fmt.Scanln(&rs)

		if rs == "" {
			fmt.Println("Canceled.")
			return nil
		}

		// Parse selection
		selection := 0
		_, err := fmt.Sscanf(strings.TrimSpace(rs), "%d", &selection)
		if err != nil || selection < 1 || selection > len(displayTags) {
			fmt.Printf("Please enter a number between 1 and %d, or press Enter to cancel: ", len(displayTags))
		} else {
			a.Tag = displayTags[selection-1]
		}
	}

	return nil
}

// FetchDownloadURL fetches the download URL for the user-selected tag.
// cacheLookupFn is used to check if a cached version exists.
//
//gocyclo:ignore
func (r *Repository) FetchDownloadURL(a *Asset) error {
	version := a.Tag[1:]
	pattern := ""
	// The archive file names were changed with the release of v0.103.0.
	if semver.Compare(a.Tag, "v0.103.0") == -1 {
		// prior to v0.103.0
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
	} else {
		// v0.103.0 and later
		switch runtime.GOOS {
		case "darwin":
			// The macOS archive file names were changed with the release of v0.153.0.
			if semver.Compare(a.Tag, "v0.153.0") == -1 {
				// prior to v0.153.0
				pattern = `hugo_extended_` + version + `_darwin-universal.tar.gz`
			} else {
				// v0.153.0 and later
				pattern = `hugo_extended_` + version + `_darwin-universal.pkg`
			}
		case "windows":
			pattern = `hugo_extended_` + version + `_windows-` + runtime.GOARCH + `.zip`
		case "linux":
			pattern = `hugo_extended_` + version + `_linux-` + runtime.GOARCH + `.tar.gz`
		default:
			return fmt.Errorf("unsupported operating system")
		}
	}
	re := regexp.MustCompile(pattern)

	release, _, err := r.client.Repositories.GetReleaseByTag(context.Background(), r.owner, r.name, a.Tag)
	if err != nil {
		return err
	}

	for _, asset := range release.Assets {
		archiveURL := asset.GetBrowserDownloadURL()
		if re.MatchString(archiveURL) {
			a.ArchiveURL = archiveURL
			switch {
			case strings.HasSuffix(a.ArchiveURL, ".pkg"):
				a.ArchiveExt = "pkg"
			case strings.HasSuffix(a.ArchiveURL, ".tar.gz"):
				a.ArchiveExt = "tar.gz"
			case strings.HasSuffix(a.ArchiveURL, ".zip"):
				a.ArchiveExt = "zip"
			default:
				return fmt.Errorf("unable to determine archive extension")
			}
			return nil
		}
	}

	return fmt.Errorf("unable to find download for %s %s/%s", a.Tag, runtime.GOOS, runtime.GOARCH)
}

// ExecPath returns the path of the executable file for this asset.
func (a *Asset) ExecPath(cachePath string) string {
	return fmt.Sprintf("%s/%s/%s", cachePath, a.Tag, a.ExecName)
}

// Tags returns the list of available tags.
func (r *Repository) Tags() []string {
	return r.tags
}

// LatestTag returns the latest tag.
func (r *Repository) LatestTag() string {
	return r.latestTag
}

// SetTagsForTesting allows tests to set tags directly.
// This is exported for testing purposes only.
func (r *Repository) SetTagsForTesting(tags []string, latestTag string) {
	r.tags = tags
	r.latestTag = latestTag
}
