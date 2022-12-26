// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
//
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

package position

import "fmt"

// A basic position value.
//
// This type exists primarily for testing. It can be used as a starting
// point, but where performance or space are important you will want a more
// condensed representation. The one in the reader package is an example.
type BasicPos struct {
	filename string
	line     int
	column   int
	offset   int
}

// -------------------------------------------------------------------------
// Implement Stringer interface
func (pos *BasicPos) String() string {
	// We always have a file name because of the way readers are created, but the
	// line and column number may not be valid

	s := pos.filename
	if pos.line > 0 {
		s = fmt.Sprintf("%s:%d:%d", s, pos.line, pos.column)
	}

	if pos.offset >= 0 {
		s = fmt.Sprintf("%s (%d)", s, pos.offset)
	}

	return s
}

// -------------------------------------------------------------------------
// Implement Position interface
func (pos *BasicPos) Filename() string {
	return pos.filename
}
func (pos *BasicPos) Line() int {
	return pos.line
}
func (pos *BasicPos) Column() int {
	return pos.column
}
func (pos *BasicPos) Offset() int {
	return pos.offset
}
func (pos *BasicPos) Raw() Position {
	return pos
}

func Pos(nm string, line, col int) *BasicPos {
	return &BasicPos{filename: nm, line: line, column: col, offset: -1}
}
func OffsetPos(nm string, line, col int, off int) *BasicPos {
	return &BasicPos{filename: nm, line: line, column: col, offset: off}
}
