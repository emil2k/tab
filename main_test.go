package main

import (
	"go/ast"
	"go/token"
	"reflect"
	"testing"
)

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
		xTT.ttIdent = "ttExportedType_ExportedMethod"
		xTT.fIdent = "ExportedMethod"
		xTT.tIdent = "ExportedType"
		xTT.tt, _ = containsVar(pkg, "ttExportedType_ExportedMethod")
		xTT.f, _ = containsMethod(pkg, "ExportedMethod", "ExportedType")
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
	idents, err := fileTTIdents("testdata/x/x_test.go")
	if err != nil {
		t.Errorf(err.Error())
	}
	expected := []string{
		"ttExportedFunction",
		"ttExportedType_ExportedMethod",
		"ttExportedType_DoesNotExist",
	}
	if !reflect.DeepEqual(idents, expected) {
		t.Errorf("identifiers %s, expected %s", idents, expected)
	}
}

// TestProcessFile attempts to process file, there should be no errors.
func TestProcessFile(t *testing.T) {
	if err := processFile("testdata/x/x_test.go", "x"); err != nil {
		t.Errorf(err.Error())
	}
}
