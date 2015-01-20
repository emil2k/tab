package x

import (
	"fmt"
)

func ExportedFunction(a int, b string) error {
	return nil
}

func nonExportedFunction(a int, b string) error {
	return nil
}

var ExportedVar int = 1

var nonExportedVar int = 1

type ExportedType struct{}

type nonExportedType struct{}

func (e ExportedType) ExportedMethod(a, b int) error {
	return fmt.Errorf("nop")
}

func (e ExportedType) nonExportedMethod(a, b int) error {
	return fmt.Errorf("nop")
}
