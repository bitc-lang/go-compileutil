// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
//
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

package reader

import (
	"io"
	"os"
	"testing"

	"github.com/jsshapiro/go-compileutil/testing_cwd"
)

type Dummy = testing_cwd.Dummy

func doTestReader(r Reader, input []byte, t *testing.T) {
	if r.Offset() != 0 {
		t.Fatalf("Fresh string reader does not start at offset 0")
	}

	if r.Line(0, false) != 1 {
		t.Fatalf("Line number for offset zero is not one")
	}

	for ndx, c := range input {
		off := r.Offset()
		if ndx != int(off) {
			t.Fatalf("Test index %d does not match reader offset %d", ndx, off)
		}

		if b, err := r.Next(); err != nil || b != c {
			t.Fatalf("Byte at offset %d does not match expected %c (error %s)",
				off, c, err)
		}
	}

	b, err := r.Next()
	if err != io.EOF {
		t.Fatalf("Expected EOF at end of reader but got 0x%x (error %s)",
			b, err)

	}

	r.SetOffset(1)
	b, err = r.Peek()
	if err != nil || b != input[1] {
		t.Fatalf("Byte at position 1 did not match after SetOffset(1)")
	}
}

func TestStringReader(t *testing.T) {
	s := "abc"

	r, err := OnString(s)

	expectedPos := "<string>:1:1 (0)"
	if expectedPos != r.Position().String() {
		t.Fatalf("Initial position %d does not give expected position string %s", r.Offset(), expectedPos)
	}

	if err != nil {
		t.Fatalf("Error %s instantiating Reader on string", err.Error())
	}

	doTestReader(r, []byte(s), t)
}

func TestByteReader(t *testing.T) {
	bytes := []byte("abc")

	r, err := OnBytes(bytes)

	if err != nil {
		t.Fatalf("Error %s instantiating Reader on bytes", err.Error())
	}

	doTestReader(r, bytes, t)
}

func TestFileReader(t *testing.T) {
	// Test cases run within their containing directory

	fileName := "reader/reader_test1"

	content, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Error %s reading comparison data from reader_test1", err)
	}

	var r Reader
	r, err = OnFile(fileName)
	if err != nil {
		t.Fatalf("Error %s instantiating Reader on file", err)
	}

	doTestReader(r, content, t)
}

func checkPos(t *testing.T, r Reader, pos int, expect string) {
	r.SetOffset(Offset(pos))
	ps := r.Position().String()
	if ps != expect {
		t.Fatalf("Unexpected position string \"%s\" for offset %d", ps, r.Offset())
	}
}

func TestReaderPosition(t *testing.T) {
	fileName := "reader/reader_test2"

	r, err := OnFile(fileName)
	if err != nil {
		t.Fatalf("Error %v instantiating Reader on file", err)
	}

	r.SetOffset(3)
	c, _ := r.Peek()

	if c != 's' {
		t.Fatalf("Wrong character read")
	}

	checkPos(t, r, 3, "reader/reader_test2:1:4 (3)")
	checkPos(t, r, 20, "reader/reader_test2:2:1 (20)")
}
