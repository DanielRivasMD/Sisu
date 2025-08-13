/*
Copyright © 2025 Daniel Rivas <danielrivasmd@gmail.com>

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

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var rootCmd = &cobra.Command{
	Use:     "sisu",
	Long:    helpRoot,
	Example: exampleRoot,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	verbose bool
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose diagnostics")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpRoot = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"",
)
var exampleRoot = formatExample(
	"sisu",
	[]string{"help"},
)

////////////////////////////////////////////////////////////////////////////////////////////////////

// formatHelp produces the “help” header + description.
//
//	author: name, e.g. "Daniel Rivas"
//	email:  email, e.g. "danielrivasmd@gmail.com"
//	desc:   the multi‐line description, "\n"-separated.
func formatHelp(author, email, desc string) string {
	header := chalk.Bold.TextStyle(
		chalk.Green.Color(author+" "),
	) +
		chalk.Dim.TextStyle(
			chalk.Italic.TextStyle("<"+email+">"),
		)

	// prefix two newlines to your desc, chalk it cyan + dim it
	body := "\n\n" + desc
	return header + chalk.Dim.TextStyle(chalk.Cyan.Color(body))
}

// formatExample builds a multi‐line example block
// each usage is a slice of “tokens”: [ command, flagOrArg, flagOrArg, ... ].
//
//	app:    your binary name, e.g. "lilith"
//	usages: one or more usages—each becomes its own line.
func formatExample(app string, usages ...[]string) string {
	var b strings.Builder

	for i, usage := range usages {
		if len(usage) == 0 {
			continue
		}

		// first token is the subcommand
		b.WriteString(
			chalk.White.Color(app) + " " +
				chalk.White.Color(chalk.Bold.TextStyle(usage[0])),
		)

		// remaining tokens are either flags (--foo) or args
		for _, tok := range usage[1:] {
			switch {
			case strings.HasPrefix(tok, "--"):
				b.WriteString(" " + chalk.Italic.TextStyle(chalk.White.Color(tok)))
			default:
				b.WriteString(" " + chalk.Dim.TextStyle(chalk.Italic.TextStyle(tok)))
			}
		}

		if i < len(usages)-1 {
			b.WriteRune('\n')
		}
	}

	return b.String()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
