// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
//
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

package reader

import (
	"github.com/bitc-lang/go-compileutil/position"
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
// # ⚠ Caution
//
// This implementation is provisional, and probably should not be adopted by
// other parties. The Pos type is exported for efficiency reasons, and a better
// implementation is coming. The reason I did not adopt a design more similar to
// the Go tokenizer position manager is that I wanted to be able to support
// input and position tracking for multiple interactive streams. I have since
// realized that expanding the Go-style single integer Pos values to int64 would
// allow this, with the caveat that we would need to set an upper bound (say
// 2GB) on total input length for any given unit of compilation.
//
// That implementation would be both more compact nad more efficient than this
// one, and I will probably migrate. The resulting Pos value will be source
// compatible but not binary compatible.
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
