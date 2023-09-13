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
	"strings"

	"github.com/jmooring/hvm/pkg/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// An application contains details about the application. Some are constants, while
// others depend on the user environment.
type application struct {
	CacheDirPath    string // path to the application cache directory
	ConfigFilePath  string // path to the application configuration
	DefaultDirName  string // name of the "default" directory within the application cache directory
	DefaultDirPath  string // path to the "default" directory within the application cache directory
	DotFileName     string // name of the dot file written to the current directory (e.g., .hvm)
	Name            string // name of the application
	RepositoryName  string // name of the GitHub repository without the .git extension
	RepositoryOwner string // account owner of the GitHub repository
}

// A configuration contains the current configuration parameters from environment
// variables, the configuration file, or default values, in that order.
type configuration struct {
	GithubToken      string // a GitHub personal access token
	NumTagsToDisplay int    // number of tags to display when using the "use" and "install" commands
}

var App application = application{
	DefaultDirName:  "default",
	DotFileName:     ".hvm",
	Name:            "hvm",
	RepositoryName:  "hugo",
	RepositoryOwner: "gohugoio",
}
var Config configuration

var (
	buildDate     string = runtime.GOOS
	commitHash    string
	version       string = "dev"
	versionString string = fmt.Sprintf("%s %s %s/%s %s %s", App.Name, version, runtime.GOOS, runtime.GOARCH, commitHash, buildDate)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   App.Name,
	Short: "Hugo Version Manager",
	Long: `Hugo Version Manager (` + App.Name + `) allows you to download and manage multiple versions
of the extended edition of the Hugo static site generator. You can use hvm to
specify which version of Hugo to use in the current directory.

If you do not specify a version of Hugo to use in the current directory, the
Hugo executable will be found by searching the PATH environment variable.`,
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
	cobra.OnInitialize(initConfig, initApp)

	rootCmd.PersistentFlags().BoolP("help", "h", false, "Display help")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Display the "+App.Name+" version")
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
	err = os.MkdirAll(filepath.Join(userConfigDir, App.Name), 0777)
	cobra.CheckErr(err)

	// Define config file.
	viper.AddConfigPath(filepath.Join(userConfigDir, App.Name))
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
	viper.SetEnvPrefix(strings.ToUpper(App.Name))
	viper.AutomaticEnv()

	// Validate config values.
	k := "numTagsToDisplay"
	if viper.GetInt(k) == 0 {
		err = fmt.Errorf("configuration: %s must be a non-zero integer: see %s", k, App.ConfigFilePath)
		cobra.CheckErr(err)
	}
	k = "githubToken"
	if !helpers.IsString(viper.Get(k)) {
		err = fmt.Errorf("configuration: %s must be a string: see %s", k, App.ConfigFilePath)
		cobra.CheckErr(err)
	}

	// Unmarshal the config to the Config struct.
	err = viper.Unmarshal(&Config)
	cobra.CheckErr(err)
}

// initApp initializes the application and creates the application cache
// directory.
func initApp() {
	userCacheDir, err := os.UserCacheDir()
	cobra.CheckErr(err)

	App.CacheDirPath = filepath.Join(userCacheDir, App.Name)
	App.ConfigFilePath = viper.ConfigFileUsed()
	App.DefaultDirPath = filepath.Join(userCacheDir, App.Name, App.DefaultDirName)

	err = os.MkdirAll(App.CacheDirPath, 0777)
	cobra.CheckErr(err)
}
