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

	"github.com/jmooring/hvm/pkg/helpers"
	"github.com/spf13/cobra"
)

// disableCmd represents the disable command
var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable version management in the current directory",
	Long:  "Disable version management in the current directory.",
	Run: func(cmd *cobra.Command, args []string) {
		err := disable()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(disableCmd)
}

// disable disables version management in the current directory.
func disable() error {
	exists, err := helpers.Exists(App.DotFilePath)
	if err != nil {
		return err
	}

	if exists {
		err := os.Remove(App.DotFilePath)
		if err != nil {
			return err
		}
	}

	fmt.Println("Version management has been disabled in the current directory.")

	return nil
}
