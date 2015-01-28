package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

// putTTDecl either updates or creates the test function described by the passed
// tt declaration under the variable that declares it, in the file specified by
// the path.
func putTTDecl(path string, td ttDecl) error {
	// Slurp the file with ReadAll.
	content, err := slurpFile(path)
	if err != nil {
		return err
	}
	// Find old test function declaration and if necessary remove it.
	fs, f, err := parseBytes(content)
	if err != nil {
		return err
	}
	rmStart, rmEnd, ok := funcDeclRange(fs, f, td.testName())
	if ok {
		content = replaceRange(content, []byte{}, rmStart, rmEnd)
		// Need to update the AST and fileset, because content has
		// changed and it needs to be used for determine append range.
		fs, f, err = parseBytes(content)
		if err != nil {
			return err
		}
	}
	// Find range where to place the new test declaration.
	appendStart, appendEnd, appendEOF, ok := appendRange(fs, f, content, td.ttIdent)
	if !ok {
		return fmt.Errorf("%s not found in file %s", td.ttIdent, path)
	}
	// Template out the test function from the declaration.
	tdh, err := newTTHolder(td, !appendEOF)
	if err != nil {
		return err
	}
	testContent := renderTTTestFunction(*tdh)
	content = replaceRange(content, testContent, appendStart, appendEnd)
	// Write the new file to disk.
	if err := writeFile(path, content); err != nil {
		return err
	}
	return nil
}

// writeFile writes out the file to the given path, truncates the file if
// necessary.
// Returns an error if there is an issue with opening or writing to the path.
func writeFile(path string, content []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0755) // TODO mirror permissions
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write(content)
	return nil
}

// slurpFile opens up and read the full content of the specified path.
func slurpFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return content, err
}

// parseBytes parses the AST of a file from a byte slice, includes and returns
// all errors.
func parseBytes(in []byte) (*token.FileSet, *ast.File, error) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "DOESNTMATTER", in,
		parser.AllErrors|parser.ParseComments)
	return fs, f, err
}

// replaceRange replaces the range specified by the start/end offset with the
// sub slice in the input slice.
func replaceRange(in []byte, sub []byte, start, end int) []byte {
	out := make([]byte, 0, len(in)+len(sub)+start-end)
	out = append(out, in[:start]...)
	out = append(out, sub...)
	out = append(out, in[end:]...)
	return out
}

// funcDeclRange if the a func declaration exists in the file with the specified
// ident provides the offset range where it resides, including documentation
// comments (adjacent to declaration).
// If a func declaration is not found returns false for the third result.
func funcDeclRange(fs *token.FileSet, f *ast.File, ident string) (start, end int, ok bool) {
	if obj := f.Scope.Lookup(ident); obj != nil {
		if fd, ok := obj.Decl.(*ast.FuncDecl); ok {
			sp, ep := fd.Pos(), fd.End()
			// Determine where the functions documentation begins.
			for _, c := range fd.Doc.List {
				if c.Pos() < sp {
					sp = c.Pos()
				}
			}
			return fs.PositionFor(sp, true).Offset,
				fs.PositionFor(ep, true).Offset, true
		}
	}
	return 0, 0, false
}

// appendRange finds the node with the specified ident in the file's scope and
// returns the range of offsets that includes adjacent whitespace that would
// need to be replaced to append something right after it.
// Returns whether the range reaches the end of file, this is important when
// deciding whether to add whitespace after the node.
// The contents of the file should be passed via src, must be the same size,
// used by Scanner to find the extent of the whitespace up to the next keyword
// or comment.
func appendRange(fs *token.FileSet, f *ast.File, src []byte, ident string) (start, end int, eof, ok bool) {
	if obj := f.Scope.Lookup(ident); obj != nil {
		if n, ok := obj.Decl.(ast.Node); ok {
			sp, ep := n.End(), n.End()
			// Scan file to find where the whitespace ends.
			s := new(scanner.Scanner)
			tf := fs.File(sp)
			s.Init(tf, src, nil, scanner.ScanComments)
			for {
				tp, tok, _ := s.Scan()
				if tp < ep {
					continue
				} else if tp > ep {
					if tok == token.EOF {
						eof = true
						ep = tp
						break
					}
					if tok == token.COMMENT || tok.IsKeyword() {
						ep = tp
						break
					}
				}
			}
			return fs.PositionFor(sp, true).Offset,
				fs.PositionFor(ep, true).Offset, eof, true
		}
	}
	return 0, 0, false, false
}

