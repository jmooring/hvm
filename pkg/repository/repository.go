/*
Copyright 2026 Veriphor LLC

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
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/google/go-github/v81/github"
	"github.com/jmooring/hvm/pkg/cache"
	gh "github.com/jmooring/hvm/pkg/github"
	"golang.org/x/mod/semver"
)

// Repository represents a GitHub repository.
type Repository struct {
	owner        string         // Owner of the GitHub repository
	name         string         // Name of the GitHub repository
	tags         []string       // Repository tags
	latestTag    string         // Latest repository tag
	client       *github.Client // A GitHub API client
	cacheDirPath string         // Local cache directory for tag list caching
}

// ValidEditions is the canonical ordered list of Hugo edition names.
var ValidEditions = []string{"standard", "withdeploy", "extended", "extended_withdeploy"}

// Asset represents a GitHub release asset for a given release, OS, and architecture.
type Asset struct {
	ArchiveDirPath  string // Directory path of the downloaded archive
	ArchiveExt      string // Extension of the downloaded archive: pkg, tar.gz, or zip
	ArchiveFilePath string // File path of the downloaded archive
	ArchiveURL      string // Download URL for this asset
	Edition         string // Edition of the release asset (e.g. standard, extended)
	Tag             string // User-selected tag associated with the release for this asset
	ExecName        string // Name of the executable file
}

// NewRepository creates a new Repository instance and fetches releases.
// cacheDirPath is the local hvm cache directory used to persist the release
// list between invocations. Pass an empty string to disable caching.
func NewRepository(owner, name string, client *github.Client, cacheDirPath string) (*Repository, error) {
	r := &Repository{
		owner:        owner,
		name:         name,
		client:       client,
		cacheDirPath: cacheDirPath,
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
	// Load the cached tag list if available.
	var cached []string
	if r.cacheDirPath != "" {
		var err error
		cached, err = loadTagCache(r.cacheDirPath)
		if err != nil {
			if errors.Is(err, errCorruptCache) {
				// Invalid JSON — warn, delete the file so it self-heals, and do a full fetch.
				fmt.Fprintf(os.Stderr, "Warning: corrupt release cache; deleting and re-fetching\n")
				_ = os.Remove(filepath.Join(r.cacheDirPath, cache.TagListFileName))
			} else {
				// Read error (e.g. permission denied) — warn but leave the file alone.
				fmt.Fprintf(os.Stderr, "Warning: could not read release cache: %s\n", err)
			}
			cached = nil
		}
	}

	if len(cached) > 0 {
		// Fetch the first page of tags to check whether a new one has appeared.
		recent, _, err := r.client.Repositories.ListTags(
			context.Background(), r.owner, r.name,
			&github.ListOptions{PerPage: 20},
		)
		if err != nil {
			// GitHub unreachable or rate-limited — fall back to cache with a warning.
			fmt.Fprintf(os.Stderr, "Warning: %s; using cached release list\n", gh.ErrReason(err))
			r.tags = cached
			r.latestTag = firstStableTag(cached)
			return nil
		}
		if len(recent) > 0 && recent[0].GetName() == cached[0] {
			// Cache is current.
			r.tags = cached
			r.latestTag = firstStableTag(cached)
			return nil
		}
		// A new tag exists — fall through to full fetch.
	}

	// Full fetch from GitHub.
	opts := &github.ListOptions{PerPage: 100}
	var tagNames []string
	for {
		tags, resp, err := r.client.Repositories.ListTags(context.Background(), r.owner, r.name, opts)
		if err != nil {
			if len(cached) > 0 {
				fmt.Fprintf(os.Stderr, "Warning: %s; using cached release list\n", gh.ErrReason(err))
				r.tags = cached
				r.latestTag = firstStableTag(cached)
				return nil
			}
			return fmt.Errorf("%s", gh.ErrReason(err))
		}
		for _, tag := range tags {
			name := tag.GetName()
			// Tags prior to v0.54.0 were not semantically versioned.
			if semver.Compare(name, "v0.54.0") >= 0 {
				tagNames = append(tagNames, name)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if len(tagNames) == 0 {
		return fmt.Errorf("no tags found")
	}

	r.latestTag = firstStableTag(tagNames)
	r.tags = tagNames

	// Persist to cache (best-effort).
	if r.cacheDirPath != "" {
		_ = saveTagCache(r.cacheDirPath, tagNames)
	}

	return nil
}

// firstStableTag returns the first tag in the list with no semver pre-release
// suffix. Falls back to the first tag if none are stable, and returns an empty
// string for an empty list.
func firstStableTag(tags []string) string {
	for _, tag := range tags {
		if semver.Prerelease(tag) == "" {
			return tag
		}
	}
	if len(tags) > 0 {
		return tags[0]
	}
	return ""
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
	if !semver.IsValid(version) {
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
func (r *Repository) SelectTag(a *Asset, msg string, sortAscending bool, numTagsToDisplay int) error {
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
		n = math.MaxInt
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
		// The tags array is zero-based, but choices are one-based.
		fmt.Printf("%-20s", fmt.Sprintf("%3d) %s", i+1, tag))
		if (i+1)%4 == 0 {
			fmt.Print("\n")
		} else if i == len(displayTags)-1 {
			fmt.Print("\n")
		}
	}

	// Prompt and collect user input
	fmt.Println()
	fmt.Printf("%s (Enter to cancel): ", msg)

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
			fmt.Printf("Please enter a number between 1 and %d, or Enter to cancel: ", len(displayTags))
		} else {
			a.Tag = displayTags[selection-1]
		}
	}

	fmt.Printf("Selected: %s\n", a.Tag)
	return nil
}

// FetchEditions fetches all available edition download URLs for the asset's
// tag on the current OS and architecture. The returned map is keyed by edition
// name (e.g. "standard", "extended", "extended_withdeploy", "withdeploy").
func (r *Repository) FetchEditions(a *Asset) (map[string]string, error) {
	release, _, err := r.client.Repositories.GetReleaseByTag(context.Background(), r.owner, r.name, a.Tag)
	if err != nil {
		return nil, err
	}

	editions := map[string]string{}
	for _, asset := range release.Assets {
		url := asset.GetBrowserDownloadURL()
		edition, ok := parseEdition(a.Tag, url)
		if ok {
			editions[edition] = url
		}
	}

	if len(editions) == 0 {
		return nil, fmt.Errorf("no downloads found for %s %s/%s", a.Tag, runtime.GOOS, runtime.GOARCH)
	}
	return editions, nil
}

// parseEdition returns the edition name for a given asset download URL on the
// current OS and architecture, or false if the URL does not match.
func parseEdition(tag, url string) (string, bool) {
	version := tag[1:] // strip leading "v"

	// Determine the expected filename suffix for this OS/arch/version.
	var suffix string
	switch runtime.GOOS {
	case "darwin":
		if semver.Compare(tag, "v0.103.0") == -1 {
			suffix = "_macOS-64bit.tar.gz"
		} else if semver.Compare(tag, "v0.153.0") == -1 {
			suffix = "_darwin-universal.tar.gz"
		} else {
			suffix = "_darwin-universal.pkg"
		}
	case "windows":
		if semver.Compare(tag, "v0.103.0") == -1 {
			suffix = "_Windows-64bit.zip"
		} else {
			suffix = "_windows-" + runtime.GOARCH + ".zip"
		}
	case "linux":
		if semver.Compare(tag, "v0.103.0") == -1 {
			if runtime.GOARCH == "arm64" {
				return "", false // .deb not supported
			}
			suffix = "_Linux-64bit.tar.gz"
		} else {
			suffix = "_linux-" + runtime.GOARCH + ".tar.gz"
		}
	default:
		return "", false
	}

	if !strings.HasSuffix(url, suffix) {
		return "", false
	}

	// Extract the edition from the filename prefix before the version number.
	// URL base looks like: hugo[_edition]_<version>_<platform>.<ext>
	base := url[strings.LastIndex(url, "/")+1:]
	base = strings.TrimSuffix(base, "_"+version+suffix)

	switch base {
	case "hugo":
		return "standard", true
	case "hugo_extended":
		return "extended", true
	case "hugo_extended_withdeploy":
		return "extended_withdeploy", true
	case "hugo_withdeploy":
		return "withdeploy", true
	}
	return "", false
}

// SelectEdition prompts the user to select an edition from the given ordered list.
// On return, a.Edition is set to the chosen edition, or empty if cancelled.
func SelectEdition(a *Asset, msg string, editions []string) error {
	fmt.Println()
	for i, e := range editions {
		fmt.Printf("%d) %s", i+1, e)
		if i < len(editions)-1 {
			fmt.Print("  ")
		}
	}
	fmt.Println()

	fmt.Println()
	fmt.Printf("%s (Enter to cancel): ", msg)

	for a.Edition == "" {
		var s string
		fmt.Scanln(&s)
		if s == "" {
			fmt.Println("Canceled.")
			return nil
		}
		n := 0
		if _, err := fmt.Sscanf(s, "%d", &n); err != nil || n < 1 || n > len(editions) {
			fmt.Printf("Please enter a number between 1 and %d, or Enter to cancel: ", len(editions))
		} else {
			a.Edition = editions[n-1]
		}
	}

	fmt.Printf("Selected: %s\n\n", a.Edition)
	return nil
}

// SetURLFromEditions sets ArchiveURL and ArchiveExt on the asset from the
// editions map using the already-set Edition field.
// Returns an error if the edition is not present in the map.
func (a *Asset) SetURLFromEditions(editions map[string]string) error {
	url, ok := editions[a.Edition]
	if !ok {
		return fmt.Errorf("edition %q is not available for %s on %s/%s", a.Edition, a.Tag, runtime.GOOS, runtime.GOARCH)
	}
	a.ArchiveURL = url
	switch {
	case strings.HasSuffix(url, ".pkg"):
		a.ArchiveExt = "pkg"
	case strings.HasSuffix(url, ".tar.gz"):
		a.ArchiveExt = "tar.gz"
	case strings.HasSuffix(url, ".zip"):
		a.ArchiveExt = "zip"
	default:
		return fmt.Errorf("unrecognised archive extension in URL: %s", url)
	}
	return nil
}

// ExecPath returns the path of the executable file for this asset.
func (a *Asset) ExecPath(cachePath string) string {
	return filepath.Join(cachePath, a.Tag, a.Edition, a.ExecName)
}

// SetTagsForTesting sets tags and latestTag directly. For use in tests only.
func (r *Repository) SetTagsForTesting(tags []string, latestTag string) {
	r.tags = tags
	r.latestTag = latestTag
}
