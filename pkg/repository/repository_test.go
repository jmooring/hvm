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

package repository

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-github/v81/github"
)

func TestNewRepository_FetchTagsAndLatest(t *testing.T) {
	// Mock GitHub API server for /repos/{owner}/{repo}/tags
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/gohugoio/hugo/tags" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return tags including one below threshold and one pre-release.
		_, _ = w.Write([]byte(`[
		  {"name":"v0.153.0"},
		  {"name":"v0.152.0"},
		  {"name":"v0.152.0-beta"},
		  {"name":"v0.53.0"}
		]`))
	}))
	defer ts.Close()

	client := github.NewClient(ts.Client())
	baseURL, _ := url.Parse(ts.URL + "/")
	client.BaseURL = baseURL

	r, err := NewRepository("gohugoio", "hugo", client, "")
	if err != nil {
		t.Fatalf("NewRepository error: %v", err)
	}

	tags := r.tags
	if len(tags) == 0 {
		t.Fatal("Tags should not be empty")
	}
	var foundPrerelease bool
	for _, tag := range tags {
		if tag == "v0.53.0" {
			t.Fatal("v0.53.0 should be filtered out (below threshold)")
		}
		if tag == "v0.152.0-beta" {
			foundPrerelease = true
		}
	}
	if !foundPrerelease {
		t.Fatal("pre-release v0.152.0-beta should be included in tags")
	}
	if got := r.latestTag; got != "v0.153.0" {
		t.Fatalf("LatestTag: want v0.153.0 got %s", got)
	}
}

func TestAssetExecPath(t *testing.T) {
	cacheDir := filepath.Join("tmp", "cache")
	a := &Asset{Tag: "v1.2.3", Edition: "extended", ExecName: "hugo"}
	p := a.ExecPath(cacheDir)
	want := filepath.Join(cacheDir, "v1.2.3", "extended", "hugo")
	if p != want {
		t.Fatalf("ExecPath: want %q got %q", want, p)
	}
}

func TestSetURLFromEditions(t *testing.T) {
	var assetURL string
	switch runtime.GOOS {
	case "linux":
		assetURL = "https://example.com/hugo_extended_0.153.0_linux-" + runtime.GOARCH + ".tar.gz"
	case "windows":
		assetURL = "https://example.com/hugo_extended_0.153.0_windows-" + runtime.GOARCH + ".zip"
	case "darwin":
		assetURL = "https://example.com/hugo_extended_0.153.0_darwin-universal.pkg"
	default:
		t.Skip("unsupported OS")
	}

	editions := map[string]string{
		"extended": assetURL,
		"standard": "https://example.com/hugo_0.153.0_linux-amd64.tar.gz",
	}
	a := &Asset{Tag: "v0.153.0", Edition: "extended"}
	if err := a.SetURLFromEditions(editions); err != nil {
		t.Fatalf("SetURLFromEditions error: %v", err)
	}
	if a.ArchiveURL != assetURL {
		t.Fatalf("ArchiveURL: want %q got %q", assetURL, a.ArchiveURL)
	}
	if a.ArchiveExt == "" {
		t.Fatal("ArchiveExt should be set")
	}
}

