package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// processFile generates table driven tests inside the file specified by path
// for the specified package name.
func processFile(path, pkgName string) error {
	// Parse file for potential tt identifiers.
	ttIdents, err := fileTTIdents(path)
	if err != nil {
		return err
	}
	// Parse directory to find the functions, types, and methods associated
	// with the table test declarations.
	dir := filepath.Dir(path)
	_, err = dirTTDecls(dir, ttIdents, pkgName)
	if err != nil {
		return err
	}
	return nil
}

// fileTTIdents returns a list of all the identifiers of the potential table
// driven declarations found in the file specified by path.
func fileTTIdents(path string) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), path, nil,
		parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}
	tts := make([]string, 0)
	for _, n := range f.Decls {
		if gd, ok := n.(*ast.GenDecl); !ok {
			continue
		} else if _, ident, ok := isTTVar(gd); ok {
			tts = append(tts, ident)
		}
	}
	return tts, nil
}

// ttDecl holds a table driven test declaration.
type ttDecl struct {
	pkg *ast.Package   // package where the declaration is made
	tt  *ast.ValueSpec // variable declaration that contains tt declaration
	f   *ast.FuncDecl  // function or method declaration to test
	t   *ast.TypeSpec  // type declaration if testing a method

	ttIdent, fIdent, tIdent string
}

// dirTTDecls compiles a list of all the valid tt declarations found in the
// specified diretory that are associated with the passed tt identifiers and
// package name.
func dirTTDecls(dir string, ttIdents []string, pkgName string) ([]*ttDecl, error) {
	pkg, err := getPkg(dir, pkgName)
	if err != nil {
		return nil, err
	}
	return pkgTTDecls(pkg, ttIdents)
}

// pkgTTDecls compiles a list of all the valid tt declarations found in the
// passed package that are associated with the passed tt identifiers.
// Returns an error if any of the found tt declarations are in valid, meaning
// the struct field types don't match the receiver/input/output types.
func pkgTTDecls(pkg *ast.Package, ttIdents []string) ([]*ttDecl, error) {
	ttDecls := make([]*ttDecl, 0)
	for _, ttIdent := range ttIdents {
		if ttDecl, ok := isTTDecl(pkg, ttIdent); ok {
			if err := isTTDeclValid(ttDecl); err != nil {
				return nil, err
			} else {
				ttDecls = append(ttDecls, ttDecl)
			}
		}
	}
	return ttDecls, nil
}

// isTTDecl checks if the identifier is a tt declaration in the provided
// package, if so returns a ttDecl instance with all the necessary AST nodes,
// otherwise returns nil and false.
func isTTDecl(pkg *ast.Package, ttIdent string) (*ttDecl, bool) {
	vs, ok := containsVar(pkg, ttIdent)
	if !ok {
		return nil, false
	}
	ttD := &ttDecl{pkg: pkg, tt: vs, ttIdent: ttIdent}
	// First, attempt to find a function with the name. A function may
	// contain underscore also.
	// Otherwise attempt to find a method.
	ident := strings.TrimPrefix(ttIdent, "tt")
	if fd, ok := containsFunction(pkg, ident); ok {
		ttD.f = fd
		ttD.fIdent = ident
		return ttD, true
	} else {
		// Try to find a method, by trying all the various type and
		// method names that can be inferred from the original ident.
		parts := strings.Split(ident, "_")
		for i := 0; i < len(parts)-1; i++ {
			tIdent := strings.Join(parts[:i+1], "_") // type ident
			mIdent := strings.Join(parts[i+1:], "_") // method ident
			md, ok := containsMethod(pkg, mIdent, tIdent, true)
			if !ok {
				continue
			}
			td, ok := containsType(pkg, tIdent)
			if !ok {
				continue
			}
			ttD.f = md
			ttD.fIdent = mIdent
			ttD.t = td
			ttD.tIdent = tIdent
			return ttD, true
		}
	}
	return nil, false
}

// isTTVar checks if the node is a possible tt declaration, returns the
// ValueSpec, matched identifier, and a bool specifying whether matched. Only
// matches the first identifier in a var declaration, make sure to only declare
// on tt per var.
func isTTVar(gd *ast.GenDecl) (*ast.ValueSpec, string, bool) {
	if gd.Tok != token.VAR {
		return nil, "", false
	}
	for _, sp := range gd.Specs {
		if vs, ok := sp.(*ast.ValueSpec); ok {
			for _, n := range vs.Names {
				if strings.HasPrefix(n.Name, "tt") {
					return vs, n.Name, true
				}
			}
		}
	}
	return nil, "", false
}

// isTTDeclValid returns nil if the tt declaration is valid, otherwise an error.
// Returns an error if the the fields of the test declaration don't match the
// reciever/inputs/outputs of the function or method being tested.
func isTTDeclValid(td *ttDecl) error {
	// Check that tt declaration is a list of structs
	st, ok := isStructSlice(td.tt)
	if !ok {
		return fmt.Errorf("%s should be an array of structs",
			td.ttIdent)
	}
	// Gather expressions
	fes, ses := funcExpr(td.f), structExpr(st)
	if len(fes) != len(ses) {
		return fmt.Errorf("expression count does not match in %s and %s",
			td.ttIdent, td.fIdent)
	}
	for i, fe := range fes {
		if !isTTExprValid(td.pkg, fe, ses[i]) {
			return fmt.Errorf("expressions %T (%v) and %T (%v) don't match in %s and %s",
				fe, fe, ses[i], ses[i], td.fIdent, td.ttIdent)
		}
	}
	return nil
}

// isTTExprValid returns true if the field of the struct declaring the tt test
// properly match the field of the function or method it is testing, both
// expressions must located in the passed package.
// In case of a selector expression, it will import packages as necessary.
// Returns true if the types are the same or if the function has a variadic
// input the struct must have an array of the same type representing the input.
// Returns true if the struct contains a function with no parameters but
// returns the same type as function field, i.e. `func() int` for `int`.
// If the struct expression is an interface then the function must also be an
// interface, and if the function expression is an interface the struct
// expression must meet its requirements.
func isTTExprValid(pkg *ast.Package, funcExpr, structExpr ast.Expr) bool {
	// Function may have variadic input which must be represented by an
	// ast.ArrayType with the same type in the struct.
	if vi, ok := funcExpr.(*ast.Ellipsis); ok {
		if at, ok := structExpr.(*ast.ArrayType); ok {
			if exprEqual(pkg, pkg, vi.Elt, at.Elt) {
				return true
			}
		}
		return false
	}
	if exprEqual(pkg, pkg, funcExpr, structExpr) {
		return true
	}
	// Struct may have a function that returns the necessary type for the
	// function expr, without parameters.
	if ft, ok := structExpr.(*ast.FuncType); ok {
		return ft.Params.NumFields() == 0 &&
			ft.Results.NumFields() == 1 &&
			exprEqual(pkg, pkg, funcExpr, ft.Results.List[0].Type)
	}
	return false
}
