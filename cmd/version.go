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
	"log"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
	"golang.org/x/oauth2"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the " + App.Name + " version and check for a newer release",
	Long:  "Display the " + App.Name + " version and check for a newer release",
	Run: func(cmd *cobra.Command, args []string) {
		err := displayVersion()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func displayVersion() error {
	fmt.Println(versionString)

	latestVersion, err := getLatestHVMVersion(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if semver.Compare(latestVersion, version) == 1 {
		fmt.Println()
		fmt.Printf("A newer version of %s is available: %s\n", App.Name, latestVersion)
		fmt.Printf("Download the latest release here: %s\n", App.UpdateURL)
	}

	return nil
}

func getLatestHVMVersion(ctx context.Context) (string, error) {
	client := newGitHubClient()

	release, _, err := client.Repositories.GetLatestRelease(ctx, App.RepositoryOwner, App.RepositoryName)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}

	if release.TagName == nil {
		return "", fmt.Errorf("release found, but tag name is nil")
	}

	return *release.TagName, nil
}

func newGitHubClient() *github.Client {
	if Config.GitHubToken == "" {
		return github.NewClient(nil)
	} else {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: Config.GitHubToken},
		)
		tc := oauth2.NewClient(ctx, ts)

		return github.NewClient(tc)
	}
}
