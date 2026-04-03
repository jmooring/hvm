# Hugo Version Manager

[![test status](https://github.com/jmooring/hvm/actions/workflows/test.yaml/badge.svg)](https://github.com/jmooring/hvm/actions/workflows/test.yaml)
[![test coverage](https://img.shields.io/endpoint?logo=github&label=coverage&url=https://gist.githubusercontent.com/jmooring/3f9305867fe4d1e2dade602c9c66dde0/raw/hvm-coverage.json)](https://github.com/jmooring/hvm/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jmooring/hvm)](https://goreportcard.com/report/github.com/jmooring/hvm)
[![license](https://img.shields.io/github/license/jmooring/hvm)](https://github.com/jmooring/hvm/blob/main/LICENSE)
[![latest release](https://img.shields.io/github/v/release/jmooring/hvm?logo=github&v=1)](https://github.com/jmooring/hvm/releases/latest)
[![GitHub sponsors](https://img.shields.io/github/sponsors/jmooring?logo=github&label=sponsors)](https://github.com/sponsors/jmooring)

Hugo Version Manager (`hvm`) is a tool that helps you download, manage, and switch between different versions and editions of the [Hugo][] static site generator. You can also use `hvm` to install Hugo as a standalone application.

![Demonstration](assets/hvm.webp)

You can use `hvm` to:

- Download and manage multiple versions and editions of Hugo
- Switch between different versions and editions of Hugo for the current directory
- Install Hugo as a standalone application

Supported operating systems:

- Darwin (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64, arm64)

## How it works

The `hvm use` command allows you to switch between different versions and editions of Hugo for the current directory. To do this, `hvm` downloads, extracts, and caches the release asset for your operating system and architecture. It also creates an `.hvm` file for the current directory to store the version/edition.

If the current directory does not contain an `.hvm` file, the system searches the directories specified in the `PATH` environment variable to find the Hugo executable.

To use a different version and edition of Hugo, run the `hvm use` command and select or specify the desired version and edition. To use the Hugo executable found in the directories specified in your `PATH` environment variable, run the `hvm disable` command. This removes the `.hvm` file from the current directory.

Because `hvm` caches the extracted release assets, you don't have to download them again each time you switch the version or edition. You can view a list of cached assets, the size of the cache, and the cache location by running the `hvm status` command. You can also clean the cache by running the `hvm clean` command.

The `hvm install` command installs a version/edition of Hugo to use when version management is disabled for the current directory. This means that you can use `hvm` as a Hugo installer, even if you don't want to use its version management features.

The `hvm use` and `hvm install` commands present a list of available versions. After you select a version, you must select an edition. To always use the `defaultEdition` instead of prompting for the edition, set the `promptForEdition` configuration value to `false`. To directly specify a version and edition and bypass these prompts, include the desired version number or `latest` after the command. You may also append an optional edition suffix to the version as shown in the examples below:

```text
hvm use v0.159.1
hvm use 0.159.1
hvm use latest

hvm use v0.159.1/standard
hvm use 0.159.1/standard
hvm use latest/standard
```

## Installation

### Step 1 - Install the executable

Download a [prebuilt binary][] or install from source (requires Go 1.26.0 or later):

```text
go install github.com/jmooring/hvm@latest
```

### Step 2 - Create an alias

Create an alias for the `hugo` command that overrides the executable path if a
valid `.hvm` file exists for the current directory.

1. Run `hvm gen alias --help` to find the subcommand for the desired shell.
1. Run `hvm gen alias <shell> --help` to see the installation instructions.
1. Run `hvm gen alias <shell>` to generate an alias function for the specified shell.

The `hvm gen alias` command generates alias functions for bash, fish, zsh, and Windows PowerShell.

The alias function displays a brief status message each time it is called, if version management is enabled for the current directory. To disable this message, set the `hvm_show_status` variable to `false` in the alias function.

## Usage

```text
Usage:
  hvm [command]

Available Commands:
  clean       Clean the cache
  completion  Generate the autocompletion script for the specified shell
  config      Display the current configuration
  disable     Disable version management for the current directory
  gen         Generate various files
  help        Help about any command
  install     Install a version/edition to use when version management is disabled
  remove      Remove the version/edition used when version management is disabled
  status      Display the status
  use         Select or specify a version/edition for the current directory
  version     Display the hvm version and check for a newer release

Flags:
  -h, --help      Display help
  -v, --version   Display the hvm version

Use "hvm [command] --help" for more information about a command.
```

## Configuration

To locate the configuration file, run the `hvm config` command. This will print the path to the configuration file to the console. Keys in the configuration file are case insensitive.

To set configuration values with an environment variable, create an environment variable prefixed with `HVM_`. For example, to set the `sortAscending` configuration value to true:

```text
export HVM_SORTASCENDING=true
```

If a configuration value is set in multiple places, environment variables take precedence over values in the configuration file, which take precedence over built-in defaults.

**defaultEdition** (`string`)

The edition `hvm use` and `hvm install` select when you omit the edition and no selection prompt appears. The default is `standard`.

**gitHubToken** (`string`)

GitHub limits the number of requests that can be made to its API per hour to 60 for unauthenticated clients. If you exceed this limit, `hvm` will display a message indicating when the limit will be reset. This is typically within minutes.

If you regularly exceed this limit, you can create a GitHub personal access token with public repository (`public_repo`) scope. With a personal access token, GitHub limits API requests to 5,000 per hour. The corresponding environment variables are `HVM_GITHUB_TOKEN` and `HVM_GITHUBTOKEN`. If both are set, `HVM_GITHUB_TOKEN` takes precedence.

**numTagsToDisplay** (`int`)

By default, the `hvm use` and `hvm install` commands display the 30 most recent releases. To display all releases since v0.54.0, set the value to `-1`. Releases before v0.54.0 were not semantically versioned. The default is `30`.

**promptForEdition** (`bool`)

Whether `hvm use` and `hvm install` prompt for an edition when you omit the edition from the command. Setting this to `false` instructs `hvm` to select the `defaultEdition` instead. The default is `true`.

**sortAscending** (`bool`)

By default, the `hvm use` and `hvm install` commands display the list of recent releases in descending order. To display the list in ascending order, set this value to `true`. The default is `false`.

## Continuous integration and deployment (CI/CD)

For production workflows utilizing CI/CD (e.g., on Cloudflare, GitHub Pages, GitLab Pages, Netlify, Render, or Vercel), the Hugo Version Manager enables a reproducible build environment. The simplest and most reliable approach leverages the `.hvm` file:

1. Check the `.hvm` file into source control
1. Read the `.hvm` file in your build script to determine which Hugo version and edition to install

This guarantees that your site is always built with the specific version and edition it was developed for.

See this example of a site hosted with GitHub Pages:\
<https://github.com/jmooring/hosting-github-pages-hvm>

## In the news

Discover what others are saying about the Hugo Version Manager.

Date|Link
:--|:--
2025-12-26|<https://www.brycewray.com/posts/2025/12/new-normal-starting-hugo-0.153.x/>
2025-12-15|<https://home.expurple.me/posts/fearless-hugo-updates-with-hvm/>
2025-06-04|<https://harrycresswell.com/writing/hugo-version-manager/>
2025-05-23|<https://ahmedjama.com/blog/2025/05/hugo-version-manager/>
2025-04-07|<https://tsalikis.blog/posts/switching_hugo_versions/>
2025-03-24|<https://navendu.me/posts/hugo-version-manager/>
2023-12-18|<https://perplex.desider.at/news/hugo-version-manager/>

[Hugo]: https://github.com/gohugoio/hugo/#readme
[prebuilt binary]: https://github.com/jmooring/hvm/releases/latest
