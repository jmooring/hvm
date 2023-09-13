# Hugo Version Manager

[![test status](https://github.com/jmooring/hvm/actions/workflows/test.yaml/badge.svg)](https://github.com/jmooring/hvm/actions/workflows/test.yaml)
[![latest release](https://img.shields.io/github/v/release/jmooring/hvm?logo=github)](https://github.com/jmooring/hvm/releases/latest)
[![GitHub sponsors](https://img.shields.io/github/sponsors/jmooring?logo=github&label=sponsors)](https://github.com/sponsors/jmooring)
[![Go Report Card](https://goreportcard.com/badge/github.com/jmooring/hvm)](https://goreportcard.com/report/github.com/jmooring/hvm)

Hugo Version Manager (hvm) allows you to download and manage multiple versions of the extended edition of the [Hugo] static site generator. You can use hvm to specify which version of Hugo to use in the current directory.

![Demonstration](assets/hvm.gif)

If you do not specify a version of Hugo to use in the current directory, the Hugo executable will be found by searching the `PATH` environment variable.

Supported operating systems:

- Darwin (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64)

Use hvm in your development environment if you need to frequently switch between versions.

## How it works

When you run the `hvm use` command, it will display a list of recent Hugo releases. You can then select a version to use in the current directory. The hvm application will download, extract, and cache the release asset for your operating system and architecture. It will also create an `.hvm` file in the current directory. This file contains the path to the cached Hugo executable.

In the second step of the installation instructions, you will create an alias for the `hugo` command. This alias will obtain the path to the cached Hugo executable from the `.hvm` file.

To use a different version of Hugo in the same directory, run the `hvm use` command again. To use the Hugo executable in your system PATH, run the `hvm disable` command.

The extracted release assets are cached. You can run the `hvm status` command to display a list of cached assets, the size of the cache, and the cache location. You can also run the `hvm clean` command to clean the cache.

The `hvm install` command installs a default version to use when version management is disabled in the current directory. You can use hvm as an installer, even if you skip the second step of the installation instructions.

## Installation

### Step 1 - Install the executable

Download a [prebuilt binary] or build from source with [Go] 1.21 or later:

```text
go install github.com/jmooring/hvm@latest
```

### Step 2 - Create an alias

Create an alias for the `hugo` command that overrides the executable path if a
valid `.hvm` file exists in the current directory.

<details>
<summary>Darwin</summary>
Add this function to $HOME/.zshrc

```zsh
# Hugo Version Manager: override path to the hugo executable.
hugo() {
  hvm_common_msg="Run 'hvm use' to fix or 'hvm disable' to disable version management."
  hvm_show_status=true
  if [ -f ".hvm" ]; then
    hugo_bin=$(cat ".hvm" 2> /dev/null)
    if ! echo "${hugo_bin}" | grep -q "hugo$"; then
      >&2 printf "The .hvm file in this directory is invalid.\\n"
      >&2 printf "%s\\n" "${hvm_common_msg}"
      return 1
    fi
    if [ ! -f "${hugo_bin}" ]; then
      >&2 printf "Unable to find %s.\\n" "${hugo_bin}"
      >&2 printf "%s\\n" "${hvm_common_msg}"
      return 1
    fi
    if [ "${hvm_show_status}" == true ]; then
      >&2 printf "Hugo version management is enabled in this directory.\\n"
      >&2 printf "Run 'hvm status' for details, or 'hvm disable' to disable.\\n\\n"
    fi
  else
    hugo_bin=$(which hugo)
  fi
  "${hugo_bin}" "$@"
}
```

</details>

<details>
<summary>Linux</summary>
Add this function to $HOME/.bashrc

```bash
# Hugo Version Manager: override path to the hugo executable.
hugo() {
  hvm_common_msg="Run 'hvm use' to fix or 'hvm disable' to disable version management."
  hvm_show_status=true
  if [ -f ".hvm" ]; then
    hugo_bin=$(cat ".hvm" 2> /dev/null)
    if ! echo "${hugo_bin}" | grep -q "hugo$"; then
      >&2 printf "The .hvm file in this directory is invalid.\\n"
      >&2 printf "%s\\n" "${hvm_common_msg}"
      return 1
    fi
    if [ ! -f "${hugo_bin}" ]; then
      >&2 printf "Unable to find %s.\\n" "${hugo_bin}"
      >&2 printf "%s\\n" "${hvm_common_msg}"
      return 1
    fi
    if [ "${hvm_show_status}" == true ]; then
      >&2 printf "Hugo version management is enabled in this directory.\\n"
      >&2 printf "Run 'hvm status' for details, or 'hvm disable' to disable.\\n\\n"
    fi
  else
    hugo_bin=$(which hugo)
  fi
  "${hugo_bin}" "$@"
}
```

</details>

<details>
<summary>Windows</summary>

TBD

</details>

## Usage

```text
Usage:
  hvm [command]

Available Commands:
  clean       Clean the cache
  completion  Generate the autocompletion script for the specified shell
  config      Display the current configuration
  disable     Disable version management in the current directory
  help        Help about any command
  install     Install a default version to use when version management is disabled
  remove      Remove the default version
  status      Display the status
  use         Select a version to use in the current directory
  version     Display the hvm version

Flags:
  -h, --help      Display help
  -v, --version   Display the hvm version

Use "hvm [command] --help" for more information about a command.
```

## Configuration

### Number of releases to display

By default, the `hvm use` and `hvm install` commands display the 30 most recent releases. To change the number of releases displayed, you can do one of the following:

- Set the `numtagstodisplay` option in the hvm configuration file. You can find the path to the configuration file by running the `hvm config` command.
- Set the `HVM_NUMTAGSTODISPLAY` environment variable.

To display all releases since v0.54.0, set the value to `-1`. Releases before v0.54.0 were not semantically versioned.

### GitHub personal access token

GitHub limits the number of requests that can be made to its API per hour to 60 for unauthenticated clients. If you exceed this limit, hvm will display a message indicating when the limit will be reset. This is typically within minutes.

If you regularly exceed this limit, you can create a GitHub personal access token with read-only public repository access. Then, you can do either of the following:

- Set the `githubtoken` option in the hvm configuration file. You can find the path to the configuration file by running the hvm config command.
- Set the `HVM_GITHUBTOKEN` environment variable.

With a personal access token, GitHub limits API requests to 5,000 per hour.

[go]: https://go.dev/doc/install
[hugo]: https://github.com/gohugoio/hugo/#readme
[installation instructions]: #installation
[prebuilt binary]: https://github.com/jmooring/hvm/releases/latest
