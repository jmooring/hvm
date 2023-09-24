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
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed aliases/powershell.ps1
var powershellScript string

// powershellCmd represents the powershell command
var powershellCmd = &cobra.Command{
	Use:   "powershell",
	Short: "Generate an alias function for powershell",
	Long: `Generate an alias function for Windows PowerShell.

Instructions:

1. Create a PowerShell profile: New-Item -Path $profile -ItemType File
2. Open the file with Notepad: notepad $profile
3. Add the output from this command to the file you just opened
4. Save the file and exit Notepad
5. Refresh the current session: . $profile

The alias function displays a brief status message each time it is called, if
version management is enabled in the current directory. To disable this
message, set the "hvm_show_status" variable to "$false" in the alias function.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(powershellScript)
	},
}

func init() {
	aliasCmd.AddCommand(powershellCmd)
}
