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

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display the current configuration",
	Long:  "Display the current configuration and the path to the configuration file.",
	Run: func(cmd *cobra.Command, args []string) {
		err := displayConfig()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

// displayConfig displays the current configuration and the path to the
// configuration file.
func displayConfig() error {
	t, err := toml.Marshal(Config)
	cobra.CheckErr(err)

	fmt.Println(string(t))
	fmt.Println("Configuration file:", App.ConfigFilePath)

	return nil
}
