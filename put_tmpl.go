package main

var ttTmplString string = `

{{ .Doc }}
func {{ .Name }}(t *testing.T) {
	for i, tt := range {{ .TTIdent }} {
		{{ if .Results }}{{ .Results }} := {{ end }}{{ .CallExpr }}({{ .Params }}){{ range .Checks }}
		if {{ .Got }} != {{ .Expected }} {
			t.Errorf("%d : got %v, expected %v", i, {{ .Got }}, {{ .Expected }})
		}{{ end }}
	}
}{{ if .AppendNewlines }}

{{ end }}
`
