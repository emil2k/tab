# Tab
[![Travis
branch](https://img.shields.io/travis/emil2k/tab.svg?style=flat)](https://travis-ci.org/emil2k/tab)
[![Coverage
Status](https://img.shields.io/coveralls/emil2k/tab.svg?style=flat)](https://coveralls.io/r/emil2k/tab)

**WARNING: This is a work in progress, if you want to help jump in.**

A tool for generating [table driven
tests](https://github.com/golang/go/wiki/TableDrivenTests) in Go.

## Installation

```
go get github.com/emil2k/tab
```

## Compatibility

- Go 1.4+
- Should work on OSX and Linux, someone should test it on Windows.

## Usage

Inititate a variable that holds your table driven tests using the following
naming convention where `F` is a function and `M` is a method on type `T` :

```go
// ttF generates a table driven test for function F.
var ttF = []struct{
	...
}{
	...
}

// ttT_M generates a table driven test for method M of type T.
var ttT_M = []struct{
	...
}{
	...
}
```

All the types and functions specified by `T` and `F` must be located in the same
package as the variable.

The `struct`s representing the test must define fields of the same type as the
inputs and expected outputs of the function or method. The fields should be
ordered with the inputs first and the outputs afterwards mirroring the function
signature. When testing a method of type `T` the first field must be an instance
of the type `T` which will be used as a receiver for the test.

If the function has a variadic input it must be represented as a slice.
Additionally, all the fields can be represented by a function with no parameters
that returns the necessary type, i.e `func() int` for `int`.

Afterwards, add a `go generate` directive to the file for generating the tests :

```go
//go:generate tab
```

To generate the tests in the package directory run :

```
go generate
```

The tool will place or update a test function underneath each table test
variable, with the following naming convention :

```go
// ttF generates a table driven test for function F.
var ttF = []struct{
	...
}{
	...
}

// TestTTF is an automatically generated table driven test for the function F
// using the tests defined in ttF.
func TestTTF(t *testing.T) {
	...
}

// ttT_M generates a table driven test for method M of type T.
var ttT_M = []struct{
	...
}{
	...
}

// TestTTT_M is an automatically generate table driven test for the method T.M
// using the tests defined in ttT_M.
func TestTTT_M(t *testing.T) {
	...
}
```

The generated functions will test that the outputs match expections.

## Example

```go
package main

//go:generate tab

func DummyFunction(a, b int) (c, d, e int, f float64, err error) {
	// dummy function to test
	return
}

var ttDummyFunction = []struct {
	// inputs
	a, b int
	// outputs
	c, d, e int
	f       float64
	err     error
}{
	{1, 2, 5, 6, 7, 5.4, nil},
	{1, 2, 5, 6, 7, 5.4, nil},
	{1, 2, 5, 6, 7, 5.4, nil}, // and on and on ...
}
```

After running `go generate` it adds a table test underneath :

```go
package main

//go:generate tab

func DummyFunction(a, b int) (c, d, e int, f float64, err error) {
	// dummy function to test
	return
}

var ttDummyFunction = []struct {
	// inputs
	a, b int
	// outputs
	c, d, e int
	f       float64
	err     error
}{
	{1, 2, 5, 6, 7, 5.4, nil},
	{1, 2, 5, 6, 7, 5.4, nil},
	{1, 2, 5, 6, 7, 5.4, nil}, // and on and on ...
}

// TestTTDummyFunction is an automatically generated table driven test for the
// function DummyFunction using the tests defined in ttDummyFunction.
func TestTTDummyFunction(t *testing.T) {
	for i, tt := range ttDummyFunction {
		c, d, e, f, err := DummyFunction(tt.a, tt.b)
		if c != tt.c {
			t.Errorf("%d : c : got %v, expected %v", i, c, tt.c)
		}
		if d != tt.d {
			t.Errorf("%d : d : got %v, expected %v", i, d, tt.d)
		}
		if e != tt.e {
			t.Errorf("%d : e : got %v, expected %v", i, e, tt.e)
		}
		if f != tt.f {
			t.Errorf("%d : f : got %v, expected %v", i, f, tt.f)
		}
		if err != tt.err {
			t.Errorf("%d : err : got %v, expected %v", i, err, tt.err)
		}
	}
}
```

## TODO

Improve error messages of the generated tests, can base on the output type :

- Allow naming of expected values with field tags.
- If outputs a struct it could test that each field of the struct matches
  expectations and output errors for individual fields.
- If outputs a map it could test that all the keys match expectations and output
  errors for individual keys. Similar test for arrays/slices.
- If outputs a function it could be provided with another set of table tests
  for the expected value which the returned function must pass for the test to
  pass.

## TODO

Provide flags to prevent replacement of existing function and the option to
place each table test into a separate test function.

## TODO

By default inequality is evaluated using `!=`. Equality can also be determined
by defining a custom function using the following naming convention :

```go
// tt_T determines equality for all fields of type T in this package.
// Returns true if a & b are equal.
func tt_T(a, b T) bool {
	...
}

// ttT_M_X determines equality for the field X, which is of type T and is an
// output of the tests generated by ttT_M.
// Returns true if a & b are equal.
func ttT_M_X(a, b T) bool {
	...
}
```
