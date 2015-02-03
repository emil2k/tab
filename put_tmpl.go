package main

var ttTmplString string = `

{{ .Doc }}
func {{ .Name }}(t *testing.T) {
	for i, tt := range {{ .TTIdent }} {
		{{ if .Results }}{{ .Results }} := {{ end }}{{ .CallExpr }}({{ .Params }}){{ range .Checks }}
		if {{ .Expected }} != {{ .Got }} {
			t.Errorf("%d : expected %v, got %v", i, {{ .Expected }}, {{ .Got }})
		}{{ end }}
	}
}{{ if .AppendNewlines }}

{{ end }}
`
