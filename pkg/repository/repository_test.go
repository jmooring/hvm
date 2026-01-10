package repository

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"testing"

	"github.com/google/go-github/github"
)

func TestNewRepository_FetchTagsAndLatest(t *testing.T) {
	// Mock GitHub API server for /repos/{owner}/{repo}/tags
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/gohugoio/hugo/tags" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return tags including one below threshold
		_, _ = w.Write([]byte(`[
		  {"name":"v0.153.0"},
		  {"name":"v0.152.0"},
		  {"name":"v0.53.0"}
		]`))
	}))
	defer ts.Close()

	client := github.NewClient(ts.Client())
	baseURL, _ := url.Parse(ts.URL + "/")
	client.BaseURL = baseURL

	r, err := NewRepository("gohugoio", "hugo", client)
	if err != nil {
		t.Fatalf("NewRepository error: %v", err)
	}

	tags := r.Tags()
	if len(tags) == 0 {
		t.Fatal("Tags should not be empty")
	}
	for _, tag := range tags {
		if tag == "v0.53.0" {
			t.Fatal("v0.53.0 should be filtered out")
		}
	}
	if got := r.LatestTag(); got != "v0.153.0" {
		t.Fatalf("LatestTag: want v0.153.0 got %s", got)
	}
}

func TestAssetExecPath(t *testing.T) {
	a := &Asset{Tag: "v1.2.3", ExecName: "hugo"}
	p := a.ExecPath("/tmp/cache")
	if p != "/tmp/cache/v1.2.3/hugo" {
		t.Fatalf("ExecPath: want %q got %q", "/tmp/cache/v1.2.3/hugo", p)
	}
}

func TestFetchDownloadURL(t *testing.T) {
	// Mock GitHub API server for release by tag
	version := "v0.153.0"
	var assetName string
	switch runtime.GOOS {
	case "linux":
		assetName = "hugo_extended_0.153.0_linux-" + runtime.GOARCH + ".tar.gz"
	case "windows":
		assetName = "hugo_extended_0.153.0_windows-" + runtime.GOARCH + ".zip"
	case "darwin":
		assetName = "hugo_extended_0.153.0_darwin-universal.pkg"
	default:
		// Unsupported OS path; skip
		t.Skip("unsupported OS for FetchDownloadURL test")
	}

	assetURL := "https://example.com/" + assetName

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/gohugoio/hugo/releases/tags/"+version {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"assets":[{"browser_download_url":"` + assetURL + `"}]}`))
	}))
	defer ts.Close()

	client := github.NewClient(ts.Client())
	baseURL, _ := url.Parse(ts.URL + "/")
	client.BaseURL = baseURL

	r := &Repository{owner: "gohugoio", name: "hugo", client: client}
	a := &Asset{Tag: version}
	if err := r.FetchDownloadURL(a); err != nil {
		t.Fatalf("FetchDownloadURL error: %v", err)
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
	r.SetTagsForTesting([]string{"v1.2.3", "v1.2.2", "v1.2.1"}, "v1.2.3")
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
	r.SetTagsForTesting([]string{"v1.2.3", "v1.2.2", "v1.2.1"}, "v1.2.3")
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

	err = r.SelectTag(a, "Select a version", false, 3, func(tag string) (bool, error) { return false, nil })
	if err != nil {
		t.Fatalf("SelectTag error: %v", err)
	}
	if a.Tag != "v1.2.2" {
		t.Fatalf("SelectTag: want v1.2.2 got %s", a.Tag)
	}
}

func TestSelectTag_Cancel(t *testing.T) {
	r := &Repository{}
	r.SetTagsForTesting([]string{"v1.2.3", "v1.2.2", "v1.2.1"}, "v1.2.3")
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

	err = r.SelectTag(a, "Select a version", false, 3, func(tag string) (bool, error) { return false, nil })
	if err != nil {
		t.Fatalf("SelectTag cancel error: %v", err)
	}
	if a.Tag != "" {
		t.Fatalf("SelectTag cancel: want empty tag got %s", a.Tag)
	}
}

func TestFetchDownloadURL_NoMatch(t *testing.T) {
	// Setup OS-specific asset name that will NOT match expected pattern
	version := "v0.153.0"
	assetName := "not_matching_asset.ext"
	assetURL := "https://example.com/" + assetName

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/gohugoio/hugo/releases/tags/"+version {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"assets":[{"browser_download_url":"` + assetURL + `"}]}`))
	}))
	defer ts.Close()

	client := github.NewClient(ts.Client())
	baseURL, _ := url.Parse(ts.URL + "/")
	client.BaseURL = baseURL

	r := &Repository{owner: "gohugoio", name: "hugo", client: client}
	a := &Asset{Tag: version}
	if err := r.FetchDownloadURL(a); err == nil {
		t.Fatal("expected error when no asset matches pattern")
	}
}
