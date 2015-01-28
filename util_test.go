package main

import (
	"errors"
	"go/ast"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

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

// getTestDirCopy creates a temporary directory copies the src directory into it
// and returns its path. Prefix is used is prefixed to the temporary directory.
// User is responsible for removing the directory after use.
// In case of errors it immediately fails the test with a proper message.
func getTestDirCopy(t *testing.T, prefix, src string) string {
	temp, err := ioutil.TempDir("", prefix)
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
// With the opt.verbose option set outputs the src and destination of each
// copied file.
// Skips hidden files base on the opt.hidden option.
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
