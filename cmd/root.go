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

// Package cmd implements CLI commands.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jmooring/hvm/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type config struct {
	GithubToken      string `toml:"githubToken,omitempty" comment:"GitHub personal access token"`
	NumTagsToDisplay int    `toml:"numTagsToDisplay" comment:"Number of tags to display with the \"use\" and \"install\" commands"`
}

var Config config

var (
	cacheDir      string
	buildDate     string
	commitHash    string
	version       string = "dev"
	versionString string = fmt.Sprintf("%s %s %s/%s %s %s", appName, version, runtime.GOOS, runtime.GOARCH, commitHash, buildDate)
)

const (
	appName        string = "hvm"
	dotFileName    string = ".hvm"
	defaultDirName string = "default"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "Hugo Version Manager",
	Long: `Hugo Version Manager (` + appName + `) allows you to download and manage multiple versions
of the extended edition of the Hugo static site generator. You can use hvm to
specify which version of Hugo to use in the current directory.

If you do not specify a version of Hugo to use in the current directory, the
Hugo executable will be found by searching the PATH environment variable.
`,
	Version: versionString,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initCache)

	rootCmd.PersistentFlags().BoolP("help", "h", false, "Display help")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Display the "+appName+" version")
	rootCmd.SetVersionTemplate(fmt.Sprintf("%s\n", versionString))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Set default values.
	viper.SetDefault("numTagsToDisplay", 30)
	viper.SetDefault("githubToken", "")

	// Create config directory.
	userConfigDir, err := os.UserConfigDir()
	cobra.CheckErr(err)
	err = os.MkdirAll(filepath.Join(userConfigDir, appName), 0777)
	cobra.CheckErr(err)

	// Define config file.
	viper.AddConfigPath(filepath.Join(userConfigDir, appName))
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	// Create config file if it doesn't exist.
	viper.SafeWriteConfig()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found. Ignore.
		} else {
			cobra.CheckErr(err)
		}
	}

	// Get config values from env vars.
	viper.SetEnvPrefix("HVM")
	viper.AutomaticEnv()

	// Validate config values.
	k := "numTagsToDisplay"
	if viper.GetInt(k) == 0 {
		err = fmt.Errorf("configuration: %s must be a non-zero integer: see %s", k, viper.ConfigFileUsed())
		cobra.CheckErr(err)
	}
	k = "githubToken"
	if !helpers.IsString(viper.Get(k)) {
		err = fmt.Errorf("configuration: %s must be a string: see %s", k, viper.ConfigFileUsed())
		cobra.CheckErr(err)
	}

	// Unmarshal the config to the Config struct.
	err = viper.Unmarshal(&Config)
	cobra.CheckErr(err)
}

// initCache creates the cache directory.
func initCache() {
	userCacheDir, err := os.UserCacheDir()
	cobra.CheckErr(err)
	cacheDir = filepath.Join(userCacheDir, appName)
	err = os.MkdirAll(cacheDir, 0777)
	cobra.CheckErr(err)
}
