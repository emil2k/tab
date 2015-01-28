package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"testing"
)

// testsIsTTDeclValid are table tests for isTTDeclValid.
var testsIsTTDeclValid = []struct {
	tt, f, t string // idents
	hasErr   bool   // whether should return error
}{
	{"ttSimpleMatch", "SimpleMatch", "", false},
	{"ttSimpleMatch", "SimpleMisMatch", "", true},
	{"ttAdvancedMatch", "AdvancedMatch", "", false},
	{"ttAdvancedMatch", "AdvancedMisMatch", "", true}, // flips the chan
	{"ttAdvancedMatch", "AdvancedMisMatch2", "", true},
	{"ttReaderMatch", "ReaderMatch", "", false},
	{"ttReaderMatch", "ReaderMisMatch", "", true},
	{"ttReadWriterMatch", "ReadWriterMatch", "", false},
	{"ttReadWriterMatch", "ReadWriterMisMatch", "", true},
	{"ttReaderMatch", "ReaderInterfaceMatch", "", false}, // internal
	{"ttReaderMatch", "ReaderInterfaceMisMatch", "", true},
	{"ttReaderExternalStructMatch", "ReaderInterfaceMatch", "", false},
	{"ttReaderExternalStructMatch", "ReaderInterfaceMisMatch", "", true},
	{"ttReaderExternalInterfaceMatch", "ReaderInterfaceMatch", "", false},
	{"ttReaderExternalInterfaceMatch", "ReaderInterfaceMisMatch", "", true},
	{"ttVariadicMatch", "VariadicMatch", "", false},
	{"ttVariadicMatch", "VariadicMisMatch", "", true},
	{"ttVariadicMatch", "VariadicMisMatchType", "", true},
	{"ttStructFunctionMatch", "StructFunctionMatch", "", false},
	{"ttStructFunctionMatch", "StructFunctionMisMatch", "", true},
	{"ttMethodTypeMatch_MethodValueMatch", "MethodValueMatch", "MethodTypeMatch", false},
	{"ttMethodTypeMatch_MethodValueMatch_Pointer", "MethodValueMatch", "MethodTypeMatch", false},
	{"ttMethodTypeMatch_MethodPointerMatch", "MethodPointerMatch", "MethodTypeMatch", false},
	{"ttMethodTypeMatch_MethodPointerMisMatch", "MethodPointerMatch", "MethodTypeMatch", true},
}

// TestIsTTDeclValid tests the isTTDeclValid function.
func TestIsTTDeclValid(t *testing.T) {
	pkg := getTestPkg(t, "testdata/m", "m")
	for _, td := range testsIsTTDeclValid {
		pre := fmt.Sprintf("tt : %s : f : %s : t : %s", td.tt, td.f, td.t)
		ttDecl := &ttDecl{pkg: pkg, ttIdent: td.tt,
			fIdent: td.f, tIdent: td.t}
		tt, ok := containsVar(pkg, td.tt)
		if !ok {
			t.Error(td.tt, "should contain")
			t.FailNow()
		}
		ttDecl.tt = tt
		if len(td.t) > 0 {
			m, ok := containsMethod(pkg, td.f, td.t, true)
			if !ok {
				t.Error(pre, td.f, "should contain")
				t.FailNow()
			}
			ttDecl.f = m
			xt, ok := containsType(pkg, td.t)
			if !ok {
				t.Error(pre, td.t, "should contain")
				t.FailNow()
			}
			ttDecl.t = xt
		} else {
			f, ok := containsFunction(pkg, td.f)
			if !ok {
				t.Error(pre, td.f, "should contain")
				t.FailNow()
			}
			ttDecl.f = f
		}
		if err := isTTDeclValid(ttDecl); td.hasErr && err == nil {
			t.Error(pre, "expected error, returned nil")
			t.FailNow()
		} else if !td.hasErr && err != nil {
			t.Errorf("%s : not expecting error, returned %v\n",
				pre, err)
			t.FailNow()
		}
	}
}

// genDeclValueWrap put ValueSpecs into the Specs splice in a dummy var GenDecl.
func genDeclValueWrap(vss ...*ast.ValueSpec) *ast.GenDecl {
	ss := make([]ast.Spec, 0)
	for _, vs := range vss {
		ss = append(ss, ast.Spec(vs))
	}
	return &ast.GenDecl{Tok: token.VAR, Specs: ss}
}

