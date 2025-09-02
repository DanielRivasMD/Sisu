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
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/DanielRivasMD/horus"
	"github.com/aarondl/null/v8"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"

	"github.com/DanielRivasMD/Sisu/db"
)

////////////////////////////////////////////////////////////////////////////////////////////////////

var rootCmd = &cobra.Command{
	Use:     "sisu",
	Long:    helpRoot,
	Example: exampleRoot,
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func Execute() {
	horus.CheckErr(rootCmd.Execute())
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// date format
const DateYMD = "2006-01-02"

var (
	verbose bool
	dbPath  string // populated by the --db flag
)

////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose diagnostics")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "sisu.db", "path to sqlite database")
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func persistentPreRun(cmd *cobra.Command, args []string) {
	if _, err := db.InitDB(dbPath); err != nil {
		log.Fatalf("init DB: %v", err)
	}
}

func persistentPostRun(cmd *cobra.Command, args []string) {
	if db.Conn != nil {
		_ = db.Conn.Close()
		db.Conn = nil
	}
}

// EnsureDB open the DB using the same logic as PersistentPreRun
// reuse that here so __complete has a live connection
func EnsureDB() error {
	if db.Conn != nil {
		return nil
	}
	_, err := db.InitDB(dbPath)
	return err
}

////////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////////////////////////
// Reusable parsers, validators, assigners, and tiny field builders
// for TUI forms across commands, aligned with your current SQL schema.
//
// Schema highlights (Go types you’ll likely get from SQLBoiler):
// - tasks:     Name string (required), Tag/Description null.String, Target/Start null.Time, Archived null.Bool or bool
// - sessions:  Task int64, Date time.Time or null.Time (depends on NULL), Mins/Feedback null.Int64, Notes null.String
// - milestones:Task int64, Type/Message null.String, Value null.Int64, Done time.Time or null.Time (depends on NULL)
// - reviews:   Task int64, Week null.Int64, Summary null.String
// - coach:     Trigger/Content string (required), Date time.Time or null.Time (depends on NULL)
// - calendar:  Date time.Time or null.Time (depends on NULL), Note string (required)
//
// Adjust nullable vs non-null helpers (FDate vs FOptDate) to match your generated models.
////////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////////////////////////
// Validators
////////////////////////////////////////////////////////////////////////////////////////////////////

// VRequired enforces non-empty input.
func VRequired(label string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s is required", label)
		}
		return nil
	}
}

// VBool validates a free-form boolean (true/false/1/0/yes/no).
func VBool(label string) func(string) error {
	return func(s string) error {
		s = strings.TrimSpace(strings.ToLower(s))
		switch s {
		case "", "0", "1", "true", "t", "false", "f", "yes", "y", "no", "n":
			return nil
		default:
			return fmt.Errorf("%s must be a boolean (true/false/yes/no/1/0)", label)
		}
	}
}

// VIntRange validates optional integer input within [min, max].
// Blank passes (use with optional fields). Pair with ParseOptInt64.
func VIntRange(min, max int64) func(string) error {
	return func(s string) error {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return errors.New("must be a number")
		}
		if n < min || n > max {
			return fmt.Errorf("must be between %d and %d", min, max)
		}
		return nil
	}
}

// VDate enforces a required YYYY-MM-DD date (for NOT NULL time.Time).
func VDate(label string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s is required", label)
		}
		if _, err := time.Parse(DateYMD, s); err != nil {
			return errors.New("date must be YYYY-MM-DD")
		}
		return nil
	}
}

