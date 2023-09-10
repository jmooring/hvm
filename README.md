# Hugo Version Manager

[![test status](https://github.com/jmooring/hvm/actions/workflows/test.yaml/badge.svg)](https://github.com/jmooring/hvm/actions/workflows/test.yaml)
[![latest release](https://img.shields.io/github/v/release/jmooring/hvm?logo=github)](https://github.com/jmooring/hvm/releases/latest)
[![GitHub sponsors](https://img.shields.io/github/sponsors/jmooring?logo=github&label=sponsors)](https://github.com/sponsors/jmooring)
[![Go Report Card](https://goreportcard.com/badge/github.com/jmooring/hvm)](https://goreportcard.com/report/github.com/jmooring/hvm)

Hugo Version Manager (hvm) allows you download multiple versions of the extended edition of the [Hugo] static site generator, and to specify which version to use in the current directory.

![Demonstration](demo/hvm.gif)

If you do not specify a version to use in the current directory, the path to the hugo executable is determined by the `PATH` environment variable.

Supported operating systems:

- Darwin (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64)

Use hvm in your development environment if you need to frequently switch between versions.

## How it works

When you run `hvm use` it displays a list of recent Hugo releases, prompting you to select a version to use in the current directory. It then downloads, extracts, and caches the release asset for your operating system and architecture, writing an .hvm file in the current directory. The .hvm file contains the path to the cached hugo executable.

In the second step of the [installation instructions] below, you will create an alias for the `hugo` command that obtains the path to the cached hugo executable from the .hvm file.

To use a different version of Hugo in the same directory, run `hvm use` again. To use the Hugo executable in your system `PATH`, run `hvm disable`.

Extracted release assets are cached. Run `hvm status` to display a list of cached assets, the size of the cache, and the cache location. Run `hvm clean` to clean the cache.

The `hvm install` command installs a default version to use when version management is disabled in the current directory. You can use hvm as an installer, even if you skip the second step of the installation instructions below.

## Installation

### Step 1 - Install the executable

Download a [prebuilt binary] or build from source with [Go] 1.21 or later:

```text
go install github.com/jmooring/hvm@latest
```

### Step 2 - Create an alias

Create an alias for the `hugo` command that overrides the executable path if a
valid .hvm file exists in the current directory.

<details>
<summary>Darwin: add this function to $HOME/.zshrc</summary>

```zsh
# Hugo Version Manager: override path to the hugo executable.
hugo() {
  hvm_common_msg="Run 'hvm use' to fix or 'hvm disable' to disable version management."
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
  else
    hugo_bin=$(which hugo)
  fi
  "${hugo_bin}" "$@"
}
```

</details>

<details>
<summary>Linux: add this function to $HOME/.bashrc</summary>

```bash
# Hugo Version Manager: override path to the hugo executable.
hugo() {
  hvm_common_msg="Run 'hvm use' to fix or 'hvm disable' to disable version management."
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
  else
    hugo_bin=$(which hugo)
  fi
  "${hugo_bin}" "$@"
}
```

</details>

<details>
<summary>Windows: place this batch file in your PATH</summary>

```bash
TBD
```

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

By default, `hvm use` and `hvm install` display the 30 most recent releases. To change the number of releases displayed:

- Set `numtagstodisplay` in the configuration file (run `hvm config` to display the path)
- Set the `HVM_NUMTAGSTODISPLAY` environment variable

Display all releases since v0.54.0 by setting this value to `-1`. Releases before v0.54.0 were not semantically versioned.

## Rate limiting

GitHub imposes a rate limit on all API clients, limiting unauthenticated clients to 60 requests per hour. If you exceed the limit, hvm displays a message indicating when the limit will be reset, typically within minutes.

[go]: https://go.dev/doc/install
[hugo]: https://github.com/gohugoio/hugo/#readme
[installation instructions]: #installation
[prebuilt binary]: https://github.com/jmooring/hvm/releases/latest