// TestIsTTVar tests isTTVar to make sure it matches variable declarations
// starting with "tt", and the negative case.
func TestIsTTVar(t *testing.T) {
	pkg := getTestPkg(t, "testdata/x", "x")
	if n, ok := containsVar(pkg, "ttExportedFunction"); !ok {
		t.Error("should contain")
	} else if vs, ident, ok := isTTVar(genDeclValueWrap(n)); !ok {
		t.Error("should be tt var")
	} else if ident != "ttExportedFunction" {
		t.Error("ident does not match")
	} else if !reflect.DeepEqual(n, vs) {
		t.Error("nodes should equal")
	}
	// Test a variablet that should not match
	if n, ok := containsVar(pkg, "ExportedVar"); !ok {
		t.Error("should contain")
	} else if _, _, ok := isTTVar(genDeclValueWrap(n)); ok {
		t.Error("should not be tt var")
	}
	// Test the false response for a GenDecl that is not a Var.
	if vs, ident, ok := isTTVar(&ast.GenDecl{Tok: token.IMPORT}); ok {
		t.Error("should not be tt var")
	} else if len(ident) > 0 {
		t.Error("should return empty string")
	} else if vs != nil {
		t.Error("should return nil")
	}
}

// TestIsTTDecl tests isTTDecl checking it returns a proper ttDecl for a
// function and a method. Also checks with a tt decl that does not exist and one
// where the declaration exists but not the method it intends to test.
func TestIsTTDecl(t *testing.T) {
	pkg := getTestPkg(t, "testdata/x", "x")
	// Test tt decl for function.
	if tt, ok := isTTDecl(pkg, "ttExportedFunction"); !ok {
		t.Error("should be a tt decl")
	} else {
		xTT := &ttDecl{}
		xTT.pkg = pkg
		xTT.ttIdent = "ttExportedFunction"
		xTT.fIdent = "ExportedFunction"
		xTT.tt, _ = containsVar(pkg, "ttExportedFunction")
		xTT.f, _ = containsFunction(pkg, "ExportedFunction")
		if !reflect.DeepEqual(tt, xTT) {
			t.Error("tt decl not as expected")
		}
	}
	// Test tt decl for method.
	if tt, ok := isTTDecl(pkg, "ttExportedType_ExportedMethod"); !ok {
		t.Error("should be a tt decl")
	} else {
		xTT := &ttDecl{}
		xTT.pkg = pkg
		xTT.ttIdent = "ttExportedType_ExportedMethod"
		xTT.fIdent = "ExportedMethod"
		xTT.tIdent = "ExportedType"
		xTT.tt, _ = containsVar(pkg, "ttExportedType_ExportedMethod")
		xTT.f, _ = containsMethod(pkg, "ExportedMethod", "ExportedType", false)
		xTT.t, _ = containsType(pkg, "ExportedType")
		if !reflect.DeepEqual(tt, xTT) {
			t.Error("tt decl not as expected")
		}
	}
	// Test tt decl that does not exist.
	if _, ok := isTTDecl(pkg, "ttDoesNotExist"); ok {
		t.Error("should not exist")
	}
	// Test tt decl where the declaration exists but the method it intends
	// to test does not exist.
	if _, ok := isTTDecl(pkg, "ttExportedType_DoesNotExist"); ok {
		t.Error("method should not exist")
	}
}

// TestFileTTIdents tests fileTTIdents making sure it retrieves all identifiers
// that may be tt declarations.
func TestFileTTIdents(t *testing.T) {
	idents, err := fileTTIdents("testdata/x/x_pass_test.go")
	if err != nil {
		t.Errorf(err.Error())
	}
	expected := []string{
		"ttExportedFunction",
		"ttExportedType_ExportedMethod",
	}
	if !reflect.DeepEqual(idents, expected) {
		t.Errorf("identifiers %s, expected %s", idents, expected)
	}
}

// TestProcessFile attempts to process file, there should be no errors.
func TestProcessFile(t *testing.T) {
	if _, err := fileTTDecls("testdata/x/x_pass_test.go", "x"); err != nil {
		t.Error("should not get error", err.Error())
	}
	if _, err := fileTTDecls("testdata/x/x_fail_test.go", "x"); err == nil {
		t.Error("should get an error")
	}
}
