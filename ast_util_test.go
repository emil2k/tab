package main

import (
	"testing"
)

// TestGetPkgNotFound tests that getPkg properly returns not found error.
func TestGetPkgNotFound(t *testing.T) {
	_, err := getPkg("testdata/x", "doesnotexist")
	if err != ErrPkgNotFound {
		t.Error("package does not exist")
	}
}

// TestGetPkg tests that the retrieved package has the expected package name.
func TestGetPkg(t *testing.T) {
	pkg := getTestPkg(t, "testdata/x", "x")
	if pkg.Name != "x" {
		t.Errorf("package name %s, expected testdata/x\n", pkg.Name)
	}
	if pkg.Scope == nil {
		t.Errorf("package scope is nil")
	}
}

// TestContainsFunction tests containsFunction to make sure that it matches
// functions and excludes methods.
// Specifically tests matching of exported and non exported functions.
func TestContainsFunction(t *testing.T) {
	pkg := getTestPkg(t, "testdata/x", "x")
	if fd, ok := containsFunction(pkg, "ExportedFunction"); !ok {
		t.Error("should contain")
	} else if fd.Name.Name != "ExportedFunction" {
		t.Error("ident does not match")
	}
	if _, ok := containsFunction(pkg, "nonExportedFunction"); !ok {
		t.Error("should contain")
	}
	if _, ok := containsFunction(pkg, "doesnotexist"); ok {
		t.Error("should not contain")
	}
	if _, ok := containsFunction(pkg, "ExportedMethod"); ok {
		t.Error("should not match method")
	}
}

// TestContainsMethod tests containsMethod to make sure it does not match
// functions, makes sure that receiver ident matched.
// Specifically tests matching of exported and nonexported methods.
func TestContainsMethod(t *testing.T) {
	pkg := getTestPkg(t, "testdata/x", "x")
	if fd, ok := containsMethod(pkg, "ExportedMethod", "ExportedType"); !ok {
		t.Error("should contain")
	} else if fd.Name.Name != "ExportedMethod" {
		t.Error("ident does not match")
	}
	if _, ok := containsMethod(pkg, "nonExportedMethod", "ExportedType"); !ok {
		t.Error("should contain")
	}
	if _, ok := containsMethod(pkg, "ExportedMethod", "wrongtype"); ok {
		t.Error("method exists but not for this type")
	}
	if _, ok := containsMethod(pkg, "doesnotexist", "ExportedType"); ok {
		t.Error("type exists but not the method")
	}
}

// TestContainsType tests containsType to make sure it matches existing types,
// and does not match variables or functions.
// Specifically tests matching of exported and nonexported types.
func TestContainsType(t *testing.T) {
	pkg := getTestPkg(t, "testdata/x", "x")
	if ts, ok := containsType(pkg, "ExportedType"); !ok {
		t.Error("should contain")
	} else if ts.Name.Name != "ExportedType" {
		t.Error("ident does not match")
	}
	if _, ok := containsType(pkg, "nonExportedType"); !ok {
		t.Error("should contain")
	}
	if _, ok := containsType(pkg, "doesnotexist"); ok {
		t.Error("should not contain")
	}
	if _, ok := containsType(pkg, "A"); ok {
		t.Error("should not match vars")
	}
	if _, ok := containsType(pkg, "ExportedFunction"); ok {
		t.Error("should not match functions")
	}
}

// TestContainsVar tests containsVar to make sure it matches variable, and does
// not match functions, methods, or types.
// Specifically tests matching of exported and nonexported vars.
func TestContainsVar(t *testing.T) {
	pkg := getTestPkg(t, "testdata/x", "x")
	if vs, ok := containsVar(pkg, "ExportedVar"); !ok {
		t.Error("should contain")
	} else if vs.Names[0].Name != "ExportedVar" {
		t.Error("ident does not match")
	}
	if _, ok := containsVar(pkg, "nonExportedVar"); !ok {
		t.Error("should contain")
	}
	if _, ok := containsVar(pkg, "doesnotexist"); ok {
		t.Error("should not contain")
	}
	if _, ok := containsVar(pkg, "ExportedFunction"); ok {
		t.Error("should not match function")
	}
	if _, ok := containsVar(pkg, "ExportedMethod"); ok {
		t.Error("should not match methods")
	}
	if _, ok := containsVar(pkg, "ExportedType"); ok {
		t.Error("should not match types")
	}
}
