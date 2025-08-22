////////////////////////////////////////////////////////////////////////////////////////////////////

package cmd

////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"github.com/ttacon/chalk"
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

////////////////////////////////////////////////////////////////////////////////////////////////////

var helpRoot = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"",
)

var helpMigrate = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"Runs every pending `.up.sql` migration against SQLite database"+
		"If the database file does not exist yet, it will be created automatically",
)

var helpTrack = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"",
)

var helpTask = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"Track the high-level routines or goals",
)

var helpMilestone = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"Used for incentives, streaks, or mastery checkpoints",
)

////////////////////////////////////////////////////////////////////////////////////////////////////
