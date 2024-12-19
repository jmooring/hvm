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
		Dir:   "testscripts",
		Setup: setup,
	})
}

func TestCommandConfig(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testscripts/config",
		Setup: setup,
	})
}

func TestCommandInstall(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testscripts/install",
		Setup: setup,
	})
}

func TestCommandRemove(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testscripts/remove",
		Setup: setup,
	})
}

func TestCommandStatus(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testscripts/status",
		Setup: setup,
	})
}

func TestCommandUse(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testscripts/use",
		Setup: setup,
	})
}

func setup(env *testscript.Env) error {
	env.Setenv("HVM_GITHUBTOKEN", os.Getenv("HVM_GITHUBTOKEN"))
	return nil
}
