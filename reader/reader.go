// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
//
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

// Package reader provides byte-level I/O and position tracking for compilers
// and interpreters, including backtracking support.
package reader

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
	"sync"

	"github.com/bitc-lang/go-compileutil/position"
)

// A Reader represents an input unit of compilation.
type Reader interface {
	// Close this input, discarding any consumed bytes but preserving any line
	// and column information that has been constructed.
	//
	// This function does not close the underlying file descriptor if the input
	// source is a stream.
	Close() error

	// Return the reader's current input position
	Position() position.Position

	// Return the reader's current input offset
	Offset() Offset

	// Re-set the reader's current offset
	SetOffset(o Offset) error

	// Return true iff we are at end of input
	IsAtEOI() bool

	// Return the byte at the specified offset within this input unit.
	//
	// If the input unit is a stream, and the requested position exceeds the
	// number of bytes read so far, this operation will block for input.
	ByteAt(o Offset) (byte, error)

	// Get the byte at the current offset without advancing the position.
	//
	// If the input unit is a stream, and the current position exceeds the
	// number of bytes read so far, this operation will block for input.
	Peek() (byte, error)

	// Return the byte at the current offset and advance the offset.
	//
	// If the input unit is a stream, and the current position exceeds the
	// number of bytes read so far, this operation will block for input.
	Next() (byte, error)

	// Return a user readable string representation of offset o.
	//
	// Adjusts for line directives iff adjusted is true.
	//
	// Defined for any position p < r+1, where r is the greatest position that
	// has been successfully accessed by ByteAt().
	PositionString(o Offset, adjusted bool) string

	// Return the file name associated with this position.
	//
	// Adjusts for line directives iff adjusted is true.
	//
	// Defined for any position p < r+1, where r is the greatest position that
	// has been successfully accessed by ByteAt().
	Filename(o Offset, adjusted bool) string

	// Return the line number associated with offset o (starts at 1).
	//
	// Adjusts for line directives iff adjusted is true.
	//
	// Defined for any position p < r+1, where r is the greatest position that
	// has been successfully accessed by ByteAt().
	Line(o Offset, adjusted bool) int

	// Return the column number associated with offset o (starts at 1).
	//
	// Adjusts for line directives iff adjusted is true.
	//
	// Defined for any position p < r+1, where r is the greatest position that
	// has been successfully accessed by ByteAt().
	Column(o Offset, adjusted bool) int

	// Return the name, line and column number associated with offset o.
	//
	// Adjusts for line directives iff adjusted is true.
	//
	// Line and column numbers start at 1. Especially when adjusting, this is
	// significanty more efficient than extracting the elements individually.
	//
	// Defined for any position p < r+1, where r is the greatest position that
	// has been successfully accessed by ByteAt().
	NameLineAndColumn(o Offset, adjusted bool) (string, int, int)
}

var mu sync.Mutex

type reader struct {
	name         string   // Name of this input unit.
	content      []byte   // Bytes loaded so far.
	lines        []Offset // Starting offset for each line seen to date.
	offset       Offset   // Current offset in the input streaam or file.
	updatedTo    Offset   // Line starts have been computed to here.
	source       fs.File  // Input file
	ioChunkSize  int      // How much to read
	isCharDevice bool     // True iff input is a character device
	closeSource  bool     // Whether to close the source on reader close
	err          error    // Last I/O error
}

const blockChunkSize = 1024
const ttyChunkSize = 1

func (r *reader) Close() error {
	if r.closeSource {
		r.source.Close()
	}
	r.content = nil
	return nil
}

func (r *reader) Position() position.Position {
	return &Pos{input: r, off: r.offset}
}

// Return the reader's current input offset
func (r *reader) Offset() Offset {
	return r.offset
}

// Read bytes until the content buffer contains the offset o
func (r *reader) expandTo(o Offset) error {
	if int(o) < len(r.content) {
		return nil
	}

	if r.ioChunkSize == 0 || r.source == nil {
		return io.EOF
	}

	// io.Read() is allowed to return a short result, so this needs to be a loop:
	for int(o) >= len(r.content) && r.err == nil {
		nBytes := (int(o) - len(r.content)) + 1
		if r.ioChunkSize > 1 {
			nBytes += (r.ioChunkSize - 1)
			nBytes &= -r.ioChunkSize
		}

		newBytes := make([]byte, nBytes)
		nBytes, r.err = r.source.Read(newBytes)

		if r.err != nil && r.err != io.EOF {
			panic(fmt.Sprintf("Content expansion returns %d bytes (err %v) reading %d from %s for offset %d in content %d",
				nBytes, r.err, cap(newBytes), r.fileName(0, false), o, len(r.content)))
			// return io.EOF
		}

		r.content = append(r.content, newBytes[:nBytes]...)
	}

	r.updateLines()

	// io.Read() can return an error even if it successfully returns the desired
	// bytes. But if we have the bytes we need
	if int(o) < len(r.content) {
		return nil
	}

	return r.err
}

// Return the byte at offset o from this reader.
//
// Note that this may block.
func (r *reader) ByteAt(o Offset) (byte, error) {
	if err := r.expandTo(o); err != nil {
		return 0, err
	}

	if int(o) >= len(r.content) {
		return 0, io.EOF
	}

	// if expandTo returned no error, we have enough room in r.content to fetch
	// the byte.
	return r.content[o], nil

}

func (r reader) IsAtEOI() bool {
	return r.err != nil
}

