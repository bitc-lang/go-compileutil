// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

package indentedwriter

import (
	"bytes"
	"fmt"
	"testing"
)

const check1 = `Unindented line
line withmultiple chunks
  Indented by 2
    Indented by 4
Unindented again
`

func TestMultiIndent(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	iw := IndentedWriterOn(buf)

	iw.Print("Unindented line\n")
	iw.Print("line with")
	iw.Print("multiple chunks\n")

	iw2 := iw.Indent(2)
	iw2.Print("Indented by 2\n")
	iw2.Indent(2).Println("Indented by 4")

	iw.Println("Unindented again")

	fmt.Print(check1)
	if buf.String() != check1 {
		t.Fatalf("Indented string mismatch on check1")
	}
}
