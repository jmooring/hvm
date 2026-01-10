/*
Copyright © 2026 Joe Mooring <joe.mooring@veriphor.com>

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

// Package version provides build and version information.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

// readBuildInfo is a test seam that allows mocking build info in unit tests.
// In production, it points to debug.ReadBuildInfo.
var readBuildInfo = debug.ReadBuildInfo

// These variables are optionally set at build time via ldflags.
// For example:
//
//	-X github.com/jmooring/hvm/pkg/version.Version=v1.2.3
//	-X github.com/jmooring/hvm/pkg/version.CommitHash=abcdef7
//	-X github.com/jmooring/hvm/pkg/version.BuildDate=2026-01-09T00:00:00Z
var (
	Version    string
	CommitHash string
	BuildDate  string
)

// Info contains build and version metadata.
type Info struct {
	Name       string
	Version    string
	CommitHash string
	BuildDate  string
}

// NewInfo creates a new Info instance with current build information.
func NewInfo(name string) Info {
	v := Version
	if v == "" {
		v = getDynamicVersion()
	}

	ch := CommitHash
	if ch == "" {
		ch = getShortCommitHash()
	}

	bd := BuildDate
	if bd == "" {
		bd = time.Now().UTC().Format(time.RFC3339)
	}

	return Info{
		Name:       name,
		Version:    v,
		CommitHash: ch,
		BuildDate:  bd,
	}
}

// String returns a formatted version string with all metadata.
func (i Info) String() string {
	v := fmt.Sprintf("%s %s %s/%s %s %s", i.Name, i.Version, runtime.GOOS, runtime.GOARCH, i.CommitHash, i.BuildDate)
	return strings.Join(strings.Fields(v), " ")
}

// getShortCommitHash returns the first seven characters of the VCS revision
// (Git commit SHA) from the embedded build information. It returns an empty
// string if the build information is unavailable or the binary was not
// built within a version control system.
func getShortCommitHash() string {
	if info, ok := readBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				hash := setting.Value
				if len(hash) > 7 {
					return hash[:7]
				}
				return hash
			}
		}
	}
	return ""
}

// getDynamicVersion returns the application version from the embedded build
// information. For tagged releases, it returns the tag (e.g., "v0.9.0"). For
// development builds, it truncates the Go-generated pseudo-version to its
// base and suffixes it with "-dev" (e.g., "v0.9.1-dev"). It returns "dev"
// if the build information is unavailable.
func getDynamicVersion() string {
	info, ok := readBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "dev"
	}

	v := info.Main.Version // e.g., "v0.9.1-0.20230101-abcdef" (Pseudo-version)

	// A pseudo-version always contains a dash and a timestamp-like string.
	// Examples: v0.9.1-0.2023... or v0.9.1-dev.0.2023...
	if strings.Contains(v, "-0.") || strings.Contains(v, "-dev.") {
		// Split at the first dash to get just the "v0.9.1" part
		parts := strings.Split(v, "-")
		return parts[0] + "-dev"
	}

	// If it's a clean tag (v0.9.0), it returns as-is.
	return v
}
