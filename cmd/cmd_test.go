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
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// TestCommand runs testscripts for the main command set.
func TestCommand(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testscripts",
		Setup: setup,
	})
}

// TestCommandConfig runs testscripts for the config command.
func TestCommandConfig(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testscripts/config",
		Setup: setup,
	})
}

// TestCommandInstall runs testscripts for the install command.
func TestCommandInstall(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testscripts/install",
		Setup: setup,
	})
}

// TestCommandRemove runs testscripts for the remove command.
func TestCommandRemove(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testscripts/remove",
		Setup: setup,
	})
}

// TestCommandStatus runs testscripts for the status command.
func TestCommandStatus(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testscripts/status",
		Setup: setup,
	})
}

// TestCommandUse runs testscripts for the use command.
func TestCommandUse(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:   "testscripts/use",
		Setup: setup,
	})
}

// setup initializes the test environment with necessary environment variables.
func setup(env *testscript.Env) error {
	env.Setenv("HVM_GITHUBTOKEN", os.Getenv("HVM_GITHUBTOKEN"))
	switch runtime.GOOS {
	case "darwin":
		env.Setenv("HOME", "home")
	case "windows":
		// User cache and config dirs: we use os.UserCacheDir and os.UserCongfigDir
		env.Setenv("LocalAppData", "cache")
		env.Setenv("AppData", "config")
	case "linux":
		// User cache and config dirs: we use os.UserCacheDir and os.UserCongfigDir
		env.Setenv("XDG_CACHE_HOME", env.Getenv("WORK")+"/cache")
		env.Setenv("XDG_CONFIG_HOME", env.Getenv("WORK")+"/config")
	default:
		return fmt.Errorf("unsupported operating system")
	}
	return nil
}
