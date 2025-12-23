# Hugo Version Manager

[![test status](https://github.com/jmooring/hvm/actions/workflows/test.yaml/badge.svg)](https://github.com/jmooring/hvm/actions/workflows/test.yaml)
[![latest release](https://img.shields.io/github/v/release/jmooring/hvm?logo=github)](https://github.com/jmooring/hvm/releases/latest)
[![GitHub sponsors](https://img.shields.io/github/sponsors/jmooring?logo=github&label=sponsors)](https://github.com/sponsors/jmooring)
[![Go Report Card](https://goreportcard.com/badge/github.com/jmooring/hvm)](https://goreportcard.com/report/github.com/jmooring/hvm)

Hugo Version Manager (hvm) is a tool that helps you download, manage, and switch between different versions of the [Hugo] static site generator. You can also use hvm to install Hugo as a standalone application.

![Demonstration](assets/hvm.gif)

You can use hvm to:

- Download and manage multiple versions of Hugo
- Switch between different versions of Hugo in the current directory
- Install Hugo as a standalone application

Supported operating systems:

- Darwin (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64)

## How it works

The `hvm use` command allows you to switch between different versions of Hugo in the current directory. It does this by downloading, extracting, and caching the release asset for your operating system and architecture. It also creates an `.hvm` file in the current directory, which contains the version identifier.

If you do not specify a version of Hugo to use in the current directory, the Hugo executable will be found by searching the PATH environment variable.

To use a different version of Hugo, run the `hvm use` command and select or specify the desired version. To use the Hugo executable in your system PATH, run the `hvm disable` command.

The extracted release assets are cached, so you don't have to download them again each time you switch versions. You can view a list of cached assets, the size of the cache, and the cache location by running the `hvm status` command. You can also clean the cache by running the `hvm clean` command.

The `hvm install` command installs a default version of Hugo to use when version management is disabled in the current directory. This means that you can use hvm as a Hugo installer, even if you don't want to use its version management features.

The `hvm use` and `hvm install` commands typically present a list of available versions. To directly specify a version and bypass this list, include the desired version number or `latest` after the command, as shown in the examples below:

```text
hvm use v0.153.0
hvm use 0.153.0
hvm use latest
```

## Installation

### Step 1 - Install the executable

Download a [prebuilt binary] or install from source (requires Go 1.24.4 or later):

```text
go install github.com/jmooring/hvm@latest
```

### Step 2 - Create an alias

Create an alias for the `hugo` command that overrides the executable path if a
valid `.hvm` file exists in the current directory.

1. Run `hvm gen alias --help` to find the subcommand for the desired shell.
2. Run `hvm gen alias <shell> --help` to see the installation instructions.
3. Run `hvm gen alias <shell>` to generate an alias function for the specified shell.

The `hvm gen alias` command generates alias functions for bash, fish, zsh, and Windows PowerShell.

The alias function displays a brief status message each time it is called, if version management is enabled in the current directory. To disable this message, set the `hvm_show_status` variable to `false` in the alias function.

## Usage

```text
Usage:
  hvm [command]

Available Commands:
  clean       Clean the cache
  completion  Generate the autocompletion script for the specified shell
  config      Display the current configuration
  disable     Disable version management in the current directory
  gen         Generate various files
  help        Help about any command
  install     Install a default version to use when version management is disabled
  remove      Remove the default version
  status      Display the status
  use         Select or specify a version to use in the current directory
  version     Display the hvm version

Flags:
  -h, --help      Display help
  -v, --version   Display the hvm version

Use "hvm [command] --help" for more information about a command
```

## Configuration

To locate the configuration file, run the `hvm config` command. This will print the path to the configuration file to the console. Keys in the configuration file are case-insensitive.

To set configuration values with an environment variable, create an environment variable prefixed with `HVM_`. For example, to set the `sortAscending` configuration value to true:

```text
export HVM_SORTASCENDING=true
```

An environment variable takes precedence over the values set in the configuration file. This means that if you set a configuration value with both an environment variable and in the configuration file, the value in the environment variable will be used.

**gitHubToken** (string)

GitHub limits the number of requests that can be made to its API per hour to 60 for unauthenticated clients. If you exceed this limit, hvm will display a message indicating when the limit will be reset. This is typically within minutes.

If you regularly exceed this limit, you can create a GitHub personal access token with read-only public repository access. With a personal access token, GitHub limits API requests to 5,000 per hour.

**numTagsToDisplay** (int, default `30`)

By default, the `hvm use` and `hvm install` commands display the 30 most recent releases. To display all releases since v0.54.0, set the value to `-1`. Releases before v0.54.0 were not semantically versioned.

**sortAscending** (bool, default `false`)

By default, the `hvm use` and `hvm install` commands display the list of recent releases in descending order. To display the list in ascending order, set this value to `true`.

## Continuous integration and deployment (CD/CD)

For production workflows utilizing CI/CD (e.g., on Cloudflare, GitHub Pages, GitLab Pages, Netlify, Render, or Vercel), the Hugo Version Manager enables a reproducible build environment. The simplest and most reliable approach leverages the `.hvm` file:

1. Check the `.hvm` file into source control
2. Read the `.hvm` file in your build script to determine which Hugo version to install

This guarantees that your site is always built with the specific Hugo version it was developed for.

See this example of a site hosted with GitHub Pages:\
<https://github.com/jmooring/hosting-github-pages-hvm>

## In the news

Discover what others are saying about the Hugo Version Manager.

Date|Link
:--|:--
2025-06-04|<https://harrycresswell.com/writing/hugo-version-manager/>
2025-05-23|<https://ahmedjama.com/blog/2025/05/hugo-version-manager/>
2025-04-07|<https://tsalikis.blog/posts/switching_hugo_versions/>
2025-03-24|<https://navendu.me/posts/hugo-version-manager/>
2023-12-18|<https://perplex.desider.at/news/hugo-version-manager/>
