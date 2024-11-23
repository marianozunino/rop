/*
Copyright © 2024 Mariano Zunino <marianoz@posteo.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"strings"

	"github.com/marianozunino/selfupdater"
	"github.com/spf13/cobra"
)

func NewUpdateCmd() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update the rop tool to the latest available version.",
		Long: `Automatically update the rop tool to the latest available version from the official source.
This command checks for updates, downloads the new version, and replaces the current executable with the updated one.`,
		Example: `rop update`,
		Run: func(cmd *cobra.Command, args []string) {
			currentVersionStr := strings.TrimPrefix(Version, "v")
			updater := selfupdater.NewUpdater("marianozunino", "rop", "rop", currentVersionStr)
			updater.Update()
		},
	}

	return updateCmd
}

func init() {
	rootCmd.AddCommand(NewUpdateCmd())
}
