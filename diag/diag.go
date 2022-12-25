// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

// Package diag implements a traditional multi-level diagnostic system
// compatible with the Go error interface.
//
// The diag package implements four levels of message;
//
//   - Fatal: diagnostics that immediately terminate the program, for use
//     when a program cannot make progress at all or a problem is so severe
//     that no output should be produced.
//   - Errors: diagnostics that lead to program exit without output, but are
//     sufficiently recoverable that further useful diagnostic output remains
//     possible before exiting.
//   - Warnings: diagnostics that indicate something is not advisable, but also
//     not an error.
//   - Info: diagnostics that provide informational messages to the user, such
//     as copyright notices, versions information, and the like.
//
// When printed, or returned as an error value, diagnostics are organized in
// sorted by input position. If a sorting function is not explicitly provided,
// position strings are assumed to take the form
//
//	filename:line:column
//
// Diagnostic messages (type Diag) are organized into groups (type Diags).
// Groups can be merged using the Diags.With() message. This is useful primarily
// in compiler-like applications that implement rollback with roll-forward.
package diag

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jsshapiro/go-compileutil/position"
)

type Position position.Position

type DiagKind int

const (
	// Errors that are immediately fatal
	Fatal DiagKind = iota
	// Errors that terminate execution, typically at the end of the current
	// pass.
	Error
	// Warnings that do not terminate execution.
	Warning

	// Informational diagnostics that do not terminate execution.
	Info
)

//go:generate stringer -type=DiagKind

type diag struct {
	Pos     string
	Kind    DiagKind
	Message string
}

type diags struct {
	// Function to be used when sorting positions.
	//
	// Different applications form position strings in different ways, so the
	// sorting algorithm is app specific (and sometimes library specific).
	// The defaul algorithm is worth a try before customizing.
	Sort     func([]Diag) []Diag
	HasError bool
	diags    []Diag
}

type Diag = *diag   // Export as a heap-allocated type
type Diags = *diags // Export as a heap-allocated type

// Return a string represnting a specific diagnostic message.
func (d Diag) String() string {
	return fmt.Sprintf("%s: %s %s", d.Pos, d.Kind, d.Message)
}

// Return a string containing all diagnostics in the diagnostic group, sorted
// by the active sorting algorithm.
func (d Diags) String() string {
	s := []string{}
	for _, d := range d.Sort(d.diags) {
		s = append(s, d.String())
	}

	s = append(s, "") // Ensures trailing newline

	return strings.Join(s, "\n")
}

func (d Diags) Empty() bool {
	return len(d.diags) == 0
}

// Implement the error.Error() interface, so that a diagnostic set can be
// returned as an error value.
func (d Diags) Error() string {
	return d.String()
}

// Return a go error value or nil according to whether errors are present.
//
// Note that warning and informational diagnostics are not considered errors for
// this purpose. This is contrary to the prevailing school of thought favored by
// the Go team.
func (d Diags) AsError() error {
	if d.HasError {
		return d
	}
	return nil
}

func defaultSort(d []Diag) []Diag {
	sort.Slice(d, func(i1, i2 int) bool {
		d1 := d[i1]
		d2 := d[i2]

		if d1.Pos > d2.Pos {
			return false
		}
		if d1.Pos < d2.Pos {
			return true
		}
		// Positions are equal
		if d1.Kind > d2.Kind {
			return false
		}
		if d1.Kind < d2.Kind {
			return true
		}
		// Severities are equal
		return d1.Message < d2.Message
	})

	return d
}

// Return a new diagnostic group.
func New() Diags {
	c := &diags{
		HasError: false,
		Sort:     defaultSort,
		diags:    []Diag{},
	}
	return c
}

// Add a diagnostic with the specified location, severity, and message payload
func (c Diags) Add(where Position, kind DiagKind, msg string) Diags {
	diag := &diag{Pos: where.String(), Kind: kind, Message: msg}
	c.diags = append(c.diags, diag)
	switch kind {
	case Error:
		c.HasError = true
	case Fatal:
		fmt.Fprintln(os.Stderr, diag.String())
		os.Exit(-1)
	}

	return c
}

// Issue a fatal diagnostic giving the specified location and message.
func (c Diags) AddFatal(where Position, msg string) Diags {
	return c.Add(where, Fatal, msg)
}

// Record an error diagnostic at the specified location with the provided
// message.
func (c Diags) AddError(where Position, msg string) Diags {
	return c.Add(where, Error, msg)
}

// Record a warning diagnostic at the specified location with the provided
// message.
func (c Diags) AddWarn(where Position, msg string) Diags {
	return c.Add(where, Warning, msg)
}

// Record an informational diagnostic at the specified location with the
// provided message.
func (c Diags) AddInfo(where Position, msg string) Diags {
	return c.Add(where, Info, msg)
}

// Return a fresh diagnostic group combining the diagnostics of two existing
// groups.
//
// The sorting criteria of the receiver is used by the fresh diagnostic group.
func (c Diags) With(d Diags) Diags {
	fresh := []Diag{}
	fresh = append(fresh, c.diags...)
	fresh = append(fresh, d.diags...)

	return &diags{
		HasError: c.HasError || d.HasError,
		Sort:     c.Sort,
		diags:    fresh,
	}
}
