package report

// pageTemplate is the report page. It has no script and no external request:
// fonts are the system fonts, the graph is inline. Every status is a word, and
// the colour only adds to it. The colours follow prefers-color-scheme.
const pageTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light dark">
<title>{{.Workspace}} workspace report</title>
<style>
:root { color-scheme: light dark; --bg: #ffffff; --fg: #111827; --muted: #4b5563; --line: #d1d5db; --panel: #f3f4f6; --link: #1d4ed8; --fail: #b91c1c; --warn: #92400e; --pass: #166534; }
@media (prefers-color-scheme: dark) { :root { --bg: #0b1020; --fg: #e5e7eb; --muted: #9ca3af; --line: #374151; --panel: #151b2e; --link: #93c5fd; --fail: #fca5a5; --warn: #fcd34d; --pass: #86efac; } }
* { box-sizing: border-box; }
body { margin: 0; background: var(--bg); color: var(--fg); font: 16px/1.5 system-ui, -apple-system, "Segoe UI", sans-serif; }
a { color: var(--link); }
:focus-visible { outline: 3px solid var(--link); outline-offset: 2px; }
.skip { position: absolute; left: 8px; top: -48px; background: var(--bg); padding: 8px 12px; border: 2px solid var(--link); }
.skip:focus { top: 8px; }
header, main { max-width: 72rem; margin: 0 auto; padding: 16px; }
h1 { font-size: 1.5rem; margin: 0 0 4px; }
h2 { font-size: 1.25rem; margin: 32px 0 8px; border-bottom: 1px solid var(--line); padding-bottom: 4px; }
p { margin: 8px 0; }
.muted { color: var(--muted); }
nav ul { list-style: none; display: flex; flex-wrap: wrap; gap: 8px 16px; padding: 0; margin: 12px 0 0; }
.scroll { overflow-x: auto; border: 1px solid var(--line); background: var(--panel); padding: 8px; }
.scroll svg { color: var(--fg); fill: var(--fg); max-width: none; }
table { border-collapse: collapse; width: 100%; }
caption { text-align: left; font-weight: 600; padding: 4px 0; }
th, td { text-align: left; vertical-align: top; padding: 6px 8px; border-bottom: 1px solid var(--line); }
th { background: var(--panel); }
.status { font-weight: 700; white-space: nowrap; }
.status.fail { color: var(--fail); border-left: 4px solid var(--fail); padding-left: 6px; }
.status.warn { color: var(--warn); border-left: 4px dashed var(--warn); padding-left: 6px; }
.status.pass { color: var(--pass); border-left: 4px dotted var(--pass); padding-left: 6px; }
summary { cursor: pointer; font-weight: 600; margin: 8px 0; }
code { font: 0.9em ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
ul.plain { padding-left: 20px; }
@media print { .skip, nav { display: none; } .scroll { overflow: visible; } }
</style>
</head>
<body>
<a class="skip" href="#main">Skip to the report</a>
<header>
<h1>{{.Workspace}}</h1>
<p class="muted">Workspace report: {{len .Projects}} projects. Written by smith report.</p>
<nav aria-label="Sections">
<ul>
<li><a href="#map">Workspace map</a></li>
<li><a href="#health">Health</a></li>
<li><a href="#drift">Derived files</a></li>
<li><a href="#choices">Stack and choices</a></li>
<li><a href="#release">Release plan</a></li>
</ul>
</nav>
</header>
<main id="main">

<section aria-labelledby="map">
<h2 id="map" tabindex="-1">Workspace map</h2>
<p>Each box is a project. An arrow points from a project to one it depends on. The box says the kind and the language of the project in words.</p>
{{- if .Cycles}}
<p class="status fail">Fail: import cycles</p>
<ul class="plain">{{range .Cycles}}<li>{{.}}</li>{{end}}</ul>
{{- end}}
<div class="scroll" role="group" aria-label="Dependency graph" tabindex="0">{{.GraphSVG}}</div>
<table>
<caption>The same graph as a table</caption>
<thead><tr><th scope="col">Project</th><th scope="col">Folder</th><th scope="col">Kind</th><th scope="col">Language</th><th scope="col">Depends on</th></tr></thead>
<tbody>
{{- range .Projects}}
<tr><th scope="row">{{.Name}}</th><td><code>{{.Root}}</code></td><td>{{.Kind}}</td><td>{{.Language}}</td><td>{{if .Deps}}{{join .Deps}}{{else}}nothing{{end}}</td></tr>
{{- end}}
</tbody>
</table>
</section>

<section aria-labelledby="health">
<h2 id="health" tabindex="-1">Health</h2>
<p>smith doctor: {{.Health.Failed}} failed, {{.Health.Warned}} warnings, {{.Health.Passed}} passed.</p>
{{- if or .Health.Failed .Health.Warned}}
<table>
<caption>Checks that need attention</caption>
<thead><tr><th scope="col">Status</th><th scope="col">Check</th><th scope="col">Detail</th><th scope="col">Remedy</th></tr></thead>
<tbody>
{{- range .Health.Checks}}{{if ne .Status "Pass"}}
<tr><td><span class="status {{statusClass .Status}}">{{.Status}}</span></td><th scope="row">{{.Name}}</th><td>{{.Detail}}</td><td>{{.Remedy}}</td></tr>
{{- end}}{{end}}
</tbody>
</table>
{{- else}}
<p><span class="status pass">Pass</span> Every check passed.</p>
{{- end}}
{{- if .Health.Passed}}
<details>
<summary>Checks that passed: {{.Health.Passed}}</summary>
<ul class="plain">
{{- range .Health.Checks}}{{if eq .Status "Pass"}}<li><span class="status pass">Pass</span> {{.Name}}{{if .Detail}}: {{.Detail}}{{end}}</li>
{{end}}{{end -}}
</ul>
</details>
{{- end}}
</section>

<section aria-labelledby="drift">
<h2 id="drift" tabindex="-1">Derived files</h2>
{{- if not .Drift.Enabled}}
<p>Derived files are off. Set <code>enabled = true</code> under <code>[derive]</code> in smith.toml to turn them on.</p>
{{- else if .Drift.Error}}
<p><span class="status fail">Fail</span> smith could not build the derive plan: {{.Drift.Error}}</p>
{{- else if .Drift.Items}}
<p>Files that need action: {{len .Drift.Items}}. Files that are current: {{.Drift.Current}}. Run <code>smith derive</code> to see each diff.</p>
<table>
<caption>Derived files that need action</caption>
<thead><tr><th scope="col">File</th><th scope="col">State</th><th scope="col">Owner</th><th scope="col">Renderer</th><th scope="col">Detail</th></tr></thead>
<tbody>
{{- range .Drift.Items}}
<tr><th scope="row"><code>{{.Path}}</code></th><td>{{.Status}}</td><td>{{.Owner}}</td><td><code>{{.Renderer}}</code></td><td>{{.Detail}}</td></tr>
{{- end}}
</tbody>
</table>
{{- else}}
<p><span class="status pass">Pass</span> All {{.Drift.Current}} derived files are current.</p>
{{- end}}
</section>

<section aria-labelledby="choices">
<h2 id="choices" tabindex="-1">Stack and choices</h2>
<table>
<caption>What each project was made from</caption>
<thead><tr><th scope="col">Project</th><th scope="col">Stack</th><th scope="col">Choices</th><th scope="col">Conventions</th></tr></thead>
<tbody>
{{- range .Projects}}
<tr><th scope="row">{{.Name}}</th><td>{{if .Stack}}{{.Stack}}{{else}}none{{end}}</td><td>{{if .Choices}}{{join .Choices}}{{else}}none{{end}}</td><td>{{if .Conventions}}{{join .Conventions}}{{else}}none{{end}}</td></tr>
{{- end}}
</tbody>
</table>
</section>

<section aria-labelledby="release">
<h2 id="release" tabindex="-1">Release plan</h2>
<p>What a merge of this commit to <code>{{.Branch}}</code> would release. smith reads the commits since each package's last tag. It does not ask a registry.</p>
{{- if .Release.Error}}
<p><span class="status fail">Fail</span> smith could not plan the release: {{.Release.Error}}</p>
{{- else if .Release.Items}}
<table>
<caption>Packages that would release</caption>
<thead><tr><th scope="col">Package</th><th scope="col">From</th><th scope="col">To</th><th scope="col">Bump</th><th scope="col">Why</th></tr></thead>
<tbody>
{{- range .Release.Items}}
<tr><th scope="row">{{.Project}}</th><td>{{if .Current}}{{.Current}}{{else}}never released{{end}}</td><td>{{.Next}}</td><td>{{.Level}}</td><td>{{.Reason}}</td></tr>
{{- end}}
</tbody>
</table>
{{- else}}
<p>Nothing would release: no package has a releasing commit since its last tag.</p>
{{- end}}
{{- if .Release.Unchanged}}
<details>
<summary>Packages with nothing to release: {{len .Release.Unchanged}}</summary>
<p>{{join .Release.Unchanged}}</p>
</details>
{{- end}}
</section>

</main>
</body>
</html>
`
