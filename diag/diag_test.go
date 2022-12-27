// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

package diag

import (
	"testing"

	"github.com/bitc-lang/go-compileutil/position"
)

const basicString = `x:1:2: Warning: Danger, Will Robinson!
x:2:27: Error: Does not compute!
`

const infoString = `x:1:2: Warning: Danger, Will Robinson!
x:1:5: Info: That's great information!
x:2:27: Error: Does not compute!
`

const sortString = `x:1:2: Warning: Danger, Will Robinson!
x:1:5: Error: Errors before information!
x:1:5: Info: That's great information!
x:2:27: Error: Does not compute!
`

func TestDiagBasics(t *testing.T) {
	diags := New()

	if !diags.Empty() {
		t.Fatalf("Empty diags is not empty")
	}

	if diags.AsError() != nil {
		t.Fatalf("Empty diags should return nil error value")
	}

	diags.AddWarn(position.Pos("x", 1, 2), "Danger, Will Robinson!")
	if diags.AsError() != nil {
		t.Fatalf("Diags without errors should return nil error value")
	}

	diags.AddError(position.Pos("x", 2, 27), "Does not compute!")
	if diags.AsError() == nil {
		t.Fatalf("Diags with errors should return non-nil error value")
	}

	if diags.String() != basicString {
		t.Fatalf("Expected basic output does not validate")
	}
}

func TestDiagMerge1(t *testing.T) {
	diags1 := New()
	diags2 := New()

	diags1.AddWarn(position.Pos("x", 1, 2), "Danger, Will Robinson!")
	diags2.AddError(position.Pos("x", 2, 27), "Does not compute!")

	diags := diags1.With(diags2)
	if diags.AsError() == nil {
		t.Fatalf("Diags with errors should return non-nil error value")
	}

	if diags.String() != basicString {
		t.Fatalf("Expected basic output does not validate")
	}

	diags.AddInfo(position.Pos("x", 1, 5), "That's great information!")
	// fmt.Print(diags)
	if diags.String() != infoString {
		t.Fatalf("Expected basic output does not validate")
	}
}

func TestDiagSort(t *testing.T) {
	diags := New()

	diags.AddWarn(position.Pos("x", 1, 2), "Danger, Will Robinson!")
	diags.AddError(position.Pos("x", 2, 27), "Does not compute!")
	diags.AddInfo(position.Pos("x", 1, 5), "That's great information!")
	// Should sort above info at same position:
	diags.AddError(position.Pos("x", 1, 5), "Errors before information!")

	// fmt.Print(diags)
	if diags.String() != sortString {
		t.Fatalf("Expected basic output does not validate")
	}
}
