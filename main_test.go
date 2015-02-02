package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// testCase runs the test case identified by the passed number, it processes the
// file main_test.go in the `a` folder and compares it with the same file in the
// `b` folder. If they don't match it fails the test.
func testCase(t *testing.T, n int) {
	casePath := filepath.Join("testdata", "cases", strconv.Itoa(n))
	tmp := getTestDir(t, filepath.Join(casePath, "a"))
	defer os.RemoveAll(tmp)
	aFile := filepath.Join(tmp, "main_test.go")
	process(aFile, "main")
	bFile := filepath.Join(casePath, "b", "main_test.go")
	testFiles(t, aFile, bFile)
}

// TestExampleCase runs the example test case.
func TestExampleCase(t *testing.T) {
	testCase(t, 1)
}
