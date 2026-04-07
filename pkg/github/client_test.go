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

package github_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	githubv81 "github.com/google/go-github/v81/github"
	gh "github.com/jmooring/hvm/pkg/github"
)

func TestGetLatestRelease(t *testing.T) {
	// Mock GitHub API server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/jmooring/hvm/releases/latest" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name":"v1.2.3"}`))
	}))
	defer ts.Close()

	client := githubv81.NewClient(ts.Client())
	baseURL, _ := url.Parse(ts.URL + "/")
	client.BaseURL = baseURL

	tag, err := gh.GetLatestRelease(context.Background(), client, "jmooring", "hvm")
	if err != nil {
		t.Fatalf("GetLatestRelease error: %v", err)
	}
	if tag != "v1.2.3" {
		t.Fatalf("want v1.2.3 got %s", tag)
	}
}

func TestNewClient(t *testing.T) {
	// empty token should create an unauthenticated client without error
	c := gh.NewClient("")
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
	// non-empty should also create a client
	c = gh.NewClient("token123")
	if c == nil {
		t.Fatal("NewClient with token returned nil")
	}
}

func TestErrReason_RateLimit(t *testing.T) {
	reset := time.Now().Add(10 * time.Minute)
	err := &githubv81.RateLimitError{
		Rate: githubv81.Rate{Reset: githubv81.Timestamp{Time: reset}},
	}
	got := gh.ErrReason(err)
	if !strings.HasPrefix(got, "GitHub API rate limit exceeded") {
		t.Errorf("ErrReason rate limit: unexpected message %q", got)
	}
	if !strings.Contains(got, "resets at") {
		t.Errorf("ErrReason rate limit: expected reset time in message, got %q", got)
	}
}

func TestErrReason_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer ts.Close()

	client := githubv81.NewClient(ts.Client())
	baseURL, _ := url.Parse(ts.URL + "/")
	client.BaseURL = baseURL

	_, _, err := client.Repositories.GetLatestRelease(context.Background(), "o", "r")
	if err == nil {
		t.Fatal("expected error from mock server")
	}
	got := gh.ErrReason(err)
	if !strings.HasPrefix(got, "GitHub API error 404") {
		t.Errorf("ErrReason API error: unexpected message %q", got)
	}
}

func TestErrReason_NetworkError(t *testing.T) {
	err := fmt.Errorf("dial tcp: connection refused")
	got := gh.ErrReason(err)
	if got != "unable to reach GitHub" {
		t.Errorf("ErrReason network error: want %q got %q", "unable to reach GitHub", got)
	}
}

func TestGetLatestRelease_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"Internal Server Error"}`))
	}))
	defer ts.Close()

	client := githubv81.NewClient(ts.Client())
	baseURL, _ := url.Parse(ts.URL + "/")
	client.BaseURL = baseURL

	_, err := gh.GetLatestRelease(context.Background(), client, "jmooring", "hvm")
	if err == nil {
		t.Fatal("GetLatestRelease: expected error for API error response")
	}
}

func TestGetLatestRelease_NilTagName(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client := githubv81.NewClient(ts.Client())
	baseURL, _ := url.Parse(ts.URL + "/")
	client.BaseURL = baseURL

	_, err := gh.GetLatestRelease(context.Background(), client, "jmooring", "hvm")
	if err == nil {
		t.Fatal("GetLatestRelease: expected error when tag_name is nil")
	}
}
