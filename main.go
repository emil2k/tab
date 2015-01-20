package main

import (
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
	tt *ast.ValueSpec // variable declaration that contains tt declaration
	f  *ast.FuncDecl  // function or method declaration to test
	t  *ast.TypeSpec  // type declaration if testing a method

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
			// TODO validate ttDecl
			ttDecls = append(ttDecls, ttDecl)
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
	ttD := &ttDecl{tt: vs, ttIdent: ttIdent}
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
			md, ok := containsMethod(pkg, mIdent, tIdent)
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

func main() {
	// TODO take the GOFILE and GOPACKAGE and run it through parseFile
}
