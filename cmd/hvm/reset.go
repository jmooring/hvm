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
	"strings"

	"github.com/spf13/cobra"
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:    "reset",
	Hidden: true,
	Short:  "Clean cache, disable version management, and remove default version",
	Long: `Disable version management in the current directory, and remove the
configuration and cache directories.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := reset()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}

// reset disables version management in the current directory, and removes the
// configuration and cache directories.
func reset() error {
	fmt.Println("This will reset the configuration and remove the cache directory.")

	var r string
	for {
		fmt.Printf("Are you sure you want to do this? (y/N): ")
		fmt.Scanln(&r)
		if len(r) == 0 || strings.ToLower(string(r[0])) == "n" {
			fmt.Println("Canceled.")
			return nil
		}

		if strings.ToLower(string(r[0])) == "y" {
			err := disable()
			if err != nil {
				return err
			}
			err = os.RemoveAll(App.ConfigDirPath)
			if err != nil {
				return err
			}
			err = os.RemoveAll(App.CacheDirPath)
			if err != nil {
				return err
			}

			return nil
		}
	}
}