// VDateOptional validates optional YYYY-MM-DD date (for null.Time).
func VDateOptional() func(string) error {
	return func(s string) error {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		if _, err := time.Parse(DateYMD, s); err != nil {
			return errors.New("date must be YYYY-MM-DD")
		}
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Initials
////////////////////////////////////////////////////////////////////////////////////////////////////

// Null helpers → initial string for form fields

func OptInt64Initial(v null.Int64) string {
	if v.Valid {
		return strconv.FormatInt(v.Int64, 10)
	}
	return ""
}

func OptTimeInitial(v null.Time, layout string) string {
	if v.Valid {
		return v.Time.Format(layout)
	}
	return ""
}

func OptStringInitial(v null.String) string {
	// null.String.String is already "" when invalid, but keep symmetry
	if v.Valid {
		return v.String
	}
	return ""
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Parsers
////////////////////////////////////////////////////////////////////////////////////////////////////

// ParseInt64 parses a required int64.
func ParseInt64(s string) (any, error) {
	return strconv.ParseInt(s, 10, 64)
}

// ParseOptInt64 parses an optional int64 into null.Int64.
func ParseOptInt64(s string) (any, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return null.Int64{}, nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, err
	}
	return null.Int64From(n), nil
}

// ParseDate parses a required date (YYYY-MM-DD) into time.Time.
func ParseDate(s string) (any, error) {
	return time.Parse(DateYMD, s)
}

// ParseOptDate parses an optional date (YYYY-MM-DD) into null.Time.
func ParseOptDate(s string) (any, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return null.Time{}, nil
	}
	t, err := time.Parse(DateYMD, s)
	if err != nil {
		return nil, err
	}
	return null.TimeFrom(t), nil
}

// ParseOptString converts free text into null.String (blank → Valid:true, String:"").
func ParseOptString(s string) (any, error) {
	return null.StringFrom(s), nil
}

// ParseNonEmpty returns the string if non-empty; else error.
func ParseNonEmpty(label string) func(string) (any, error) {
	return func(s string) (any, error) {
		if strings.TrimSpace(s) == "" {
			return nil, fmt.Errorf("%s cannot be blank", label)
		}
		return s, nil
	}
}

// ParseBool supports true/false/1/0/yes/no; returns bool.
func ParseBool(s string) (any, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "t", "yes", "y":
		return true, nil
	case "0", "false", "f", "no", "n":
		return false, nil
	default:
		return nil, fmt.Errorf("invalid bool: %q", s)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Assigners (by struct field name)
////////////////////////////////////////////////////////////////////////////////////////////////////

// AssignDate sets a time.Time into a named field (for NOT NULL date/time columns).
func AssignDate(field string, holder any, v any) {
	reflect.ValueOf(holder).Elem().FieldByName(field).
		Set(reflect.ValueOf(v.(time.Time)))
}

// AssignNullableDate sets a null.Time into a named field (for nullable dates).
func AssignNullableDate(field string, holder any, v any) {
	reflect.ValueOf(holder).Elem().FieldByName(field).
		Set(reflect.ValueOf(v)) // v must be null.Time
}

// AssignInt64 sets an int64 into a named field (e.g., FK task).
func AssignInt64(field string, holder any, v any) {
	reflect.ValueOf(holder).Elem().FieldByName(field).
		SetInt(v.(int64))
}

// AssignString sets a plain string into a named field.
func AssignString(field string, holder any, v any) {
	reflect.ValueOf(holder).Elem().FieldByName(field).
		SetString(v.(string))
}

// AssignBool sets a bool into a named field.
func AssignBool(field string, holder any, v any) {
	reflect.ValueOf(holder).Elem().FieldByName(field).
		SetBool(v.(bool))
}

// Assign sets any null.* wrapper (null.String, null.Int64, null.Bool, null.Time) into a named field.
func Assign(field string, holder any, v any) {
	reflect.ValueOf(holder).Elem().FieldByName(field).
		Set(reflect.ValueOf(v))
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Tiny field builders (optional, to reduce repetitive wiring)
////////////////////////////////////////////////////////////////////////////////////////////////////

// FBool builds a boolean field (plain bool). If your model uses null.Bool, wrap with null.BoolFrom in Assign.
func FBool(label, field, initial string) Field {
	return Field{
		Label:   chalk.Cyan.Color(label),
		Initial: initial,
		Parse:   ParseBool,
		Assign:  func(h any, v any) { AssignBool(field, h, v) },
	}
}

// FInt builds a required int64 field.
func FInt(label, field, initial string, opts ...FieldOpt) Field {
	f := Field{
		Label:   chalk.Cyan.Color(label),
		Initial: initial,
		Parse:   ParseInt64,
		Assign:  func(h any, v any) { AssignInt64(field, h, v) },
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

// FOptInt builds an optional int64 field (null.Int64).
func FOptInt(label, field, initial string, opts ...FieldOpt) Field {
	f := Field{
		Label:   chalk.Cyan.Color(label),
		Initial: initial,
		Parse:   ParseOptInt64,
		Assign:  func(h any, v any) { Assign(field, h, v) },
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

// FDate builds a required date field (time.Time).
func FDate(label, field, initial string, opts ...FieldOpt) Field {
	f := Field{
		Label:    chalk.Cyan.Color(label),
		Initial:  initial,
		Validate: VDate(label),
		Parse:    ParseDate,
		Assign:   func(h any, v any) { AssignDate(field, h, v) },
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

// FOptDate builds an optional date field (null.Time).
func FOptDate(label, field, initial string, opts ...FieldOpt) Field {
	f := Field{
		Label:    chalk.Cyan.Color(label),
		Initial:  initial,
		Validate: VDateOptional(),
		Parse:    ParseOptDate,
		Assign:   func(h any, v any) { AssignNullableDate(field, h, v) },
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

// FString builds a required non-empty string field.
func FString(label, field, initial string, opts ...FieldOpt) Field {
	f := Field{
		Label:    chalk.Cyan.Color(label),
		Initial:  initial,
		Validate: VRequired(label),
		Parse:    ParseNonEmpty(label),
		Assign:   func(h any, v any) { AssignString(field, h, v) },
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

// FOptString remains optional; add opts for consistency.
func FOptString(label, field, initial string, opts ...FieldOpt) Field {
	f := Field{
		Label:   chalk.Cyan.Color(label),
		Initial: initial,
		Parse:   ParseOptString,
		Assign:  func(h any, v any) { Assign(field, h, v) },
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

////////////////////////////////////////////////////////////////////////////////////////////////////

type FieldOpt func(*Field)

func WithInitial(s string) FieldOpt                  { return func(f *Field) { f.Initial = s } }
func WithValidate(v func(string) error) FieldOpt     { return func(f *Field) { f.Validate = v } }
func WithParse(p func(string) (any, error)) FieldOpt { return func(f *Field) { f.Parse = p } }
func WithAssign(a func(any, any)) FieldOpt           { return func(f *Field) { f.Assign = a } }
func WithLabel(lbl string) FieldOpt                  { return func(f *Field) { f.Label = lbl } }

////////////////////////////////////////////////////////////////////////////////////////////////////

func FChoice(label string, initial string, allow []string, setter func(string)) Field {
	lowerSet := make(map[string]struct{}, len(allow))
	for _, v := range allow {
		lowerSet[strings.ToLower(v)] = struct{}{}
	}
	return Field{
		Label:    label,
		Initial:  initial,
		Validate: VRequired(label),
		Parse: func(s string) (any, error) {
			v := strings.ToLower(strings.TrimSpace(s))
			if _, ok := lowerSet[v]; !ok {
				return nil, fmt.Errorf("must be one of: %s", strings.Join(allow, ", "))
			}
			return v, nil
		},
		Assign: func(_ any, v any) { setter(v.(string)) },
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
