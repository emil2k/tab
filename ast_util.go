package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// ErrPkgNotFound returned when a package with the provided name is not found in
// a directory.
var ErrPkgNotFound = errors.New("package not found")

// getPkg parses the directory and returns an AST for the specified package
// name, making sure that the Scope attribute is set for the package.
// Is meant to overcome the issue that parser.ParseDir does not set the Scope
// of the package.
// Returns an error if a package with the given name cannot be found in the
// directory or the source cannot be parsed.
func getPkg(dir, pkgName string) (*ast.Package, error) {
	pkgs, err := parser.ParseDir(token.NewFileSet(), dir, nil,
		parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}
	if oPkg, ok := pkgs[pkgName]; ok {
		// Must create a new Package instance, because ParseDir does not
		// set the Scope field.
		// Ignoring error because for some reason ParseDir marks regular
		// types such as int and string unresolved and the NewPackage
		// call attempts to resolve them compiling a list of "undeclared
		// name" errors.
		pkg, _ := ast.NewPackage(token.NewFileSet(), oPkg.Files, nil, nil)
		return pkg, nil
	}
	return nil, ErrPkgNotFound
}

// containsFunction checks the passed packages scope to determine if it
// contains a function with the passed identifier. If so it returns the
// ast.FuncDecl and true, otherwise returns nil and false.
// Relies on the package's Scope to lookup identifiers will panic if it is nil.
// Does not match methods.
func containsFunction(pkg *ast.Package, ident string) (*ast.FuncDecl, bool) {
	if pkg.Scope == nil {
		panic(fmt.Sprintf("package %s scope is nil\n", pkg.Name))
	}
	if obj, ok := pkg.Scope.Objects[ident]; ok && obj.Kind == ast.Fun {
		if fd, ok := obj.Decl.(*ast.FuncDecl); ok && fd.Recv == nil {
			return fd, true
		}
	}
	return nil, false
}

// containsMethod checks the passed packages to determine if it contains a
// method with the passed identifier and passed type identifier. If so it
// returns the ast.FuncDecl and true, otherwise returns nil and false.
// Does not match functions.
func containsMethod(pkg *ast.Package, mIdent, tIdent string) (*ast.FuncDecl, bool) {
	// Methods don't show up in the package scope, so need to walk the AST
	// to find them.
	var fd *ast.FuncDecl
	var walk funcVisitor = func(n ast.Node) {
		if xFd, ok := n.(*ast.FuncDecl); ok && xFd.Name.Name == mIdent && xFd.Recv != nil {
			// Check the receiver type matches
			if xTIdent, ok := xFd.Recv.List[0].Type.(*ast.Ident); ok && xTIdent.Name == tIdent {
				fd = xFd
			}
		}
	}
	ast.Walk(walk, pkg)
	return fd, fd != nil
}

// containsType checks the passed packages scope to determine if it contains a
// type with the passed identifier. If so it returns the ast.TypeSpec and
// true, otherwise returns nil and false.
// Relies on the package's Scope to lookup identifiers will panic if it is nil.
func containsType(pkg *ast.Package, ident string) (*ast.TypeSpec, bool) {
	if pkg.Scope == nil {
		panic(fmt.Sprintf("package %s scope is nil\n", pkg.Name))
	}
	if obj, ok := pkg.Scope.Objects[ident]; ok && obj.Kind == ast.Typ {
		if ts, ok := obj.Decl.(*ast.TypeSpec); ok {
			return ts, true
		}
	}
	return nil, false
}

// containsVar checks the passed packages scope to determine if it contains a
// variable declaration with the passed identifier. If so it returns the
// ast.ValueSpec and true, otherwise returns nil and false.
// Relies on the package's Scope to lookup identifiers will panic if it is nil.
func containsVar(pkg *ast.Package, ident string) (*ast.ValueSpec, bool) {
	if pkg.Scope == nil {
		panic(fmt.Sprintf("package %s scope is nil\n", pkg.Name))
	}
	if obj, ok := pkg.Scope.Objects[ident]; ok && obj.Kind == ast.Var {
		if vs, ok := obj.Decl.(*ast.ValueSpec); ok {
			return vs, true
		}
	}
	return nil, false
}

// funcVisitor defines a simple AST Visitor that calls a function passing in
// the node and returns itself.
type funcVisitor func(n ast.Node)

// Visit calls the receiver function and returns the Visitor back.
func (v funcVisitor) Visit(n ast.Node) ast.Visitor {
	v(n)
	return v
}
