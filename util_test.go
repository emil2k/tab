package main

import (
	"bytes"
	"errors"
	"go/ast"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/emil2k/tab/lib/diff"
)

// testFiles compares the contents of two files specified by the passed paths,
// the test fails immediateley in case of error if they are not equal.
// The first file is considered the "got" value while the second is considered
// the "expected" value.
func testFiles(t *testing.T, got, expected string) {
	check := func(err error) {
		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}
	}
	gotF, err := os.Open(got)
	check(err)
	defer gotF.Close()
	expectedF, err := os.Open(expected)
	check(err)
	defer expectedF.Close()
	gotContent, err := ioutil.ReadAll(gotF)
	check(err)
	expectedContent, err := ioutil.ReadAll(expectedF)
	check(err)
	if !bytes.Equal(gotContent, expectedContent) {
		changes := diff.Bytes(gotContent, expectedContent)
		for _, c := range changes {
			if c.Del > 0 {
				t.Errorf("diff found, removed :\n%s\n",
					gotContent[c.A:c.A+c.Del])
			}
			if c.Ins > 0 {
				t.Errorf("diff found, inserted :\n%s\n",
					expectedContent[c.B:c.B+c.Ins])
			}
		}
	}
}

// getTestPkg attempt to get a package, in case of errors it fails and
// terminates the test.
func getTestPkg(t *testing.T, dir, pkgName string) *ast.Package {
	pkg, err := getPkg(dir, pkgName)
	if err != nil {
		t.Errorf("error when getting test package %s from %s : %s\n",
			pkgName, dir, err.Error())
		t.FailNow()
	}
	return pkg
}

// getTestDirCopy creates a temporary directory copies the src directory,
// returns path to temporary directory.
// In case of errors it immediately fails the test with a proper message.
func getTestDir(t *testing.T, src string) string {
	temp, err := ioutil.TempDir("", "tabtest")
	if err != nil {
		t.Error("error while creating temp dir :", err.Error())
		t.FailNow()
	}
	err = copyDir(src, temp)
	if err != nil {
		t.Error("error while copying test directory :", err.Error())
		t.FailNow()
	}
	return temp
}

// copyFileJob holds a pending copyFile call.
type copyFileJob struct {
	si       os.FileInfo
	src, dst string
}

// copyDir recursively copies the src directory to the desination directory.
// Creates directories as necessary. Attempts to chmod everything to the src
// mode.
func copyDir(src, dst string) error {
	// First compile a list of copies to execute then execute, otherwise
	// infinite copy situations could arise when copying a parent directory
	// into a child directory.
	cjs := make([]copyFileJob, 0)
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		fileDst := filepath.Join(dst, rel)
		cjs = append(cjs, copyFileJob{info, path, fileDst})
		return nil
	}
	if err := filepath.Walk(src, walk); err != nil {
		return err
	}
	// Execute copies
	for _, cj := range cjs {
		if err := copyFile(cj.si, cj.src, cj.dst); err != nil {
			return err
		}
	}
	return nil
}

// ErrIrregularFile is returned when attempts are made to copy links, pipes,
// devices, and etc.
var ErrIrregularFile = errors.New("non regular file")

// copyFile copies a file or directory from src to dst. Creates directories as
// necessary. Attempts to chmod to the src mode. Returns an error if the file
// is src file is irregular, i.e. link, pipe, or device.
func copyFile(si os.FileInfo, src, dst string) (err error) {
	switch {
	case si.Mode().IsDir():
		return os.MkdirAll(dst, si.Mode())
	case si.Mode().IsRegular():
		closeErr := func(f *os.File) {
			// Properly return a close error
			if cerr := f.Close(); err == nil {
				err = cerr
			}
		}
		sf, err := os.Open(src)
		if err != nil {
			return err
		}
		defer closeErr(sf)
		df, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer closeErr(df)
		// Copy contents
		if _, err = io.Copy(df, sf); err != nil {
			return err
		} else if err = df.Sync(); err != nil {
			return err
		} else {
			return df.Chmod(si.Mode())
		}
	default:
		return ErrIrregularFile
	}
}
