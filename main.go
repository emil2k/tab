package main

import (
	"fmt"
	"os"
)

// main gets the GOFILE and GOPACKAGE environment variables set by `go
// generate` and passes them to process.
func main() {
	goFile, goPkg := os.Getenv("GOFILE"), os.Getenv("GOPACKAGE")
	if len(goFile) == 0 || len(goPkg) == 0 {
		fmt.Fprintf(os.Stderr, "tab : command must be called using `go generate`\n")
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "tab : processing file %s in package %s\n", goFile, goPkg)
	n, err := process(goFile, goPkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tab : %s\n", err.Error())
	}
	// Success.
	fmt.Fprintf(os.Stdout, "tab : processed file %s, placed %d table driven test(s)\n",
		goFile, n)
	os.Exit(0)
}

// process processes a file in the given package and returns the number of
// table test placed or an error if there is an issue.
func process(file, pkg string) (int, error) {
	ttDecls, err := fileTTDecls(file, pkg)
	if err != nil {
		return 0, fmt.Errorf("error looking for table test declarations : %s", err.Error())
	}
	// Put the found declarations in the file.
	for _, td := range ttDecls {
		if err := putTTDecl(file, *td); err != nil {
			return 0, fmt.Errorf("error putting table driven test : %s", err.Error())
		}
	}
	return len(ttDecls), nil
}
