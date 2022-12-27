# Utility code for compilers, interpreters, and related programs.

This package implements several utility packages for compilers, interpreters,
and similar programs. Such programs tend to have common elements, and I grew
tired of writing and rewriting the same code over and over again - each failing
to capture bug fixes from its predecessor. Some of these packages are intended
to be useful for _any_ compiler. Others are useful in selected common cases.

The main packages here are:

- `diag` - A package for conventional diagnostics supporting multiple diagnostic
  levels (info, warn, error, fatal). Diagnostic sets comply with Go's `error`
  interface. Diagnostic sets are mergeable, allowing them to be used in parsers
  that rely on roll back and replay.
- `indentedwriter` - An implementation of `io.Writer` that facilitates proper
  output indentation.
- `intern` - A package that maps byte sequences into unique instances, mildly
  specialized for common cases of symbol names found in compilation units.
- `position` - An abstract type (an interface) for recording input positions,
  with a usable example implementation.
- `reader` - Support for low-level source file I/O, including condensed position
  tracking and input backtracking.
- `testing_cwd` - A package you can import to ensure that _go test_, _dlv test_,
  and others all execute their tests from the same working directory. This is
  helpful if you want to write tests that source input files from the tree and
  have them work under the various different testing tools.

The principal documentation for these packages is the go documentation, which
can be browsed online at
[pkg.go.com](https://pkg.go.dev/github.com/bitc-lang/go-compileutil).