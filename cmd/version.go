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

	gh "github.com/jmooring/hvm/pkg/github"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the " + app.Name + " version and check for a newer release",
	Long:  "Display the " + app.Name + " version and check for a newer release",
	Run: func(cmd *cobra.Command, args []string) {
		err := displayVersion()
		cobra.CheckErr(err)
	},
}

// init registers the version command with the root command.
func init() {
	rootCmd.AddCommand(versionCmd)
}

// displayVersion prints the current version and checks for updates.
func displayVersion() error {
	fmt.Println(versionString)

	latestVersion, err := getLatestHVMVersion(context.Background())
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
		return nil
	}

	if semver.Compare(latestVersion, versionInfo.Version) == 1 {
		fmt.Println()
		fmt.Printf("A newer version of %s is available: %s\n", app.Name, latestVersion)
		fmt.Printf("Download the latest release here: %s\n", app.UpdateURL)
	}

	return nil
}

// getLatestHVMVersion fetches the latest hvm release version from GitHub.
func getLatestHVMVersion(ctx context.Context) (string, error) {
	client := gh.NewClient(config.GitHubToken)
	return gh.GetLatestRelease(ctx, client, app.RepositoryOwner, app.RepositoryName)
}
