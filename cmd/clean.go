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
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean the cache",
	Long: `Clean the cache, excluding the version installed with the "install"
command.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := clean()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}

// clean cleans the cache, excluding the version installed with the "install"
// command.
func clean() error {
	cacheSize, err := getCacheSize()
	if err != nil {
		return err
	}

	if cacheSize == 0 {
		fmt.Println("The cache is already empty.")
		return nil
	}

	fmt.Println("This will delete cached versions of the Hugo executable.")

	var r string
	for {
		fmt.Printf("Are you sure you want to clean the cache? (y/N): ")
		fmt.Scanln(&r)
		if r == "" || strings.EqualFold(string(r[0]), "n") {
			fmt.Println("Canceled.")
			return nil
		}

		if strings.EqualFold(string(r[0]), "y") {
			d, err := os.ReadDir(App.CacheDirPath)
			if err != nil {
				return err
			}

			for _, f := range d {
				if f.Name() != App.DefaultDirName {
					err := os.RemoveAll(filepath.Join(App.CacheDirPath, f.Name()))
					if err != nil {
						return err
					}
				}
			}
			fmt.Println("Cache cleaned.")

			return nil
		}
	}
}
