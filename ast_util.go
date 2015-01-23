package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"strconv"
)

// ErrPkgNotFound returned when a package with the provided name is not found in
// a directory.
var ErrPkgNotFound = errors.New("package not found")

// importer is used to import package given the import path, using getPkg.
// Returns ast.Obj of kind pkg with the scope in the Data field and the package
// itself in the Decl field, adds the object to the passed imports map for
// caching.
func importer(imports map[string]*ast.Object, path string) (pkg *ast.Object, err error) {
	if obj, ok := imports[path]; ok {
		return obj, nil
	}
	pkgInfo, err := build.Import(path, "", 0)
	if err != nil {
		return nil, err
	}
	oPkg, err := getPkg(pkgInfo.Dir, pkgInfo.Name)
	if err != nil {
		return nil, err
	}
	oo := ast.NewObj(ast.Pkg, pkgInfo.Name)
	oo.Decl = oPkg
	oo.Data = oPkg.Scope
	imports[path] = oo
	return oo, nil
}

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
		pkg, _ := ast.NewPackage(token.NewFileSet(), oPkg.Files,
			nil, nil)
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
// Depending on the whether `pointer` is set determines whether to include
// methods with the `*tIdent` receiver in the search, instead of just `tIdent`
// receivers.
// Does not match functions.
func containsMethod(pkg *ast.Package, mIdent, tIdent string, pointer bool) (*ast.FuncDecl, bool) {
	ts, ok := containsType(pkg, tIdent)
	if !ok {
		return nil, false
	}
	// Methods don't show up in the package scope, so need to walk the AST
	// to find them.
	var fd *ast.FuncDecl
	walk := func(p *ast.Package, mid, tid string, wpointer bool) funcVisitor {
		return func(n ast.Node) {
			if xFd, ok := n.(*ast.FuncDecl); ok && xFd.Name.Name == mid && xFd.Recv != nil {
				// Check the receiver name & whether pointer
				// matches.
				_, xTs, xPointer, _ := resolveExpr(p, xFd.Recv.List[0].Type)
				if (wpointer || (!wpointer && !xPointer)) &&
					xTs != nil && xTs.Name.Name == tid {
					fd = xFd
				}
			}
		}
	}
	ast.Walk(walk(pkg, mIdent, tIdent, pointer), pkg)
	if fd != nil {
		return fd, true
	}
	// Check embeded fields in a struct type, method could have them as a
	// receiver.
	if st, ok := ts.Type.(*ast.StructType); ok && st.Fields != nil &&
		st.Fields.NumFields() > 0 {
		for _, f := range st.Fields.List {
			if len(f.Names) != 0 {
				continue
			}
			ePkg, eTs, ePointer, _ := resolveExpr(pkg, f.Type)
			// Check for method in embedded type's package with it
			// as a receiver.
			if eTs != nil {
				ast.Walk(walk(ePkg, mIdent, eTs.Name.Name, ePointer), ePkg)
				if fd != nil {
					return fd, true
				}
			}
		}
	}
	return nil, false
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

// exprEqual returns true if the types of the two expressions match, checks
// idents, function signatures, channels, and interface types.
// If expression `b` is an interface then `a` must also be an interface, and if
// expression `a` is an interface then interface `b` must meet its requirements.
// Resolves expressions in the passed corresponding packages.
// If expression `a` is an ast.StarExpr then `b` must also be an ast.StarExpr
// pointing to an equivalent type, otherwise `b` can be either.
func exprEqual(ap, bp *ast.Package, a, b ast.Expr) bool {
	if a == nil || b == nil {
		return a == b
	}
	// Attempt to resolve idents and selector expressions, update the
	// respective packages if necessary.
	ap, _, apoint, ao := resolveExpr(ap, a)
	bp, bts, bpoint, bo := resolveExpr(bp, b)
	// If a is `a` is a pointer expression, then `b` must also be.
	if apoint && !bpoint {
		return false
	}
	switch at := ao.(type) {
	case *ast.MapType:
		if bt, ok := bo.(*ast.MapType); ok {
			return exprEqual(ap, bp, at.Key, bt.Key) &&
				exprEqual(ap, bp, at.Value, bt.Value)
		}
	case *ast.StructType:
		if bt, ok := bo.(*ast.StructType); ok {
			return fieldListEqual(ap, bp, at.Fields, bt.Fields)
		}
	case *ast.ArrayType:
		if bt, ok := bo.(*ast.ArrayType); ok {
			return exprEqual(ap, bp, at.Elt, bt.Elt) &&
				exprEqual(ap, bp, at.Len, bt.Len)
		}
	case *ast.InterfaceType:
		return exprInterface(ap, bp, at, bts, false)
	case *ast.FuncType:
		if bt, ok := bo.(*ast.FuncType); ok {
			return fieldListEqual(ap, bp, at.Params, bt.Params) &&
				fieldListEqual(ap, bp, at.Results, bt.Results)
		}
	case *ast.ChanType:
		if bt, ok := bo.(*ast.ChanType); ok {
			return at.Dir == bt.Dir && exprEqual(ap, bp, at.Value, bt.Value)
		}
	case *ast.Ident:
		// For matching all other first class types, i.e. intX, floatX,
		// complexX, byte, rune.
		if bt, ok := bo.(*ast.Ident); ok &&
			at.Name == bt.Name {
			return true
		}
	default:
		panic(fmt.Sprintf("unhandled type %T (%v), compared to %T (%v)\n",
			ao, ao, bo, bo))
	}
	return false
}

// exprInterface return true if the the passed type meets the requirements of
// the interface within their respective packages.
// If provided type is an interface it must include the methods in the passed
// interface.
// If `pointer` is set it includes methods with pointer receivers.
func exprInterface(ifacePkg, tsPkg *ast.Package, iface *ast.InterfaceType, ts *ast.TypeSpec, pointer bool) bool {
	if iface == nil || iface.Methods.NumFields() == 0 {
		return true
	}
	for _, m := range iface.Methods.List {
		mPkg, _, _, mObj := resolveExpr(ifacePkg, m.Type)
		switch x := mObj.(type) {
		case *ast.InterfaceType:
			// Embedded interface found, should have its methods
			// aswell.
			if ok := exprInterface(mPkg, tsPkg, x, ts, pointer); !ok {
				return false
			}
		case *ast.FuncType:
			for _, n := range m.Names {
				// Find the potential method declaration,
				// depending on whether the passed type is an
				// interface or another type.
				var md *ast.FuncType
				if tsi, ok := ts.Type.(*ast.InterfaceType); ok {
					if md, ok = ifaceContainsMethod(tsPkg, tsi, n.Name); !ok {
						return false
					}
				} else {
					if mdecl, ok := containsMethod(tsPkg, n.Name, ts.Name.Name, pointer); ok {
						md = mdecl.Type
					}
				}
				if md == nil {
					// Corresponding method not found
					return false
				}
				// Check that the method signatures match
				if mm := exprEqual(ifacePkg, tsPkg, x, md); !mm {
					return false
				}
			}
		default:
			panic(fmt.Sprintf("unhandled interface field %T\n",
				m.Type))
		}
	}
	return true
}

// ifaceContainsMethod checks if the interface contains a method with the passed
// name, if not returns nil and false.
// Takes into account embedded interfaces.
func ifaceContainsMethod(pkg *ast.Package, iface *ast.InterfaceType, name string) (*ast.FuncType, bool) {
	for _, m := range iface.Methods.List {
		mPkg, _, _, mObj := resolveExpr(pkg, m.Type)
		switch x := mObj.(type) {
		case *ast.InterfaceType:
			// Embedded interface found, check if it contains the
			// method.
			if xf, ok := ifaceContainsMethod(mPkg, x, name); ok {
				return xf, true
			}
		case *ast.FuncType:
			for _, n := range m.Names {
				if n.Name == name {
					return x, true
				}
			}
		}
	}
	return nil, false
}

// resolveExpr attempts to resolve expressions such as an ident or selector
// expression down to their underlying type.
// Returns the type spec and the underlying object XXXType, returns in nil when
// cannot resolve something or if there is no type spec for the type, i.e. int,
// float32, etc.
// Returns whether the type spec is a pointer type.
// Returns the package where the type spec was found, may change if there is
// a selector expression.
func resolveExpr(pkg *ast.Package, in ast.Expr) (resolvedPackage *ast.Package, typeSpec *ast.TypeSpec, pointer bool, obj interface{}) {
	if x, ok := in.(*ast.StarExpr); ok {
		in = x.X
		pointer = true
	}
	switch x := in.(type) {
	case *ast.Ident:
		resolvedPackage, typeSpec, obj = resolveIdent(pkg, x)
	case *ast.SelectorExpr:
		resolvedPackage, typeSpec, obj = resolveSelectorExpr(pkg, x)
	default:
		obj = in
	}
	return
}

// resolveSelectorExpr attempts to resolve a selector expression to its
// underlying type assuming the selector is a package identifier and that the
// field ident is in its scope. Imports a package if necessary.
// Returns the input and the passed package if cannot resolve.
// Returns the package where the type spec was found.
func resolveSelectorExpr(pkg *ast.Package, in *ast.SelectorExpr) (*ast.Package, *ast.TypeSpec, interface{}) {
	if xi, ok := in.X.(*ast.Ident); ok {
		// Attempt to lookup the idents obj in case it is a package,
		// otherwise will have to try to import it.
		selObj, ok := objPkgLookup(xi.Obj, in.Sel.Name)
		if ok {
			return resolveObjDecl(pkg, selObj)
		}
		// Find the file the selector is in, attempt to find the
		// package it refers to, import it, and lookup the
		// object in it.
		if f, ok := lookupFile(pkg, in); ok {
			if pkgObj, ok := lookupImport(pkg, f, xi.Name); ok {
				if selPkg, ok := pkgObj.Decl.(*ast.Package); ok {
					selObj, _ = objPkgLookup(pkgObj, in.Sel.Name)
					return resolveObjDecl(selPkg, selObj)
				}
			}
		}
	}
	return pkg, nil, in
}

// resolveIdent attempts to resolve an ident expression into it's underlying
// type.
// Returns the input and the passed package if cannot resolve.
// Returns the package where the type spec was found, may change if there is
// a selector expression.
func resolveIdent(pkg *ast.Package, in *ast.Ident) (*ast.Package, *ast.TypeSpec, interface{}) {
	if npkg, ts, decl := resolveObjDecl(pkg, in.Obj); decl != nil {
		return npkg, ts, decl
	}
	return pkg, nil, in
}

// resolveObjDecl resolves the declaration of an object, returning the type
// spec of the object and the declaration of the underlying object, i.e.
// XXXTypes. Returns a nil declaration if the object has no declaration, i.e.
// int, float32, etc., and returns a nil type spec and the passed package when
// the object does not refer to an ast.TypeSpec.
// Returns the package where the type spec was found, may change if there is
// a selector expression.
func resolveObjDecl(pkg *ast.Package, obj *ast.Object) (*ast.Package, *ast.TypeSpec, interface{}) {
	if obj != nil && obj.Decl != nil {
		if ts, ok := obj.Decl.(*ast.TypeSpec); ok {
			// Recurse to the underlying type, but return this type
			// spec and package.
			_, _, _, robj := resolveExpr(pkg, ts.Type)
			return pkg, ts, robj
		}
		return pkg, nil, obj.Decl // no type spec found
	}
	return nil, nil, nil
}

// lookupImport attempts to import the package refered to by the passed selector
// depending on the file. If found it returns an ast.Object of the package kind
// and true, otherwise nil and false.
// It must import the package to determine the package name, if pkg is not nil
// it uses its Imports field as a cache so it won't have to import repeatedly.
func lookupImport(pkg *ast.Package, file *ast.File, sel string) (*ast.Object, bool) {
	var imports map[string]*ast.Object
	if pkg != nil {
		imports = pkg.Imports
	}
	for _, i := range file.Imports {
		if i.Name != nil && (i.Name.Name == "." || i.Name.Name == "_") {
			continue
		}
		ip, err := strconv.Unquote(i.Path.Value)
		if err != nil {
			continue
		}
		pkgObj, err := importer(imports, ip)
		if err != nil {
			continue
		}
		// Must match either package name or local package name.
		if (i.Name != nil && i.Name.Name == sel) ||
			(i.Name == nil && pkgObj.Name == sel) {
			return pkgObj, true
		}
	}
	return nil, false
}

// lookupFile attempts to locate the file in the package where the AST node
// is defined, by walking the passed package.
func lookupFile(pkg *ast.Package, node ast.Node) (*ast.File, bool) {
	found := false
	var walk funcVisitor = func(n ast.Node) {
		if n == node {
			found = true
		}
	}
	for _, f := range pkg.Files {
		ast.Walk(walk, f)
		if found {
			return f, true
		}
	}
	return nil, false
}

// objPkgLookup attempts to retrieve the named object from a package scope,
// if the passed object is not a pkg object then it will return nil and false.
func objPkgLookup(pkg *ast.Object, name string) (*ast.Object, bool) {
	if pkg == nil || pkg.Kind != ast.Pkg || pkg.Data == nil {
		return nil, false
	}
	if pkgScope, ok := pkg.Data.(*ast.Scope); ok {
		if obj, ok := scopeLookup(pkgScope, name); ok {
			return obj, true
		}
	}
	return nil, false
}

// scopeLookup recursively lookups an object with the given name in the provided
// scope, recurses to outer scopes when not found.
// Returns nil and false when not found.
func scopeLookup(scope *ast.Scope, name string) (*ast.Object, bool) {
	for ; scope != nil; scope = scope.Outer {
		if obj := scope.Lookup(name); obj != nil {
			return obj, true
		}
	}
	return nil, false
}

// fieldListEqual returns true if the two field list are equivalent.
// Accounts for the fact that fields can declare types on an individual basis or
// on multiple idents but the two field lists should be considered equivalent.
// For example : `a, b int` and `a int, b int` should be considered the same.
func fieldListEqual(ap, bp *ast.Package, a, b *ast.FieldList) bool {
	if a.NumFields() != b.NumFields() {
		return false
	}
	bes := fieldListExpr(b)
	for i, ae := range fieldListExpr(a) {
		if !exprEqual(ap, bp, ae, bes[i]) {
			return false
		}
	}
	return true
}

// isStructSlice checks if the value spec is an array of structs, if so returns
// the struct type and true, otherwise returns nil and false.
// Only checks the first value in var declaration should not be used with
// multi assigning initilizations.
func isStructSlice(vs *ast.ValueSpec) (*ast.StructType, bool) {
	if len(vs.Values) > 0 {
		if cl, ok := vs.Values[0].(*ast.CompositeLit); ok {
			if at, ok := cl.Type.(*ast.ArrayType); ok {
				if st, ok := at.Elt.(*ast.StructType); ok {
					return st, true
				}
			}
		}
	}
	return nil, false
}

// structExpr returns a list of expressions representing the fields of a struct
// type.
// If mutiple idents are provided for a type in specifying the fields, i.e.
// a, b int, then provides an expression for each ident.
func structExpr(s *ast.StructType) []ast.Expr {
	oe := make([]ast.Expr, 0)
	if s == nil || s.Fields == nil {
		return oe
	}
	for _, f := range s.Fields.List {
		for i := 0; i < len(f.Names); i++ {
			oe = append(oe, f.Type)
		}
	}
	return oe
}

// funcExpr compiles a list of expressions in the signature of a function or
// method, in the following order receiver, inputs, outputs.
// If multiple ident are provided for a type in the signature, i.e. a, b int,
// then provides an expression for each ident.
func funcExpr(f *ast.FuncDecl) []ast.Expr {
	fs := make([]*ast.Field, 0)
	if f.Recv != nil {
		fs = append(fs, f.Recv.List...)
	}
	fs = append(fs, f.Type.Params.List...)
	if f.Type.Results != nil {
		fs = append(fs, f.Type.Results.List...)
	}
	oe := make([]ast.Expr, 0)
	for _, f := range fs {
		if len(f.Names) == 0 { // anonymous fields
			oe = append(oe, f.Type)
			continue
		}
		for i := 0; i < len(f.Names); i++ {
			oe = append(oe, f.Type)
		}
	}
	return oe
}

// fieldListExpr compiles a list of expressions in a field list.
// If multiple ident are provided for a type in the signature, i.e. a, b int,
// then provides an expression for each ident.
func fieldListExpr(f *ast.FieldList) []ast.Expr {
	oe := make([]ast.Expr, 0)
	if f == nil {
		return oe
	}
	for _, f := range f.List {
		if len(f.Names) == 0 { // anonymous field
			oe = append(oe, f.Type)
			continue
		}
		for _ = range f.Names {
			oe = append(oe, f.Type)
		}
	}
	return oe
}
