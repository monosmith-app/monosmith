package projecttooling

import (
	"fmt"
	"strings"
)

// ViteSettings shape the vite.config.ts that replaces create-react-app.
type ViteSettings struct {
	// OutDir is where the build writes, relative to the project.
	OutDir string
	// Base is the public path the app is served from ("" is "/").
	Base string
	// JSXInJS is true when .js files hold JSX, which create-react-app allows.
	JSXInJS bool
	// ProcessEnv is true when the code reads process.env: Vite then exposes
	// the REACT_APP_ variables there, as create-react-app did.
	ProcessEnv bool
	// Tests is true when the project has tests, which Vitest then runs.
	Tests bool
	// SetupFile is the test setup file, project-relative ("" for none).
	SetupFile string
}

// ViteConfig is the vite.config.ts of a create-react-app project moved onto
// Vite: the React plugin, the build into smith's folder, and what the
// settings ask for.
//
// Parameters: s are the settings.
// Returns: the file's text.
func ViteConfig(s ViteSettings) string {
	var b strings.Builder
	if s.Tests {
		b.WriteString("/// <reference types=\"vitest/config\" />\n")
	}
	b.WriteString("import react from '@vitejs/plugin-react'\nimport { defineConfig } from 'vite'\n\n")
	b.WriteString("// Moved from create-react-app by smith. The build goes to the workspace's dist/.\n")
	b.WriteString("export default defineConfig({\n")
	if s.Base != "" && s.Base != "/" {
		fmt.Fprintf(&b, "  base:    '%s',\n", s.Base)
	}
	b.WriteString("  plugins: [react()],\n")
	if s.ProcessEnv {
		b.WriteString("  envPrefix: 'REACT_APP_',\n  define:    { 'process.env': 'import.meta.env' },\n")
	}
	if s.JSXInJS {
		b.WriteString("  esbuild:   { loader: 'jsx', include: /src\\/.*\\.jsx?$/, exclude: [] },\n")
		b.WriteString("  optimizeDeps: { esbuildOptions: { loader: { '.js': 'jsx' } } },\n")
	}
	fmt.Fprintf(&b, "  build:   {\n    outDir:      '%s',\n    emptyOutDir: true,\n  },\n", s.OutDir)
	if s.Tests {
		b.WriteString("  test: {\n    environment: 'jsdom',\n    globals:     true,\n")
		if s.SetupFile != "" {
			fmt.Fprintf(&b, "    setupFiles:  ['%s'],\n", s.SetupFile)
		}
		b.WriteString("  },\n")
	}
	b.WriteString("})\n")
	return b.String()
}

// ViteIndexHTML is create-react-app's public/index.html as Vite's index.html:
// %PUBLIC_URL% is gone (Vite serves public/ at the root) and the page loads
// the entry module.
//
// Parameters: page is public/index.html; entry is the entry module, from the
// project folder ("/src/index.tsx").
// Returns: the new page.
func ViteIndexHTML(page, entry string) string {
	page = strings.ReplaceAll(page, "%PUBLIC_URL%", "")
	script := `<script type="module" src="` + entry + `"></script>`
	if i := strings.LastIndex(strings.ToLower(page), "</body>"); i >= 0 {
		return page[:i] + "  " + script + "\n" + page[i:]
	}
	return page + script + "\n"
}

// ViteScript is a package.json script with react-scripts replaced by Vite and
// Vitest. keep is false for a script that has no Vite counterpart (eject),
// and for test when the project has no tests.
//
// Parameters: script is the script; tests says the project has tests.
// Returns: the new script and whether it stays.
func ViteScript(script string, tests bool) (string, bool) {
	if !strings.Contains(script, "react-scripts") {
		return script, true
	}
	switch {
	case strings.Contains(script, "react-scripts eject"):
		return "", false
	case strings.Contains(script, "react-scripts test"):
		if !tests {
			return "", false
		}
		return replaceCommand(script, "react-scripts test", "vitest run"), true
	}
	script = strings.ReplaceAll(script, "react-scripts build", "vite build")
	script = strings.ReplaceAll(script, "react-scripts start", "vite")
	return script, !strings.Contains(script, "react-scripts")
}

// replaceCommand swaps a command and drops the arguments that follow it up
// to the next && or end: create-react-app's test flags are Jest's.
func replaceCommand(script, old, with string) string {
	i := strings.Index(script, old)
	rest := script[i+len(old):]
	end := strings.Index(rest, "&&")
	if end < 0 {
		return script[:i] + with
	}
	return script[:i] + with + " " + strings.TrimLeft(rest[end:], " ")
}
