package main

import (
	"go/ast"
	"go/parser"
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
	if fd, ok := containsMethod(pkg, "ExportedMethod", "ExportedType", false); !ok {
		t.Error("should contain")
	} else if fd.Name.Name != "ExportedMethod" {
		t.Error("ident does not match")
	}
	if _, ok := containsMethod(pkg, "nonExportedMethod", "ExportedType", false); !ok {
		t.Error("should contain")
	}
	if _, ok := containsMethod(pkg, "ExportedMethod", "wrongtype", false); ok {
		t.Error("method exists but not for this type")
	}
	if _, ok := containsMethod(pkg, "doesnotexist", "ExportedType", false); ok {
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

// testsStructSliceExpr are table tests for TestStrucSliceExpr.
var testsStructSliceExpr = []struct {
	name     string // var name defining struct
	isStruct bool   // whether expected to be struct
	count    int    // count of expression expected
}{
	{"StructArray", true, 3},
	{"emptyStructArray", true, 4},
	{"NotStructArray", false, 0},
	{"notArray", false, 0},
}

// TestStructSliceExpr tests isStructSlice and strucExpr, checks struct arrays,
// empty struct arrays, non struct arrays, and non arrays to make sure struct
// slice checking works. When a struct type can be retrieved checks that the
// expression count matches with structExpr.
func TestStructSliceExpr(t *testing.T) {
	pkg := getTestPkg(t, "testdata/s", "s")
	for _, tt := range testsStructSliceExpr {
		vs, ok := containsVar(pkg, tt.name)
		if !ok {
			t.Error(tt.name, "should contain")
		}
		if st, ok := isStructSlice(vs); ok != tt.isStruct {
			t.Errorf("%s is struct %t, expected %t\n",
				tt.name, ok, tt.isStruct)
		} else if x := len(structExpr(st)); x != tt.count {
			t.Errorf("%s expr count %d, expected %d\n",
				tt.name, x, tt.count)
		}
	}
}

// TestFuncExpr tests funcExpr checks the count of expressions returned on a
// function an a method, which use a combination of anonymous and named fields
// and cases where there are multiple expr per field.
func TestFuncExpr(t *testing.T) {
	pkg := getTestPkg(t, "testdata/x", "x")
	// Test a function
	f, ok := containsFunction(pkg, "ExportedFunction")
	if !ok {
		t.Error("does not contain")
	}
	test := func(xf *ast.FuncDecl, count int) {
		if x := len(funcExpr(xf)); x != count {
			t.Errorf("expr count does not match %d, expected %d\n",
				x, count)
		}
	}
	test(f, 3)
	// Test a method
	m, ok := containsMethod(pkg, "ExportedMethod", "ExportedType", false)
	if !ok {
		t.Error("does not contain")
	}
	test(m, 5)
}

// testFieldListExpr are table driven tests for fieldListExpr the expected
// outputs are counts of the parameters and the results.
var testsFieldListExpr = []struct {
	expr           string
	param, results int
}{
	{"func(a int, b int) bool", 2, 1},
	{"func(a, b int) bool", 2, 1},
	{"func(a ...int)", 1, 0},
	{"func() bool", 0, 1},
	{"func() (a, b bool)", 0, 2},
	{"func() func(a, b int) bool", 0, 1},
	{"func(a <-chan int) func(a, b int) bool", 1, 1},
}

// TestFieldListExpr test fieldListExpr by simply testing of the counts of expr
// returned match expectations for several functions.
// Tests cases with anonymous returns, variadic inputs, and channels.
func TestFieldListExpr(t *testing.T) {
	for _, tt := range testsFieldListExpr {
		f, err := parser.ParseExpr(tt.expr)
		if err != nil {
			t.Error(tt.expr+" could not be parsed : ", err.Error())
		}
		if ft, ok := f.(*ast.FuncType); !ok {
			t.Errorf(tt.expr+" tt.expr is %T not FuncType\n", f)
		} else {
			if p := fieldListExpr(ft.Params); len(p) != tt.param {
				t.Error(tt.expr + " params count doesn't match")
			}
			if r := fieldListExpr(ft.Results); len(r) != tt.results {
				t.Error(tt.expr + " results count doesn't match")
			}
		}
	}
}
