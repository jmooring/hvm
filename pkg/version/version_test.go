package version

import (
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

// TestNewInfo_UsesLdflags verifies that when package-level variables are set
// (as they would be via ldflags), NewInfo prefers those values.
func TestNewInfo_UsesLdflags(t *testing.T) {
	oldV, oldCH, oldBD := Version, CommitHash, BuildDate
	defer func() { Version, CommitHash, BuildDate = oldV, oldCH, oldBD }()

	Version = "1.2.3"
	CommitHash = "abc1234"
	BuildDate = "2026-01-09T12:34:56Z"

	name := "hvm"
	info := NewInfo(name)

	if info.Name != name {
		t.Fatalf("Name: want %q got %q", name, info.Name)
	}
	if info.Version != Version {
		t.Fatalf("Version: want %q got %q", Version, info.Version)
	}
	if info.CommitHash != CommitHash {
		t.Fatalf("CommitHash: want %q got %q", CommitHash, info.CommitHash)
	}
	if info.BuildDate != BuildDate {
		t.Fatalf("BuildDate: want %q got %q", BuildDate, info.BuildDate)
	}

	expected := strings.Join(strings.Fields(
		name+" "+Version+" "+runtime.GOOS+"/"+runtime.GOARCH+" "+CommitHash+" "+BuildDate,
	), " ")
	if got := info.String(); got != expected {
		t.Fatalf("String(): want %q got %q", expected, got)
	}
}

// TestNewInfo_Fallbacks ensures that when package-level variables are empty,
// NewInfo falls back to dynamic values and formatting is sane.
func TestNewInfo_Fallbacks(t *testing.T) {
	oldV, oldCH, oldBD := Version, CommitHash, BuildDate
	defer func() { Version, CommitHash, BuildDate = oldV, oldCH, oldBD }()

	Version, CommitHash, BuildDate = "", "", ""

	name := "hvm"
	info := NewInfo(name)

	if info.Name != name {
		t.Fatalf("Name: want %q got %q", name, info.Name)
	}

	// In local test builds, debug info typically reports "(devel)", so Version == "dev".
	if info.Version == "" {
		t.Fatalf("Version should not be empty; got %q", info.Version)
	}
	// Accept either "dev" or a semantic version.
	if info.Version != "dev" {
		// Rough check: semantic versions start with 'v' and have dots.
		if !strings.HasPrefix(info.Version, "v") || !strings.Contains(info.Version, ".") {
			t.Fatalf("Version should be 'dev' or semantic tag; got %q", info.Version)
		}
	}

	// CommitHash may be empty if no VCS metadata is present; if present, it is <= 7 chars.
	if info.CommitHash != "" && len(info.CommitHash) > 7 {
		t.Fatalf("CommitHash should be empty or <=7 chars; got %q", info.CommitHash)
	}

	// BuildDate should be RFC3339 and close to now.
	if info.BuildDate == "" {
		t.Fatal("BuildDate should not be empty")
	}
	if _, err := time.Parse(time.RFC3339, info.BuildDate); err != nil {
		t.Fatalf("BuildDate is not RFC3339: %v (got %q)", err, info.BuildDate)
	}

	// Ensure String contains the expected components.
	got := info.String()
	for _, want := range []string{name, info.Version, runtime.GOOS + "/" + runtime.GOARCH} {
		if !strings.Contains(got, want) {
			t.Fatalf("String() missing component %q; got %q", want, got)
		}
	}
}

// TestBuildInfoCases validates dynamic helpers by mocking build info via readBuildInfo.
func TestBuildInfoCases(t *testing.T) {
	// Save and restore readBuildInfo.
	oldReader := readBuildInfo
	defer func() { readBuildInfo = oldReader }()

	// Case 1: (devel) → dev
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true
	}
	if got := getDynamicVersion(); got != "dev" {
		t.Fatalf("getDynamicVersion (devel): want dev got %q", got)
	}

	// Case 2: pseudo-version → base-dev
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v0.9.1-0.20230101-abcdef"}}, true
	}
	if got := getDynamicVersion(); got != "v0.9.1-dev" {
		t.Fatalf("getDynamicVersion (pseudo): want v0.9.1-dev got %q", got)
	}

	// Case 3: clean tag → as-is
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v0.9.0"}}, true
	}
	if got := getDynamicVersion(); got != "v0.9.0" {
		t.Fatalf("getDynamicVersion (tag): want v0.9.0 got %q", got)
	}

	// Case 4: commit hash truncation (via settings)
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "123456789abcdef"}}}, true
	}
	if got := getShortCommitHash(); got != "1234567" {
		t.Fatalf("getShortCommitHash: want 1234567 got %q", got)
	}

	// Case 5: commit hash exact 7 chars
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "abcdef7"}}}, true
	}
	if got := getShortCommitHash(); got != "abcdef7" {
		t.Fatalf("getShortCommitHash (7): want abcdef7 got %q", got)
	}
}
