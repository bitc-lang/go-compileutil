// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
//
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

package reader

import (
	"github.com/jsshapiro/go-compileutil/position"
)

// An offset within an input stream.
type Offset int

// Condensed Pos value.
//
// Pos is a read-only value type. Queries on the baseline form respects line
// number directives, because this is the form most commonly useful for
// diagnostics.
//
// This is not as condensed as the Go tokenizer version, but the Go version
// cannot support multiple interactive streams in its current form.
//
// NOTE(shap): It woould be possible to migrate Pos values to an int64
// representation in source-compatible way via an implementation closer to the
// Go tokenizer. This would reduce storage requirements in 64 bit compilers, but
// it seems unlikely that position management will be the bottleneck in a
// compiler that is required to process more than 2^31 bytes of input.
type Pos struct {
	input Reader
	off   Offset
}

// Advance the position by n bytes
//
// FIX(shap): blah blah
func (p Pos) Advance(n int) Pos {
	return Pos{input: p.input, off: p.off + Offset(n)}
}

func (p Pos) Next() Pos {
	return p.Advance(1)
}

func (p Pos) Clone() Pos {
	return Pos{input: p.input, off: p.off}
}

// Return a human-readable representation of this position.
func (p Pos) String() string {
	return p.input.PositionString(p.off, true)
}

// ------------------------------------------------------------------------
// Implement the Position interface

// Return the file name associated with this position.
func (p Pos) Filename() string {
	nm, _, _ := p.input.NameLineAndColumn(p.off, true)
	return nm
}

// Return the line number (starting at 1) of this position.
//
// If adjusted is true, the value returned takes line directives (pragmas) into account.
func (p Pos) Line() int {
	_, l, _ := p.input.NameLineAndColumn(p.off, true)
	return l
}

// Return the column number (starting at 1) of this position.
//
// If adjusted is true, the value returned takes line directives (pragmas) into account.
func (p Pos) Column() int {
	_, _, c := p.input.NameLineAndColumn(p.off, true)
	return c
}

// Return the byte offset (starting at 0) of this position.
func (p Pos) Offset() int {
	return int(p.off)
}

func (p Pos) Raw() position.Position {
	return RawPos{Pos: p}
}

// Pos variant that ignores line directives (pragmas).
type RawPos struct {
	Pos
}

// Return the file name associated with this position,
// ignoring any line directives (pragmas).
func (p RawPos) Filename() string {
	return p.input.Filename(p.off, false)
}

// Return the line number (starting at 1) associated with this position,
// ignoring any line directives (pragmas).
func (p RawPos) Line() int {
	return p.input.Column(p.off, false)
}

// Return the column number (starting at 1) associated with this position,
// ignoring any line directives (pragmas).
func (p RawPos) Column() int {
	return p.input.Column(p.off, false)
}

func (p RawPos) Adjusted() position.Position {
	return p
}