func TestGetTagFromString(t *testing.T) {
	r := &Repository{}
	r.tags = []string{"v1.2.3", "v1.2.2", "v1.2.1"}
	r.latestTag = "v1.2.3"
	a := NewAsset("hugo")

	// latest
	if err := r.GetTagFromString(a, "latest"); err != nil {
		t.Fatalf("latest error: %v", err)
	}
	if a.Tag != "v1.2.3" {
		t.Fatalf("latest: want v1.2.3 got %s", a.Tag)
	}

	// explicit v
	a.Tag = ""
	if err := r.GetTagFromString(a, "v1.2.2"); err != nil {
		t.Fatalf("explicit v error: %v", err)
	}
	if a.Tag != "v1.2.2" {
		t.Fatalf("explicit v: want v1.2.2 got %s", a.Tag)
	}

	// no v prefix
	a.Tag = ""
	if err := r.GetTagFromString(a, "1.2.1"); err != nil {
		t.Fatalf("no v error: %v", err)
	}
	if a.Tag != "v1.2.1" {
		t.Fatalf("no v: want v1.2.1 got %s", a.Tag)
	}

	// not found
	a.Tag = ""
	if err := r.GetTagFromString(a, "v9.9.9"); err == nil {
		t.Fatal("expected error for not-found tag")
	}

	// invalid format
	a.Tag = ""
	if err := r.GetTagFromString(a, "x1"); err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestSelectTag_SelectSecond(t *testing.T) {
	r := &Repository{}
	r.tags = []string{"v1.2.3", "v1.2.2", "v1.2.1"}
	r.latestTag = "v1.2.3"
	a := NewAsset("hugo")

	// Prepare stdin with selection "2"
	orig := os.Stdin
	defer func() { os.Stdin = orig }()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdin = pr
	_, _ = fmt.Fprintln(pw, "2")
	pw.Close()

	err = r.SelectTag(a, "Select a version", false, 3)
	if err != nil {
		t.Fatalf("SelectTag error: %v", err)
	}
	if a.Tag != "v1.2.2" {
		t.Fatalf("SelectTag: want v1.2.2 got %s", a.Tag)
	}
}

func TestSelectTag_Cancel(t *testing.T) {
	r := &Repository{}
	r.tags = []string{"v1.2.3", "v1.2.2", "v1.2.1"}
	r.latestTag = "v1.2.3"
	a := NewAsset("hugo")

	// Prepare stdin with empty line to cancel
	orig := os.Stdin
	defer func() { os.Stdin = orig }()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdin = pr
	_, _ = fmt.Fprintln(pw, "")
	pw.Close()

	err = r.SelectTag(a, "Select a version", false, 3)
	if err != nil {
		t.Fatalf("SelectTag cancel error: %v", err)
	}
	if a.Tag != "" {
		t.Fatalf("SelectTag cancel: want empty tag got %s", a.Tag)
	}
}

func TestSetURLFromEditions_NotAvailable(t *testing.T) {
	editions := map[string]string{
		"standard": "https://example.com/hugo_0.153.0_linux-amd64.tar.gz",
	}
	a := &Asset{Tag: "v0.153.0", Edition: "extended"}
	if err := a.SetURLFromEditions(editions); err == nil {
		t.Fatal("expected error when edition not available")
	}
}

func TestSetURLFromEditions_UnknownExtension(t *testing.T) {
	editions := map[string]string{
		"standard": "https://example.com/hugo_0.153.0_linux-amd64.tar.bz2",
	}
	a := &Asset{Tag: "v0.153.0", Edition: "standard"}
	if err := a.SetURLFromEditions(editions); err == nil {
		t.Fatal("expected error for unrecognised archive extension")
	}
}

func TestSelectEdition_Select(t *testing.T) {
	a := NewAsset("hugo")
	editions := []string{"standard", "withdeploy", "extended", "extended_withdeploy"}

	orig := os.Stdin
	defer func() { os.Stdin = orig }()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdin = pr
	_, _ = fmt.Fprintln(pw, "3")
	pw.Close()

	if err := SelectEdition(a, "Select an edition", editions); err != nil {
		t.Fatalf("SelectEdition error: %v", err)
	}
	if a.Edition != "extended" {
		t.Fatalf("SelectEdition: want %q got %q", "extended", a.Edition)
	}
}

func TestSelectEdition_Cancel(t *testing.T) {
	a := NewAsset("hugo")
	editions := []string{"standard", "withdeploy", "extended", "extended_withdeploy"}

	orig := os.Stdin
	defer func() { os.Stdin = orig }()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdin = pr
	_, _ = fmt.Fprintln(pw, "")
	pw.Close()

	if err := SelectEdition(a, "Select an edition", editions); err != nil {
		t.Fatalf("SelectEdition cancel error: %v", err)
	}
	if a.Edition != "" {
		t.Fatalf("SelectEdition cancel: want empty got %q", a.Edition)
	}
}

func TestParseEdition(t *testing.T) {
	const baseURL = "https://github.com/gohugoio/hugo/releases/download/"
	mkURL := func(tag, filename string) string {
		return baseURL + tag + "/" + filename
	}

	type tc struct {
		name   string
		tag    string
		url    string
		want   string
		wantOK bool
	}

	const modernTag = "v0.153.0"
	const modernVer = "0.153.0"

	var modernSuffix string
	switch runtime.GOOS {
	case "darwin":
		modernSuffix = "_darwin-universal.pkg"
	case "windows":
		modernSuffix = "_windows-" + runtime.GOARCH + ".zip"
	case "linux":
		modernSuffix = "_linux-" + runtime.GOARCH + ".tar.gz"
	default:
		t.Skip("unsupported OS")
	}

	tests := []tc{
		// Modern tag — all four editions
		{"standard", modernTag, mkURL(modernTag, "hugo_"+modernVer+modernSuffix), "standard", true},
		{"extended", modernTag, mkURL(modernTag, "hugo_extended_"+modernVer+modernSuffix), "extended", true},
		{"extended_withdeploy", modernTag, mkURL(modernTag, "hugo_extended_withdeploy_"+modernVer+modernSuffix), "extended_withdeploy", true},
		{"withdeploy", modernTag, mkURL(modernTag, "hugo_withdeploy_"+modernVer+modernSuffix), "withdeploy", true},
		// Wrong platform suffix
		{"wrong_platform", modernTag, mkURL(modernTag, "hugo_"+modernVer+"_freebsd-amd64.tar.gz"), "", false},
		// Unknown edition prefix
		{"unknown_prefix", modernTag, mkURL(modernTag, "hugo_unknown_"+modernVer+modernSuffix), "", false},
	}

	// Pre-v0.103.0 legacy suffix
	const legacyTag = "v0.100.0"
	const legacyVer = "0.100.0"
	switch runtime.GOOS {
	case "darwin":
		tests = append(tests, tc{"legacy_darwin", legacyTag, mkURL(legacyTag, "hugo_"+legacyVer+"_macOS-64bit.tar.gz"), "standard", true})
	case "windows":
		tests = append(tests, tc{"legacy_windows", legacyTag, mkURL(legacyTag, "hugo_"+legacyVer+"_Windows-64bit.zip"), "standard", true})
	case "linux":
		if runtime.GOARCH == "arm64" {
			// Pre-v0.103.0 linux arm64 is unsupported
			tests = append(tests, tc{"legacy_linux_arm64_unsupported", legacyTag, mkURL(legacyTag, "hugo_"+legacyVer+"_Linux-64bit.tar.gz"), "", false})
		} else {
			tests = append(tests, tc{"legacy_linux", legacyTag, mkURL(legacyTag, "hugo_"+legacyVer+"_Linux-64bit.tar.gz"), "standard", true})
		}
	}

	// v0.103.0–v0.153.0 era: darwin uses different suffix
	if runtime.GOOS == "darwin" {
		const midTag = "v0.120.0"
		const midVer = "0.120.0"
		tests = append(tests, tc{"darwin_mid_era", midTag, mkURL(midTag, "hugo_"+midVer+"_darwin-universal.tar.gz"), "standard", true})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseEdition(tt.tag, tt.url)
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("parseEdition(%q, %q) = (%q, %v), want (%q, %v)", tt.tag, tt.url, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestFetchEditions(t *testing.T) {
	const tag = "v0.153.0"
	const ver = "0.153.0"

	var suffix string
	switch runtime.GOOS {
	case "darwin":
		suffix = "_darwin-universal.pkg"
	case "windows":
		suffix = "_windows-" + runtime.GOARCH + ".zip"
	case "linux":
		suffix = "_linux-" + runtime.GOARCH + ".tar.gz"
	default:
		t.Skip("unsupported OS")
	}

	mkAssetJSON := func(name string) string {
		url := "https://github.com/gohugoio/hugo/releases/download/" + tag + "/" + name
		return `{"browser_download_url":"` + url + `"}`
	}

	standardFile := "hugo_" + ver + suffix
	extendedFile := "hugo_extended_" + ver + suffix
	ignoredFile := "hugo_" + ver + "_checksums.txt"

	assetsJSON := "[" + mkAssetJSON(standardFile) + "," + mkAssetJSON(extendedFile) + "," + mkAssetJSON(ignoredFile) + "]"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name":"` + tag + `","assets":` + assetsJSON + `}`))
	}))
	defer ts.Close()

	client := github.NewClient(ts.Client())
	baseURL, _ := url.Parse(ts.URL + "/")
	client.BaseURL = baseURL

	r := &Repository{owner: "gohugoio", name: "hugo", client: client}
	a := &Asset{Tag: tag}

	editions, err := r.FetchEditions(a)
	if err != nil {
		t.Fatalf("FetchEditions error: %v", err)
	}
	if _, ok := editions["standard"]; !ok {
		t.Error("expected standard edition")
	}
	if _, ok := editions["extended"]; !ok {
		t.Error("expected extended edition")
	}
	if len(editions) != 2 {
		t.Errorf("expected 2 editions, got %d", len(editions))
	}
}

func TestFetchEditions_NoMatches(t *testing.T) {
	const tag = "v0.153.0"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name":"` + tag + `","assets":[{"browser_download_url":"https://example.com/hugo_0.153.0_checksums.txt"}]}`))
	}))
	defer ts.Close()

	client := github.NewClient(ts.Client())
	baseURL, _ := url.Parse(ts.URL + "/")
	client.BaseURL = baseURL

	r := &Repository{owner: "gohugoio", name: "hugo", client: client}
	a := &Asset{Tag: tag}

	_, err := r.FetchEditions(a)
	if err == nil {
		t.Fatal("expected error when no matching downloads found")
	}
}

func TestFirstStableTag_AllPrerelease(t *testing.T) {
	tags := []string{"v1.2.3-beta", "v1.2.2-rc.1", "v1.2.1-alpha"}
	got := firstStableTag(tags)
	if got != "v1.2.3-beta" {
		t.Fatalf("firstStableTag: want %q got %q", "v1.2.3-beta", got)
	}
}

func TestFirstStableTag_Empty(t *testing.T) {
	got := firstStableTag(nil)
	if got != "" {
		t.Fatalf("firstStableTag: want empty string got %q", got)
	}
}

func TestGetLatestTag_NoLatest(t *testing.T) {
	r := &Repository{}
	a := NewAsset("hugo")
	if err := r.GetLatestTag(a); err == nil {
		t.Fatal("GetLatestTag: expected error when latestTag is empty")
	}
}
