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

package cmd

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestCommand(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testscripts",
		Setup: func(env *testscript.Env) error {
			env.Setenv("HVM_GITHUBTOKEN", os.Getenv("HVM_GITHUBTOKEN"))
			return nil
		},
	})
}
func TestCommandConfig(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testscripts/config",
		Setup: func(env *testscript.Env) error {
			env.Setenv("HVM_GITHUBTOKEN", os.Getenv("HVM_GITHUBTOKEN"))
			return nil
		},
	})
}
func TestCommandInstall(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testscripts/install",
		Setup: func(env *testscript.Env) error {
			env.Setenv("HVM_GITHUBTOKEN", os.Getenv("HVM_GITHUBTOKEN"))
			return nil
		},
	})
}
func TestCommandRemove(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testscripts/remove",
		Setup: func(env *testscript.Env) error {
			env.Setenv("HVM_GITHUBTOKEN", os.Getenv("HVM_GITHUBTOKEN"))
			return nil
		},
	})
}
func TestCommandStatus(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testscripts/status",
		Setup: func(env *testscript.Env) error {
			env.Setenv("HVM_GITHUBTOKEN", os.Getenv("HVM_GITHUBTOKEN"))
			return nil
		},
	})
}
func TestCommandUse(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testscripts/use",
		Setup: func(env *testscript.Env) error {
			env.Setenv("HVM_GITHUBTOKEN", os.Getenv("HVM_GITHUBTOKEN"))
			return nil
		},
	})
}
