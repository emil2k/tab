package main

var ttTmplString string = `

{{ .Doc }}
func {{ .Name }}(t *testing.T) {
	for _, tt := range {{ .TTIdent }} {
		{{ if .Results }}{{ .Results }} := {{ end }}{{ .CallExpr }}({{ .Params }})
{{ range .Checks }}		if {{ .Expected }} != {{ .Got }} {
			t.Errorf("expected %v, got %v\n", {{ .Expected }}, {{ .Got }})
		}{{ end }}
	}
}{{ if .AppendNewlines }}

{{ end }}`
