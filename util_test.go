package main

import (
	"go/ast"
	"testing"
)

// getTestPkg attempt to get a package, in case of errors it fails and
// terminates the test.
func getTestPkg(t *testing.T, dir, pkgName string) *ast.Package {
	pkg, err := getPkg(dir, pkgName)
	if err != nil {
		t.Errorf("error when getting test package %s from %s : %s\n",
			pkgName, dir, err.Error())
		t.FailNow()
	}
	return pkg
}
