// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

package position

import (
	"testing"
)

func TestCWD(t *testing.T) {
	if p := Pos("file", 1, 2); p.String() != "file:1:2" {
		t.Fatalf("Bad string form of valid pos: %s", p)
	}
	if p := OffsetPos("file", 1, 2, 3); p.String() != "file:1:2 (3)" {
		t.Fatalf("Bad string form of valid offset pos: %s", p)
	}
	if p := Pos("file", 1, 2); p.Raw().String() != "file:1:2" {
		t.Fatalf("Bad string form of raw valid pos: %s", p)
	}
	if p := Pos("file", 0, 2); p.String() != "file" {
		t.Fatalf("Bad string form of invalid pos: %s", p)
	}
}
