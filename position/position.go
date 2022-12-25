// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
//
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

// Abstraction of a position for use by consumers.
//
// This interface intentionally makes no provision for updating the position.
// If it did, compiler implementations would incur two forms of overhead:
//
//   - Every query operation on a Position would involve an indirection
//   - Since positions are intended to be read-only, every advance of a Position
//     would incur an allocation overhead.
//
// The second one is the main concern, because positions are manipulated very
// frequently in parsers and formatters. The practical consequence is that such
// tools must provide their own position implementation, such as the one
// in reader.Pos.
package position

import (
	"fmt"
)

// The position abstraction
type Position interface {
	fmt.Stringer

	Filename() string // Return the file name associated with this position.
	Line() int        // Return the line number (starting at 1) of this position.
	Column() int      // Return the column number (starting at 1) of this position.
	Offset() int      // Return the byte offset (starting at 0) of this position.
	Raw() Position    // Return aform of the position that does not take line number directives into account.
}
