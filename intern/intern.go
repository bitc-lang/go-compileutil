// Copyright (c) 2022 Jonathan S. Shapiro. All rights reserved.
// Use of this source code is governed by terms that can be
// found in the LICENSE file.

// Package intern maps byte slices into unique symbol identifiers and private
// copies of those byte slices having the same value.
//
// Byte slices returned by Intern() should not be modified by the caller!
package intern

import (
	"fmt"
	"sync"
)

// This is decidedly a quick and dirty implementation!

type Symbol int

const firstPrintable = byte('!')
const lastPrintable = byte('~')

// Single character variable names show up a lot...
var unicodeISO = []byte{
	'\x00', '\x01', '\x02', '\x03', '\x04', '\x05', '\x06', '\x07',
	'\x08', '\x09', '\x0A', '\x0B', '\x0C', '\x0D', '\x0E', '\x0F',
	'\x10', '\x11', '\x12', '\x13', '\x14', '\x15', '\x16', '\x17',
	'\x18', '\x19', '\x1A', '\x1B', '\x1C', '\x1D', '\x1E', '\x1F',
	' ', '!', '"', '#', '$', '%', '&', '\'',
	'(', ')', '*', '+', ',', '-', '.', '/',
	'0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9', ':', ';', '<', '=', '>', '?',
	'@', 'A', 'B', 'C', 'D', 'E', 'F', 'G',
	'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O',
	'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W',
	'X', 'Y', 'Z', '[', '\\', ']', '^', '_',
	'`', 'a', 'b', 'c', 'd', 'e', 'f', 'g',
	'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
	'p', 'q', 'r', 's', 't', 'u', 'v', 'w',
	'x', 'y', 'z', '{', '|', '}', '~', '\x7F',
	'\x80', '\x81', '\x82', '\x83', '\x84', '\x85', '\x86', '\x87',
	'\x88', '\x89', '\x8A', '\x8B', '\x8C', '\x8D', '\x8E', '\x8F',
	'\x90', '\x91', '\x92', '\x93', '\x94', '\x95', '\x96', '\x97',
	'\x98', '\x99', '\x9A', '\x9B', '\x9C', '\x9D', '\x9E', '\x9F',
	'\xA0', '¡', '¢', '£', '¤', '¥', '¦', '§',
	'¨', '©', 'ª', '«', '¬', '\xAD', '®', '¯',
	'°', '±', '²', '³', '´', 'µ', '¶', '·',
	'¸', '¹', 'º', '»', '¼', '½', '¾', '¿',
	'À', 'Á', 'Â', 'Ã', 'Ä', 'Å', 'Æ', 'Ç',
	'È', 'É', 'Ê', 'Ë', 'Ì', 'Í', 'Î', 'Ï',
	'Ð', 'Ñ', 'Ò', 'Ó', 'Ô', 'Õ', 'Ö', '×',
	'Ø', 'Ù', 'Ú', 'Û', 'Ü', 'Ý', 'Þ', 'ß',
	'à', 'á', 'â', 'ã', 'ä', 'å', 'æ', 'ç',
	'è', 'é', 'ê', 'ë', 'ì', 'í', 'î', 'ï',
	'ð', 'ñ', 'ò', 'ó', 'ô', 'õ', 'ö', '÷',
	'ø', 'ù', 'ú', 'û', 'ü', 'ý', 'þ', 'ÿ',
}

// Symbols shorter than 1024 bytes will be appended to successive 4k byte buffers
const bufSize = 4096
const maxToBuf = 1024

var mu sync.Mutex
var next = 256 // So as not to collide with single-character variable names.
var symbols = make(map[string]int)
var index = make(map[int]([]byte))
var byteBuf = make([]byte, 0, bufSize)

// Given a byte slice, return a duplicate copy by appending it to the current
// byteBuf if possible or simply making a duplicate if it is larget than
// maxToBuf.
func dup(b []byte) []byte {
	var key []byte

	l := len(b)
	if l > maxToBuf {
		key = make([]byte, l)
		copy(key, b)
	} else {
		bufLen := len(byteBuf)
		if l > cap(byteBuf)-bufLen {
			byteBuf = make([]byte, 0, bufSize)
			bufLen = 0
		}

		byteBuf = append(byteBuf, b...)
		key = byteBuf[bufLen : bufLen+l]
	}

	return key
}

// Given a byte slice, return a (symbol, []byte) pair providing the unique
// symbol number assigned and a copy of the byte slice payload that has been
// privately recorded by the intern subsystem.
func Intern(b []byte) Symbol {
	// Programming-oriented optimization: There are lots of uses of single
	// character ASCII variable names following the examples of stone age FORTRAN
	// programs carved laboriously onto stone punch cards by prehistoric
	// programmers.
	if len(b) == 1 && b[0] <= 255 {
		return Symbol(b[0])
	}

	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("Looking up |%s|\n", string(b))

	val, ok := symbols[string(b)]
	if ok {
		return Symbol(val)
	}

	// Make a private copy so that a big input string can be GC'd when we are
	// done with it.
	key := dup(b)

	symbols[string(key)] = next

	index[next] = key

	val = next
	next++
	return Symbol(val)
}

func InternString(s string) Symbol {
	return Intern([]byte(s))
}

// Return true iff symbols s and s2 refer to the same byte slice.
func (s Symbol) Equals(s2 Symbol) bool {
	return int(s) == int(s2)
}

// Return the byte slice representation of symbol s.
//
// Note that this can have unexpected results if the byte string associated with
// s is not a valid Unicode string.
func (s Symbol) Bytes() []byte {
	i := int(s)
	if i < 256 {
		return unicodeISO[i : i+1]
	}

	mu.Lock()
	defer mu.Unlock()

	return index[int(s)]
}

// Return the string representation of symbol s.
//
// Note that this can have unexpected results if the byte string associated with
// s is not a valid Unicode string.
func (s Symbol) String() string {
	return string(s.Bytes())
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
