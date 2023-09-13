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

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"uninstall"},
	Short:   "Remove the default version",
	Long:    "Removes the default version used when version management is disabled.",
	Run: func(cmd *cobra.Command, args []string) {
		err := remove()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func remove() error {
	exists, err := helpers.Exists(App.DefaultDirPath)
	if err != nil {
		return err
	}
	if exists {
		err = os.RemoveAll(App.DefaultDirPath)
		if err != nil {
			return err
		}
		fmt.Println("Default version removed.")
	} else {
		fmt.Println("Nothing to remove.")
	}

	return nil
}
