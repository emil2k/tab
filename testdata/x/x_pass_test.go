// This file should contain a file that should PASS declaration validation.
package x

var ttExportedFunction = []struct {
	a int
	b string
	e error
}{
	{1, "beep", nil},
	{2, "bop", nil},
}

var ttExportedType_ExportedMethod = []struct {
	r *ExportedType //  pointer is always allowed
	a string
	b func() bool
	c []int
	e error
}{}
