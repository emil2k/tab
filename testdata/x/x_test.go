//go:generate z
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
	a int
	b string
	e error
}{
	{1, "beep", nil},
	{2, "bop", nil},
}

// tt declaration for a method that does not exist.
var ttExportedType_DoesNotExist = []struct {
	a int
	b string
	e error
}{
	{1, "beep", nil},
	{2, "bop", nil},
}
