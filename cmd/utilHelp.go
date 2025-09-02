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
	"Runs every pending `.up.sql` migration against SQLite database\n"+
		"If the database file does not exist yet, it will be created automatically",
)

var helpTask = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"Track the high-level routines or goals",
)

var helpTaskArchived = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"Mark tasks as archived by setting the 'archived' field to true",
)

var helpMilestone = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"Used for incentives, streaks, or mastery checkpoints",
)

var helpExport = formatHelp(
	"Daniel Rivas",
	"<danielrivasmd@gmail.com>",
	"Export specified tables from the database into CSV files in the current working directory\n"+
		"Supported tables: tasks, sessions, milestones, reviews, coach, calendar\n"+
		"Use --all to export every supported table",
)

////////////////////////////////////////////////////////////////////////////////////////////////////
