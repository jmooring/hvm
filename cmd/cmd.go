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
	"strings"

	"github.com/jmooring/hvm/pkg/helpers"
	"github.com/jmooring/hvm/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// An application contains details about the application. Some are constants, while
// others depend on the user environment.
type application struct {
	CacheDirPath    string     // Path to the application cache directory
	ConfigDirPath   string     // Path to the application configuration directory
	ConfigFilePath  string     // Path to the application configuration file
	DefaultDirName  string     // Name of the "default" directory within the application cache directory
	DefaultDirPath  string     // Path to the "default" directory within the application cache directory
	DotFileName     string     // Name of the dot file written to the current directory (e.g., .hvm)
	DotFilePath     string     // Path to the dot file
	ManagedApp      managedApp // Details about the application being managed
	Name            string     // Name of the application
	RepositoryName  string     // Name of the GitHub repository
	RepositoryOwner string     // Owner of the GitHub repository
	UpdateURL       string     // URL to update the application
	WorkingDir      string     // Current working directory
}

// A managedApp contains details about the external application being managed,
// such as its source repository location.
type managedApp struct {
	RepositoryName  string // Name of the GitHub repository
	RepositoryOwner string // Owner of the GitHub repository
}

// A configuration contains the current configuration parameters from environment
// variables, the configuration file, or default values, in that order.
type configuration struct {
	GitHubToken      string `toml:"githubToken"`      // A GitHub personal access token
	NumTagsToDisplay int    `toml:"numTagsToDisplay"` // Number of tags to display when using the "use" and "install" commands
	SortAscending    bool   `toml:"sortAscending"`    // Whether to display the tags in ascending order
}

var app application = application{
	DefaultDirName: "default",
	DotFileName:    ".hvm",
	ManagedApp: managedApp{
		RepositoryName:  "hugo",
		RepositoryOwner: "gohugoio",
	},
	Name:            "hvm",
	RepositoryName:  "hvm",
	RepositoryOwner: "jmooring",
	UpdateURL:       "https://github.com/jmooring/hvm/releases/latest",
}

var config configuration

var versionInfo = version.NewInfo(app.Name)

var versionString = versionInfo.String()

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   app.Name,
	Short: "Hugo Version Manager",
	Long: `Hugo Version Manager (` + app.Name + `) is a tool that helps you download, manage, and switch
between different versions of the Hugo static site generator. You can also use
` + app.Name + ` to install Hugo as a standalone application.`,
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

// init initializes the root command and sets up flags.
func init() {
	cobra.OnInitialize(initConfig, initApp)

	rootCmd.PersistentFlags().BoolP("help", "h", false, "Display help")
	rootCmd.Flags().BoolP("version", "v", false, "Display the "+app.Name+" version")
	rootCmd.SetVersionTemplate(fmt.Sprintf("%s\n", versionString))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Set default values.
	viper.SetDefault("githubToken", "")
	viper.SetDefault("numTagsToDisplay", 30)
	viper.SetDefault("sortAscending", false)

	// Create config directory.
	userConfigDir, err := os.UserConfigDir()
	cobra.CheckErr(err)
	err = os.MkdirAll(filepath.Join(userConfigDir, app.Name), 0o777)
	cobra.CheckErr(err)

	// Define config file.
	viper.AddConfigPath(filepath.Join(userConfigDir, app.Name))
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
	viper.SetEnvPrefix(strings.ToUpper(app.Name))
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
	err = viper.Unmarshal(&config)
	cobra.CheckErr(err)
}

// initApp initializes the application and creates the application cache
// directory.
func initApp() {
	userCacheDir, err := os.UserCacheDir()
	cobra.CheckErr(err)

	userConfigDir, err := os.UserConfigDir()
	cobra.CheckErr(err)

	wd, err := os.Getwd()
	cobra.CheckErr(err)

	app.CacheDirPath = filepath.Join(userCacheDir, app.Name)
	app.ConfigDirPath = filepath.Join(userConfigDir, app.Name)
	app.ConfigFilePath = viper.ConfigFileUsed()
	app.DefaultDirPath = filepath.Join(userCacheDir, app.Name, app.DefaultDirName)
	app.DotFilePath = filepath.Join(wd, app.DotFileName)
	app.WorkingDir = wd

	err = os.MkdirAll(app.CacheDirPath, 0o777)
	cobra.CheckErr(err)
}
