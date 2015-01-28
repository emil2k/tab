package main

import (
	"go/build"
	"os"
	"path/filepath"
	"testing"
)

// TestPutTTDecl tests putTTDecl function by copying a test package processing
// a file and checking if the package still builds.
// Processes the file twice to see if multiple operations properly replace a
// rendered test function without breaking the code.
func TestPutTTDecl(t *testing.T) {
	pkgPath := getTestDirCopy(t, "testputttdecl",
		filepath.FromSlash("testdata/x/"))
	defer os.RemoveAll(pkgPath)
	file := filepath.Join(pkgPath, "x_pass_test.go")
	tds, err := fileTTDecls(file, "x")
	if err != nil {
		t.Error("error while processing file :", err.Error())
	}
	for _, td := range tds {
		err := putTTDecl(file, *td)
		if err != nil {
			t.Error("error while putting tt decl :", err.Error())
		}
	}
	// Test that the package still builds.
	if _, err := build.ImportDir(pkgPath, 0); err != nil {
		t.Error("error while building processed package :", err)
	}
	// Process one more time to test that nothing breaks.
	tds, err = fileTTDecls(file, "x")
	if err != nil {
		t.Error("error while processing file second time :",
			err.Error())
	}
	for _, td := range tds {
		err := putTTDecl(file, *td)
		if err != nil {
			t.Error("error while putting tt decl second time :",
				err.Error())
		}
	}
	// Test that the package still builds.
	if _, err := build.ImportDir(pkgPath, 0); err != nil {
		t.Error("error while building processed package second time :",
			err)
	}
}
