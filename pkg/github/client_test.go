package github_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	githubv17 "github.com/google/go-github/github"
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

	client := githubv17.NewClient(ts.Client())
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
