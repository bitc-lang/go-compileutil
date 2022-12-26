// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

package testing_cwd

import (
	"os"
	"path"
	"testing"
)

func TestCWD(t *testing.T) {
	dir, err := os.Getwd()

	if err != nil {
		t.Fatalf("Unable to chdir to get current directory (error %v)", err)
	}

	// The repository can be checked out in any parent directory, but we can at
	// least confirm that the tail element is what we expect:
	if path.Base(dir) != "go-compileutil" {
		t.Fatalf("init routine did not chdir to expected directory")
	}

}