// ttTmpl holds the table test template used for generating tests.
var ttTmpl = template.Must(template.New("tt").Parse(ttTmplString))

// renderTTTestFunction generates the code for the table test function from the
// template holder.
func renderTTTestFunction(tdh ttHolder) []byte {
	buf := new(bytes.Buffer)
	err := ttTmpl.Execute(buf, tdh)
	if err != nil {
		panic(fmt.Sprintf("rendering table test function : %v", err))
	}
	testContent, _ := ioutil.ReadAll(buf)
	return testContent
}

// ttHolder is a holder to provide to the template engine variables necessary to
// output a table test.
type ttHolder struct {
	Name            string
	CallExpr        string // expression for calling function or method
	TTIdent         string // identifier for the structs slice to range over
	Doc             string // docstring for the test function.
	Params, Results string
	Checks          []ttCheck
	AppendNewlines  bool // whether reaches EOF
}

// ttCheck is a holder to provide to the template engine variables necessary to
// output a check that a value received for a result matches the expected value.
type ttCheck struct {
	Expected, Got string
}

// newTTHolder initiates the variables necessary to render a table test, returns
// a ttHolder. The appendNewLines is used to determine whether new lines need to
// be attached after the test function.
func newTTHolder(td ttDecl, appendNewLines bool) (*ttHolder, error) {
	name := td.testName()
	i := 0
	// Get the struct slide and compile a list of its fields.
	tds, ok := isStructSlice(td.tt)
	if !ok {
		return nil, fmt.Errorf("%s is not a struct slice\n", td.ttIdent)
	}
	var fields []string
	for _, s := range tds.Fields.List {
		for _, n := range s.Names {
			fields = append(fields, n.Name)
		}
	}
	// Determine the function or method expression.
	var ident string
	if len(td.tIdent) > 0 {
		ident = fmt.Sprintf("%s.%s", fields[0], td.fIdent)
		i++
	} else {
		ident = td.fIdent
	}
	// Determine expressions for the function/method parameters, results,
	// and the equivalence checks.
	var params, results []string
	for _, p := range td.f.Type.Params.List {
		switch p.Type.(type) {
		case *ast.Ellipsis:
			for range p.Names {
				params = append(params, fmt.Sprintf("tt.%s...", fields[i]))
				i++
			}
		default:
			for range p.Names {
				params = append(params, fmt.Sprintf("tt.%s", fields[i]))
				i++
			}
		}
	}
	var checks []ttCheck
	addResult := func() {
		field := fields[i]
		checks = append(checks, ttCheck{fmt.Sprintf("tt.%s", field), field})
		results = append(results, field)
		i++
	}
	for _, r := range td.f.Type.Results.List {
		if len(r.Names) == 0 { // unnamed return
			addResult()
			continue
		}
		for range r.Names {
			addResult()
		}
	}
	return &ttHolder{
		name,
		ident,
		td.ttIdent,
		renderComment(td.testDoc()),
		strings.Join(params, ", "),
		strings.Join(results, ", "),
		checks,
		appendNewLines,
	}, nil
}

// renderComment returns a comment string with a new line roughly every 80
// characters, without splitting up words. At the start of each new line adds
// a "//" to make it a comment. No newline is added at the end.
func renderComment(str string) string {
	if len(str) == 0 {
		return ""
	}
	comment := "// "
	out := len(comment)
	words := strings.Split(str, " ")
	for i, w := range words {
		comment += w
		out += len(w)
		if len(words) > i+1 {
			if out+len(words[i+1]) > 77 {
				comment += "\n// "
				out = 0
			} else {
				// Avoid trailing white space.
				comment += " "
				out++
			}
		}
	}
	return comment
}
