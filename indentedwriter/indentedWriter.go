// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

// Package indentedWriter implements an io.Writer that prefaces each output line
// with a defined number of spaces.
package indentedwriter

import (
	"bytes"
	"fmt"
	"io"
)

const traceBOL = false

type Printable interface {
	Print(a ...any) (n int, err error)
	Println(a ...any) (n int, err error)
	Printf(format string, a ...any) (n int, err error)
}

type baseWriter struct {
	out   io.Writer
	AtBOL bool
}

// An io.Writer that prefixes each line with a defined number of spaces.
//
// Note that IndentedWriter does not implement a Close() method, because there
// is no way to know how many other IndentedWriters may be simultaneously using
// the same output interface.
//
// IndentedWriter is not safe for concurrency.
type IndentedWriter struct {
	*baseWriter
	indent int
}

const lineFeed = byte('\n')

// Return a fresh IndentedWriter on the given Writer interface.
func IndentedWriterOn(out io.Writer) *IndentedWriter {
	return &IndentedWriter{
		baseWriter: &baseWriter{
			out:   out,
			AtBOL: false,
		},
		indent: 0,
	}
}

// Write bytes to the IndentedWriter receiver.
//
// This satisfies the io.Writer interface, which lets us pass IdentedWriter to
// many of the functions in the various Go io libraries.
func (iw *IndentedWriter) Write(b []byte) (n int, err error) {
	threads := bytes.SplitAfter(b, []byte("\n"))
	if len(threads[len(threads)-1]) == 0 {
		threads = threads[:len(threads)-1]
	}
	pfx := bytes.Repeat([]byte(" "), iw.indent)
	written := 0

	for _, bt := range threads {
		if iw.AtBOL {
			iw.out.Write(pfx)
		}
		n, err := iw.out.Write(bt)
		written += n
		if err != nil {
			break
		}

		iw.AtBOL = len(bt) > 0 && bt[len(bt)-1] == lineFeed
	}

	return written, err
}

// Given an IndentedWriter, return a fresh IndentedWriter on the same Weiter
// interface whose start-of-line indentation is increased by n.
func (iw *IndentedWriter) Indent(n int) *IndentedWriter {
	return &IndentedWriter{
		baseWriter: iw.baseWriter,
		indent:     iw.indent + n,
	}
}

// Given an IndentedWriter, return a fresh IndentedWriter on the same Weiter
// interface whose start-of-line indentation is zero.
//
// When emitting indented output, this is useful for writing pragmas such as
// line number directives.
func (iw *IndentedWriter) NoIndent() *IndentedWriter {
	return &IndentedWriter{
		baseWriter: iw.baseWriter,
		indent:     0,
	}
}

// Method for Print, allowing IndentedWriter to subsume fmt
func (iw *IndentedWriter) Print(a ...any) (n int, err error) {
	if traceBOL && iw.AtBOL {
		fmt.Printf("AT BOL %d\n", iw.indent)
	}
	return fmt.Fprint(iw, a...)
}

// Method for Println, allowing IndentedWriter to subsume fmt
func (iw *IndentedWriter) Println(a ...any) (n int, err error) {
	if traceBOL && iw.AtBOL {
		fmt.Printf("AT BOL %d\n", iw.indent)
	}
	return fmt.Fprintln(iw, a...)
}

// Method for Printf, allowing IndentedWriter to subsume fmt
func (iw *IndentedWriter) Printf(format string, a ...any) (n int, err error) {
	if traceBOL && iw.AtBOL {
		fmt.Printf("AT BOL %d\n", iw.indent)
	}
	return fmt.Fprintf(iw, format, a...)
}
