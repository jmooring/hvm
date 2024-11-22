package cmd

import (
	"fmt"
	"slices"
	"testing"
)

func newMockRepo(reverse bool) *repository {
	r := repository{
		tags:      []string{"v1.2.3", "v1.2.2", "v1.2.1", "v1.2.0", "v1.1.2", "v1.1.1", "v1.1.0", "v1.0.0", "v0.2.2", "v0.2.1", "v0.2.0", "v0.1.2", "v0.1.1", "v0.1.0"},
		latestTag: "v1.2.3",
	}
	if reverse {
		slices.Reverse(r.tags)
	}
	return &r
}

var testCases = []struct {
	given string
	want  string
}{
	// invalid strings
	{"", ""},
	{"abc", ""},
	{"a.b", ""},
	{"v1.", ""},
	{"v1.2.", ""},
	{"v1.2.3.", ""},
	{"v1.2.3.4", ""},
	{"1.", ""},
	{"1.2.", ""},
	{"1.2.3.", ""},
	{"1.2.3.4", ""},
	// valid strings
	{"v", "v1.2.3"},
	{"v1", "v1.2.3"},
	{"v1.2", "v1.2.3"},
	{"v1.2.3", "v1.2.3"},
	{"1", "v1.2.3"},
	{"1.2", "v1.2.3"},
	{"1.2.3", "v1.2.3"},
	// major only use lookup
	{"0", "v0.2.2"},
	{"v0", "v0.2.2"},
	{"1", "v1.2.3"},
	{"v1", "v1.2.3"},
	{"v1", "v1.2.3"},
	// major and minor use lookup
	{"0.1", "v0.1.2"},
	{"v0.1", "v0.1.2"},
	{"1.1", "v1.1.2"},
	{"v1.1", "v1.1.2"},
	{"0.2", "v0.2.2"},
	{"v0.2", "v0.2.2"},
	// latest-major, minor and lookup)
	{".1", "v1.1.2"},
	{"v.1", "v1.1.2"},
	{".0", "v1.0.0"},
	{"v.0", "v1.0.0"},
	// Specific major taken from latestTag (no lookup)
	{".2.1", "v1.2.1"},
	{"v.2.1", "v1.2.1"},
	{".2.1", "v1.2.1"},
	{"v.2.1", "v1.2.1"},
}

func TestGetSpecificTagDescending(t *testing.T) {
	// mock needed elements and sort config
	Config.SortAscending = false
	repo := newMockRepo(Config.SortAscending)
	asset := newAsset()
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("[%s] ShouldBe [%s]", tc.given, tc.want), func(t *testing.T) {
			asset.tag = ""
			err := repo.getSpecificTag(asset, tc.given)
			if asset.tag != tc.want {
				t.Fatalf("given(%s) -> calculated(%s) -> want(%s) : %s", tc.given, asset.tag, tc.want, err)
			}
		})
	}
}

func TestGetSpecificTagAscending(t *testing.T) {
	// mock needed elements and sort config
	Config.SortAscending = true
	repo := newMockRepo(Config.SortAscending)
	asset := newAsset()
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("[%s] ShouldBe [%s]", tc.given, tc.want), func(t *testing.T) {
			asset.tag = ""
			err := repo.getSpecificTag(asset, tc.given)
			if asset.tag != tc.want {
				t.Errorf("given(%s) -> calculated(%s) -> want(%s) : %s", tc.given, asset.tag, tc.want, err)
			}
		})
	}
}
