// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

// Package intern maps byte slices into unique symbol identifiers and private
// copies of those byte slices having the same value.
//
// Byte slices returned by Intern() should not be modified by the caller!
package intern

import (
	"sync"
)

// This is decidedly a quick and dirty implementation!

type Symbol int

// Symbols shorter than 1024 bytes will be appended to successive 4k byte buffers
const bufSize = 4096
const maxToBuf = 1024

var mu sync.Mutex
var next = 0
var symbols = make(map[string]int)
var index = make(map[int]([]byte))
var byteBuf = make([]byte, bufSize)

// Given a byte slice, return a duplicate copy by appending it to the current
// byteBuf if possible or simply making a duplicate if it is larget than
// maxToBuf.
func dup(b []byte) []byte {
	var key []byte

	l := len(b)
	if l > maxToBuf {
		key = make([]byte, l)
		copy(key, b)
	}

	bufLen := len(byteBuf)
	if l > cap(byteBuf)-bufLen {
		byteBuf = make([]byte, bufSize)
		bufLen = 0
	}

	byteBuf = append(byteBuf, b...)
	key = byteBuf[bufLen : bufLen+l]

	return key
}

// Given a byte slice, return a (symbol, []byte) pair providing the unique
// symbol number assigned and a copy of the byte slice payload that has been
// privately recorded by the intern subsystem.
func Intern(b []byte) (Symbol, []byte) {
	mu.Lock()
	defer mu.Unlock()

	val, ok := symbols[string(b)]
	if ok {
		key := index[val]
		return Symbol(val), key
	}

	// Make a private copy so that a big input string can be GC'd when we are
	// done with it.
	key := dup(b)

	symbols[string(key)] = next
	index[next] = key

	val = next
	next++
	return Symbol(val), key
}

// Return true iff symbols s and s2 refer to the same byte slice.
func (s Symbol) Equals(s2 Symbol) bool {
	return int(s) == int(s2)
}

// Return the string representation of symbol s.
//
// Note that this can have unexpected results if the byte string associated with
// s is not a valid Unicode string.
func (s Symbol) String() string {
	mu.Lock()
	defer mu.Unlock()

	return string(index[int(s)])
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Retrun true iff the byte string denoted by s is bytewise less than the
// bytestring denoted by s2 according to bytewise lexicographic comparison.
func (s Symbol) LessThan(s2 Symbol) bool {
	// Conventional maps are not safe for concurrent reads

	mu.Lock()
	defer mu.Unlock()

	v1 := index[int(s)]
	v2 := index[int(s2)]

	ln := min(len(v1), len(v2))

	for ndx := 0; ndx < ln; ndx++ {
		if v1[ndx] < v2[ndx] {
			return true
		} else if v1[ndx] > v2[ndx] {
			return false
		}
	}

	// Lexicographic. Less if v1 is shorter:
	return len(v1) < len(v2)
}
