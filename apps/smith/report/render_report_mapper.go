package report

import (
	"bytes"
	"html/template"
	"strings"
)

var page = template.Must(template.New("report").Funcs(template.FuncMap{
	"join": func(items []string) string { return strings.Join(items, ", ") },
	"statusClass": func(status string) string {
		switch status {
		case StatusFail:
			return "fail"
		case StatusWarn:
			return "warn"
		}
		return "pass"
	},
}).Parse(pageTemplate))

// Render is the report page for the data. The page is one HTML document with
// inline style and an inline SVG, no script and no external request. It returns
// an error only when the template fails, which a test would catch.
func Render(d Data) (string, error) {
	var b bytes.Buffer
	view := struct {
		Data
		GraphSVG template.HTML
	}{Data: d, GraphSVG: template.HTML(d.GraphSVG)} //nolint:gosec // the SVG is built by graphview, which escapes names
	if err := page.Execute(&b, view); err != nil {
		return "", err
	}
	return b.String(), nil
}
