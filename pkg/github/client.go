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

// Package github provides functionality for interacting with the GitHub API.
package github

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/go-github/v81/github"
	"golang.org/x/oauth2"
)

// NewClient creates and returns a new GitHub API client. If token is empty,
// returns an unauthenticated client.
func NewClient(token string) *github.Client {
	if token == "" {
		return github.NewClient(nil)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

// GetLatestRelease fetches the latest release for a given repository.
func GetLatestRelease(ctx context.Context, client *github.Client, owner, repo string) (string, error) {
	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("%s", ErrReason(err))
	}

	if release.TagName == nil {
		return "", fmt.Errorf("release found, but tag name is nil")
	}

	return *release.TagName, nil
}

// ErrReason returns a human-readable reason for a GitHub API error,
// including the rate limit reset time (local timezone) for rate limit errors.
func ErrReason(err error) string {
	var rle *github.RateLimitError
	if errors.As(err, &rle) {
		reset := rle.Rate.Reset.Time.In(time.Local)
		return fmt.Sprintf("GitHub API rate limit exceeded (resets at %s)", reset.Format("15:04:05 MST"))
	}
	var ere *github.ErrorResponse
	if errors.As(err, &ere) {
		return fmt.Sprintf("GitHub API error %d: %s", ere.Response.StatusCode, ere.Message)
	}
	return "unable to reach GitHub"
}