// func (r *reader) ByteAt(o Offset) (byte, error) {
// 	if int(0) < len(r.content) {
// 		return r.content[int(o)], nil
// 	}

// 	// return 0,
// 	//diag.New().AddError(r.Position(o), "Read past end of file\n")
// 	return 0, nil
// }

func (r *reader) fileName(o Offset, adjusted bool) string {
	if adjusted {
		if err := r.expandTo(o - 1); err != nil {
			panic(fmt.Sprintf("Offset %d out of range for reader %s", o, r.name))
		}
	}

	return r.name
}

// Set the reader's current input offset to o, reading any bytes necessary for
// that offset to be valid.
func (r *reader) SetOffset(o Offset) error {
	if int(o) >= len(r.content) {
		_, err := r.ByteAt(o)
		if err != nil {
			return err
		}
	}

	r.offset = o
	return nil
}

func (r *reader) Peek() (byte, error) {
	return r.ByteAt(r.offset)
}
func (r *reader) Next() (byte, error) {
	b, err := r.ByteAt(r.offset)
	if err == nil {
		r.offset++
	}
	return b, err
}

func (r *reader) PositionString(o Offset, adjusted bool) string {
	nm, line, col := r.NameLineAndColumn(o, adjusted)

	// We always have a file name because of the way readers are created, but the
	// line and column number may not be valid

	s := nm
	if line > 0 {
		s = fmt.Sprintf("%s:%d:%d", s, line, col)
	}

	if int(o) < 0 {
		// No offset information
		return s
	}

	if int(o) <= len(r.content) {
		s = fmt.Sprintf("%s (%d)", s, o)
	} else {
		s = fmt.Sprintf("%s (%d > %d)", s, o, len(r.content))
	}

	return s
}

func (r *reader) line(o Offset, adjusted bool) int {
	if err := r.expandTo(o - 1); err != nil {
		panic(fmt.Sprintf("Offset %d out of range for reader %s", o, r.name))
	}

	// Note that the predicate function here looks for the first line whose
	// starting offset is GREATER than the target offset o, which means that the
	// value returned is one more than the desired array position, which means
	// that it is a 1-relative result.
	//
	// This means that sort.Search is going to return the "not found" value if o
	// falls within the last line, but we know that the answer must be valid
	// because expandTo has succeeded.
	return sort.Search(len(r.lines), func(i int) bool { return r.lines[i] > o })
}

func (r *reader) NameLineAndColumn(o Offset, adjusted bool) (string, int, int) {
	s := r.fileName(o, adjusted)

	if int(o) > len(r.content) {
		return s, 0, 0
	}

	l := r.line(o, adjusted) - 1
	off := o - r.lines[l]
	return s, 1 + l, 1 + int(off)
}

func (r *reader) Filename(o Offset, adjusted bool) string {
	s, _, _ := r.NameLineAndColumn(o, adjusted)
	return s
}
func (r *reader) Line(o Offset, adjusted bool) int {
	_, l, _ := r.NameLineAndColumn(o, adjusted)
	return l
}
func (r *reader) Column(o Offset, adjusted bool) int {
	_, _, c := r.NameLineAndColumn(o, adjusted)
	return c
}

func (r *reader) updateLines() {
	for i := int(r.updatedTo); i < len(r.content); i++ {
		if r.content[i] == '\n' {
			r.lines = append(r.lines, Offset(i+1))
		}
	}
	r.updatedTo = Offset(len(r.content))
}

func (r *reader) AddContent(bytes []byte) {
	mu.Lock()
	defer mu.Unlock()

	r.content = append(r.content, bytes...)
	r.updateLines()
}

func setReaderAttrs(name string, r *reader) error {
	var err error = nil // until proven otherwise

	const devModes = os.ModeDevice | os.ModeCharDevice
	var fi fs.FileInfo

	fi, err = os.Stat(name)
	if err != nil {
		return err
	}

	r.isCharDevice = (fi.Mode() & devModes) == devModes
	if r.isCharDevice {
		r.ioChunkSize = ttyChunkSize
	}

	isBlockDevice := (fi.Mode() & devModes) == fs.ModeDevice
	isIrregular := (fi.Mode() & fs.ModeIrregular) != 0

	if isBlockDevice || isIrregular {
		return errors.New("refusing to parse block device or irregular input file")
	}

	if fi.Mode().IsRegular() {
		if err := r.expandTo(Offset(int(fi.Size()) - 1)); err != nil {
			return err
		}
	}

	return nil
}

func OnFile(name string) (Reader, error) {
	source, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	rdr := &reader{
		name:         name,
		content:      []byte{},
		lines:        []Offset{0}, // first line starts at position 0
		offset:       Offset(0),
		updatedTo:    Offset(0),
		source:       source,
		ioChunkSize:  blockChunkSize,
		isCharDevice: false,
		err:          nil,
		closeSource:  true,
	}

	return rdr, nil
}

func onNamedBytes(name string, content []byte) (Reader, error) {
	rdr := &reader{
		name:         name,
		content:      append([]byte{}, content...),
		lines:        []Offset{0}, // first line starts at position 0
		offset:       Offset(0),
		updatedTo:    Offset(0),
		source:       nil,
		ioChunkSize:  0,
		isCharDevice: false,
		err:          nil,
		closeSource:  false,
	}
	rdr.updateLines()
	return rdr, nil
}

func OnBytes(content []byte) (Reader, error) {
	return onNamedBytes("<[]byte>", content)
}

func OnString(s string) (Reader, error) {
	return onNamedBytes("<string>", []byte(s))
}
