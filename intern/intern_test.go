// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

package intern

import (
	"testing"
)

func TestBasics(t *testing.T) {
	for i := 0; i < 256; i++ {
		if Intern([]byte{byte(i)}) != Symbol(i) {
			t.Fatalf("Single-byte string %d not assigned reserved symbol", i)
		}
	}

	s := "abc"
	b := []byte{'a', 'b', 'c'}
	if s != string(b) {
		t.Fatalf("Matching string and []byte value are not equal")
	}

	s1 := InternString(s)
	s2 := Intern(b)

	if s1 != s2 {
		t.Fatalf("String value %d and []byte value %d do not generate same symbol", s1, s2)
	}
}
