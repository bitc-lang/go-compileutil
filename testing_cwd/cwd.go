// Hack to ensure that the current working directory for test execution is the
// project top level directory.
//
// Recommended usage:
//
//	import "github.com/jsshapiro/go-compileutil/testing_cwd"
//
//	type Dummy = testing_cwd.Dummy
//
// At least for now (go 1.19.4), complaining about unused type bindings hasn't
// occurred to the go authors.
package testing_cwd

import (
	"fmt"
	"os"
	"path"
	"runtime"
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Dir(filename)
	dir = path.Join(dir, "..")

	err := os.Chdir(dir)

	if err != nil {
		panic(fmt.Sprintf("Unable to chdir to %s (error %v)", dir, err))
		// } else {
		// 	wd, _ := os.Getwd()
		// 	fmt.Printf("Now running with CWD %s\n", wd)
	}
}

type Dummy int
