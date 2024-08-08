/*
Copyright © 2024 timoxa0

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"os"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var version string
var logger pterm.Logger
var pbar pterm.ProgressbarPrinter
var spinner pterm.SpinnerPrinter

var verbose bool

var rootCmd = &cobra.Command{
	Use:     "lon-tool",
	Short:   "A tool for installing linux on nabu and managing linux images",
	Long:    "A tool for installing linux on nabu and managing linux images",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			logger = *pterm.DefaultLogger.WithLevel(pterm.LogLevelDebug).WithTime(false)
		} else {
			logger = *pterm.DefaultLogger.WithLevel(pterm.LogLevelInfo).WithTime(false)
		}
		pbarStyle := pterm.NewStyle(pterm.FgLightGreen, pterm.Bold)
		pbarFillStyle := pterm.NewStyle(pterm.FgLightRed, pterm.Bold)
		pbarTitleStyle := pterm.NewStyle(pterm.FgLightBlue, pterm.Bold)
		pbar = *pterm.DefaultProgressbar.
			WithShowElapsedTime(false).
			WithRemoveWhenDone(false).
			WithShowCount(false).
			WithBarFiller(pbarFillStyle.Sprint("─")).
			WithLastCharacter("─").
			WithBarCharacter("─").
			WithTitleStyle(pbarTitleStyle).
			WithBarStyle(pbarStyle)
		spinner = *pterm.DefaultSpinner.
			WithRemoveWhenDone(true).
			WithSequence("─", "\\", "|", "/").
			WithStyle(pbarTitleStyle).
			WithDelay(time.Millisecond * 100)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "enable debug output")
}
