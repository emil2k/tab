package main

import (
	"fmt"
	"os"
)

func main() {
	// Get the GOFILE and GOPACKAGE environment variables set by
	// `go generate`.
	goFile, goPkg := os.Getenv("GOFILE"), os.Getenv("GOPACKAGE")
	if len(goFile) == 0 || len(goPkg) == 0 {
		fmt.Fprintf(os.Stderr, "tab : command must be called using `go generate`\n")
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "tab : processing file %s in package %s\n", goFile, goPkg)
	ttDecls, err := fileTTDecls(goFile, goPkg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tab : error looking for table test declarations : %s\n", err.Error())
		os.Exit(1)
	}
	// Put the found declarations in the file.
	for _, td := range ttDecls {
		if err := putTTDecl(goFile, *td); err != nil {
			fmt.Fprintf(os.Stderr, "error putting table driven test : %s\n", err.Error())
		}
	}
	// Success.
	fmt.Fprintf(os.Stdout, "tab : processed file %s, placed %d table driven test(s)\n",
		goFile, len(ttDecls))
	os.Exit(0)
}
