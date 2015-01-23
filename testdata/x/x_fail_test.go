// This file should contain a file that should FAIL declaration validation.
package x

var ttExportedFailFunction = []struct {
	a rune // should be int
	b string
	e error
}{
	{1, "beep", nil},
	{2, "bop", nil},
}

var ttExportedType_ExportedFailMethod = []struct {
	r ExportedType
	a string
	b func() int // should return a bool
	c []int
	e error
}{}
