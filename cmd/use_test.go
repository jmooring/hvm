/*
Copyright © 2024 Joe Mooring <joe.mooring@veriphor.com>

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

package cmd

import (
	"fmt"
	"testing"
)

func newMockRepo() *repository {
	r := repository{
		tags:      []string{"v1.2.3", "v1.2.2", "v1.2.1"},
		latestTag: "v1.2.3",
	}
	return &r
}

var testCases = []struct {
	given string
	want  string
}{
	// valid existing version
	{"latest", "v1.2.3"},
	{"v1.2.3", "v1.2.3"},
	{"1.2.3", "v1.2.3"},
	{"v1.2.2", "v1.2.2"},
	{"1.2.2", "v1.2.2"},
	{"v1.2.1", "v1.2.1"},
	{"1.2.1", "v1.2.1"},
	// valid missing version
	{"v1.2.0", ""},
	{"1.2.0", ""},
	// invalid strings
	{"", ""},
	{"late", ""},
	{"a.b.c", ""},
	{"1", ""},
	{"1.2.", ""},
	{"1.2.", ""},
	{"1.2.3.", ""},
	{"1.2.3.4", ""},
}

func TestGetTagFromString(t *testing.T) {
	// mock needed elements and sort config
	Config.SortAscending = false
	repo := newMockRepo()
	asset := newAsset()
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("[%s] ShouldBe [%s]", tc.given, tc.want), func(t *testing.T) {
			asset.tag = ""
			err := repo.getTagFromString(asset, tc.given)
			if asset.tag != tc.want {
				t.Fatalf("given(%s) -> calculated(%s) -> want(%s) : %s", tc.given, asset.tag, tc.want, err)
			}
		})
	}
}
