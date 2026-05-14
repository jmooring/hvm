/*
Copyright 2024 Veriphor LLC

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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jmooring/hvm/pkg/cache"
	"github.com/jmooring/hvm/pkg/repository"
)

// newMockRepo creates a mock repository for testing with predefined tags.
func newMockRepo() *repository.Repository {
	r := &repository.Repository{}
	r.SetTagsForTesting([]string{"v1.2.3", "v1.2.2", "v1.2.1"}, "v1.2.3")
	return r
}

var testCases = []struct {
	given string
	want  string
}{
	// valid existing version
	{"latest", "v1.2.3"},
	{"v1.2.3", "v1.2.3"},
	{"1.2.3", "v1.2.3"},
	{"v1.2.2", "v1.2.2"},
	{"1.2.2", "v1.2.2"},
	{"v1.2.1", "v1.2.1"},
	{"1.2.1", "v1.2.1"},
	// valid missing version
	{"v1.2.0", ""},
	{"1.2.0", ""},
	// invalid strings
	{"", ""},
	{"late", ""},
	{"a.b.c", ""},
	{"1", ""},
	{"1.2.", ""},
	{"1.2.3.", ""},
	{"1.2.3.4", ""},
}

// TestGetTagFromString tests the GetTagFromString repository method.
func TestGetTagFromString(t *testing.T) {
	// mock needed elements and sort config
	config.SortAscending = false
	repo := newMockRepo()
	asset := repository.NewAsset(cache.ExecName())
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("[%s] ShouldBe [%s]", tc.given, tc.want), func(t *testing.T) {
			asset.Tag = ""
			err := repo.GetTagFromString(asset, tc.given)
			if asset.Tag != tc.want {
				t.Fatalf("given(%s) -> calculated(%s) -> want(%s) : %s", tc.given, asset.Tag, tc.want, err)
			}
		})
	}
}

// TestFetchExpectedChecksum_Found verifies that the expected hash is returned
// when the archive filename is present in the checksums file.
func TestFetchExpectedChecksum_Found(t *testing.T) {
	const filename = "hugo_0.153.0_linux-amd64.tar.gz"
	const wantHash = "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	body := wantHash + "  " + filename + "\n" +
		"0000000000000000000000000000000000000000000000000000000000000000  other.tar.gz\n"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	got, err := fetchExpectedChecksum(&http.Client{}, ts.URL, filename)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != wantHash {
		t.Fatalf("want %q got %q", wantHash, got)
	}
}

// TestFetchExpectedChecksum_NotFound verifies that an error is returned when
// the archive filename is absent from the checksums file.
func TestFetchExpectedChecksum_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("abcdef1234567890  other.tar.gz\n"))
	}))
	defer ts.Close()

	_, err := fetchExpectedChecksum(&http.Client{}, ts.URL, "hugo_0.153.0_linux-amd64.tar.gz")
	if err == nil {
		t.Fatal("expected error for filename not present in checksums file")
	}
}

// TestFetchExpectedChecksum_BadStatus verifies that an error is returned when
// the checksums server responds with a non-200 status.
func TestFetchExpectedChecksum_BadStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	_, err := fetchExpectedChecksum(&http.Client{}, ts.URL, "hugo_0.153.0_linux-amd64.tar.gz")
	if err == nil {
		t.Fatal("expected error for HTTP error response")
	}
}

// TestDownloadAndCache_ChecksumMismatch verifies that downloadAndCache returns
// a checksum mismatch error when the downloaded archive's SHA-256 digest does
// not match the value in the checksums file. The function should return before
// attempting extraction, so no real archive is required.
func TestDownloadAndCache_ChecksumMismatch(t *testing.T) {
	const archiveFile = "hugo_0.153.0_linux-amd64.tar.gz"
	const wrongHash = "0000000000000000000000000000000000000000000000000000000000000000"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/" + archiveFile:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("fake archive content"))
		case "/checksums":
			w.WriteHeader(http.StatusOK)
			// Deliberately wrong hash — real digest of "fake archive content" is not all zeros.
			_, _ = w.Write([]byte(wrongHash + "  " + archiveFile + "\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	asset := &repository.Asset{
		ArchiveURL:   ts.URL + "/" + archiveFile,
		ChecksumsURL: ts.URL + "/checksums",
		ArchiveExt:   "tar.gz",
		Tag:          "v0.153.0",
		Edition:      "standard",
	}

	err := downloadAndCache(asset)
	if err == nil {
		t.Fatal("expected checksum mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("want 'checksum mismatch' error, got: %v", err)
	}
}
