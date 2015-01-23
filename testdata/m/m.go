// Package m is meant for contains several test cases for testing tt declaration
// validation.
package m

import (
	"bufio"
	"io"
)

// Simple cases

var ttSimpleMatch = []struct {
	n   int
	out bool
}{}

func SimpleMatch(n int) bool {
	return false
}

func SimpleMisMatch() bool {
	return false
}

// Advanced cases

var ttAdvancedMatch = []struct {
	a, b, c int
	f       func(a int, b int) <-chan int
}{}

func AdvancedMatch(a int, b, c int) func(a, b int) <-chan int {
	return nil
}

func AdvancedMisMatch(a int, b, c int) func(a, b int) chan<- int {
	return nil
}

func AdvancedMisMatch2(a int, b, c int) func() <-chan int {
	return nil
}

// Interfaces cases

type SimpleReader int

func (_ SimpleReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

var ttReaderMatch = []struct {
	r SimpleReader
}{}

func ReaderMatch(r io.Reader) {}

func ReaderMisMatch(r io.Writer) {}

// Embedded interfaces, also tests that idents don't have to match to meet
// interface requirements and a type that is not based on a first class type.

type SimpleReadWriter SimpleReader

func (_ SimpleReadWriter) Read(input []byte) (read int, problem error) {
	return 0, nil
}

func (_ SimpleReadWriter) Write(x []byte) (xx int, xxx error) {
	return 0, nil
}

var ttReadWriterMatch = []struct {
	r SimpleReadWriter
}{}

func ReadWriterMatch(r io.ReadWriter) {}

func ReadWriterMisMatch(r io.ReadCloser) {}

// Matching two interfaces, you should be able to pass the types in the struct
// through the function.
// The bufio.ReadWriter and io.ReadWriter have embedded structs and interfaces
// that the program must take into account when looking for methods.
// Also using external package io and and bufio to test that expressions are
// being properly resolved acrosss packages.

var ttReaderExternalInterfaceMatch = []struct {
	r io.ReadWriter
}{}

var ttReaderExternalStructMatch = []struct {
	r *bufio.ReadWriter
}{}

func ReaderInterfaceMatch(r io.Reader) {}

func ReaderInterfaceMisMatch(r io.WriteSeeker) {}

// Variadic input test.

var ttVariadicMatch = []struct {
	n []int
}{}

func VariadicMatch(n ...int) {}

func VariadicMisMatch(n int) {}

func VariadicMisMatchType(n ...string) {}

// Struct contains functions that returns the type.

var ttStructFunctionMatch = []struct {
	in func() bool
}{}

func StructFunctionMatch(f bool) {}

func StructFunctionMisMatch(f int) {}

// Method tests, with some map and struct inputs.

type MethodTypeMatch int

func (m MethodTypeMatch) MethodValueMatch(in map[string]string) {}

func (m *MethodTypeMatch) MethodPointerMatch(in struct{}) {}

var ttMethodTypeMatch_MethodValueMatch = []struct {
	m  MethodTypeMatch
	in map[string]string
}{}

var ttMethodTypeMatch_MethodValueMatch_Pointer = []struct {
	m  *MethodTypeMatch
	in map[string]string
}{}

var ttMethodTypeMatch_MethodPointerMatch = []struct {
	m  *MethodTypeMatch
	in struct{}
}{}

// This should not match as the type MethodTypeMatch won't have the method
// pointer match in its method list.
var ttMethodTypeMatch_MethodPointerMisMatch = []struct {
	m  MethodTypeMatch
	in struct{}
}{}
