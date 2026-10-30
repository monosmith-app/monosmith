package scaffolding

// templates are every bundle's files. flow.json is the marker that makes the
// directory a project; its content is the author's.
var templates = map[string]string{
	"[[.Dir]]/app.json": `{
  "name": "[[.Name]]",
[[- if .Stack]]
  "stack": "[[.Stack]]",
[[- end]]
  "version": "0.0.0",
  "dependsOn": []
}
`,
	"[[.Dir]]/flow.json": "{}\n",
}
