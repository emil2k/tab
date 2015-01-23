// Package s contains various declartions of struct arrays for testing purposes.
package s

import (
	"io"
)

var StructArray = []struct {
	a, b int
	c    string
}{
	{1, 2, "boo"},
	{3, 4, "shoe"},
}

var emptyStructArray = []struct {
	a, b int
	r    io.Reader
	f    func() error
}{}

var NotStructArray = []int{1, 2, 3}

var notArray int = 1
