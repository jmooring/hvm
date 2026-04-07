/*
Copyright 2023 Veriphor LLC

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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jmooring/hvm/pkg/cache"
	gh "github.com/jmooring/hvm/pkg/github"
	"github.com/jmooring/hvm/pkg/helpers"
	"github.com/jmooring/hvm/pkg/repository"
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
	DefaultEdition   string `toml:"defaultEdition"`   // Default edition of the hugo executable to "use" or "install"
	GitHubToken      string `toml:"githubToken"`      // A GitHub personal access token
	NumTagsToDisplay int    `toml:"numTagsToDisplay"` // Number of tags to display when using the "use" and "install" commands
	PromptForEdition bool   `toml:"promptForEdition"` // Whether to prompt the user to select an edition when using the "use" or "install" commands
	SortAscending    bool   `toml:"sortAscending"`    // Whether to display the tags in ascending order
}

// orderedAvailableEditions returns the subset of repository.ValidEditions that are
// present in the editions map, preserving the ValidEditions order.
func orderedAvailableEditions(editions map[string]string) []string {
	var available []string
	for _, e := range repository.ValidEditions {
		if _, ok := editions[e]; ok {
			available = append(available, e)
		}
	}
	return available
}

// resolveEdition resolves the edition for asset from an explicit value, an
// interactive prompt, or the configured default. It returns true if the user
// cancelled the prompt.
func resolveEdition(asset *repository.Asset, editions map[string]string, explicitEdition, promptMsg string) (bool, error) {
	if explicitEdition != "" {
		if !slices.Contains(repository.ValidEditions, explicitEdition) {
			s, err := helpers.JoinWithConjunction(repository.ValidEditions, "or")
			if err != nil {
				return false, err
			}
			return false, fmt.Errorf("edition %q is invalid, must be one of %s", explicitEdition, s)
		}
		if _, ok := editions[explicitEdition]; !ok {
			return false, fmt.Errorf("edition %q is not available for %s", explicitEdition, asset.Tag)
		}
		asset.Edition = explicitEdition
	} else if config.PromptForEdition {
		err := repository.SelectEdition(asset, promptMsg, orderedAvailableEditions(editions))
		if err != nil {
			return false, err
		}
		if asset.Edition == "" {
			return true, nil // the user cancelled the prompt
		}
	} else {
		if _, ok := editions[config.DefaultEdition]; !ok {
			return false, fmt.Errorf("edition %q is not available for %s", config.DefaultEdition, asset.Tag)
		}
		asset.Edition = config.DefaultEdition
	}
	return false, nil
}

// promptYesNo prints question with a (Y/n) or (y/N) suffix and loops until
// the user enters "y", "n", or Enter (which selects the default). It returns
// true if the user confirmed.
func promptYesNo(question string, defaultYes bool) bool {
	suffix := "(y/N)"
	if defaultYes {
		suffix = "(Y/n)"
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s %s: ", question, suffix)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// EOF or any read error on non-interactive stdin: treat as "no".
			if err != io.EOF {
				return false
			}
			// EOF with a partial line — fall through to evaluate what was read.
		}
		r := strings.TrimSpace(line)
		if r == "" {
			return defaultYes
		}
		if strings.EqualFold(string(r[0]), "y") {
			return true
		}
		if strings.EqualFold(string(r[0]), "n") {
			return false
		}
		if err == io.EOF {
			// Unrecognised input at EOF: treat as "no".
			return false
		}
		fmt.Printf("%s %s: ", question, suffix)
	}
}

// resolveAsset resolves the asset for the given version string, tag prompt
// message, and edition prompt message. It returns nil if the user cancelled
// an interactive prompt. version may be empty (triggers interactive tag
// selection), a bare tag ("v0.153.0"), or a tag/edition pair ("v0.153.0/extended").
func resolveAsset(version, tagMsg, editionMsg string) (*repository.Asset, error) {
	asset := repository.NewAsset(cache.ExecName())

	client := gh.NewClient(config.GitHubToken)
	repo, err := repository.NewRepository(app.ManagedApp.RepositoryOwner, app.ManagedApp.RepositoryName, client, app.CacheDirPath)
	if err != nil {
		return nil, err
	}

	var explicitEdition string
	if version == "" {
		err := repo.SelectTag(asset, tagMsg, config.SortAscending, config.NumTagsToDisplay)
		if err != nil {
			return nil, err
		}
		if asset.Tag == "" {
			return nil, nil // the user cancelled tag selection; do nothing
		}
	} else {
		tag := version
		if idx := strings.Index(tag, "/"); idx != -1 {
			explicitEdition = tag[idx+1:]
			if explicitEdition == "" {
				return nil, fmt.Errorf("invalid version/edition %q: edition must not be empty", version)
			}
			tag = tag[:idx]
		}
		err := repo.GetTagFromString(asset, tag)
		if err != nil {
			return nil, err
		}
	}

	editions, err := repo.FetchEditions(asset)
	if err != nil {
		return nil, err
	}

	cancelled, err := resolveEdition(asset, editions, explicitEdition, editionMsg)
	if err != nil {
		return nil, err
	}
	if cancelled {
		return nil, nil // the user cancelled edition selection; do nothing
	}

	if err := asset.SetURLFromEditions(editions); err != nil {
		return nil, err
	}

	return asset, nil
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
between different versions and editions of the Hugo static site generator.
You can also use ` + app.Name + ` to install Hugo as a standalone application.`,
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
	viper.SetDefault("defaultEdition", "standard")
	viper.SetDefault("githubToken", "")
	viper.SetDefault("numTagsToDisplay", 30)
	viper.SetDefault("promptForEdition", true)
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
	if val := os.Getenv("HVM_GITHUB_TOKEN"); val != "" {
		viper.Set("githubToken", val)
	}
	viper.AutomaticEnv()

	// Validate config value data types.
	k := "defaultEdition"
	if !helpers.IsString(viper.Get(k)) {
		err = fmt.Errorf("configuration: %s must be a string: see %s", k, viper.ConfigFileUsed())
		cobra.CheckErr(err)
	}

	k = "githubToken"
	if !helpers.IsString(viper.Get(k)) {
		err = fmt.Errorf("configuration: %s must be a string: see %s", k, viper.ConfigFileUsed())
		cobra.CheckErr(err)
	}

	k = "numTagsToDisplay"
	if viper.GetInt(k) == 0 {
		err = fmt.Errorf("configuration: %s must be a non-zero integer: see %s", k, viper.ConfigFileUsed())
		cobra.CheckErr(err)
	}

	// Validate the defaultEdition value.
	defaultEdition := viper.GetString("defaultEdition")
	if !slices.Contains(repository.ValidEditions, defaultEdition) {
		s, err := helpers.JoinWithConjunction(repository.ValidEditions, "or")
		cobra.CheckErr(err)
		err = fmt.Errorf("configuration: %s %q is invalid, must be one of %s: see %s", "defaultEdition", defaultEdition, s, viper.ConfigFileUsed())
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

	n, err := cache.EnsureSchema(app.CacheDirPath, app.DefaultDirName)
	cobra.CheckErr(err)
	if n > 0 {
		fmt.Fprintf(os.Stderr, "Info: cache migrated to new format: %d old cached version(s) removed\n", n)
	}
}
