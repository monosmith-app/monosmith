package tauri

import "strings"

// srcTauriTemplates are the files of a new src-tauri (path relative to it).
// {{name}}, {{lib}}, {{identifier}} and {{dist}} are replaced when written.
var srcTauriTemplates = map[string]string{
	"Cargo.toml": `[package]
name = "{{name}}"
version = "0.0.0"
edition = "2021"

[lib]
name = "{{lib}}_lib"
crate-type = ["staticlib", "cdylib", "rlib"]

[build-dependencies]
tauri-build = { version = "2", features = [] }

[dependencies]
tauri = { version = "2", features = [] }

# The app is built on its own, with its own Cargo.lock and a target directory
# under dist/: an empty workspace keeps a Cargo.toml further up from claiming it.
[workspace]
`,
	"build.rs": `fn main() {
    tauri_build::build()
}
`,
	"src/main.rs": `// Hides the console window on Windows in release builds.
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

fn main() {
    {{lib}}_lib::run()
}
`,
	"src/lib.rs": `#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .run(tauri::generate_context!())
        .expect("error while running the Tauri application");
}
`,
	"capabilities/default.json": `{
  "$schema": "../gen/schemas/desktop-schema.json",
  "identifier": "default",
  "description": "What the main window may do",
  "windows": ["main"],
  "permissions": ["core:default"]
}
`,
	"tauri.conf.json": `{
  "$schema": "https://schema.tauri.app/config/2",
  "productName": "{{name}}",
  "version": "../package.json",
  "identifier": "{{identifier}}",
  "build": {
    "frontendDist": "{{dist}}",
    "devUrl": "http://localhost:5173",
    "beforeDevCommand": "npm run start"
  },
  "app": {
    "windows": [{ "title": "{{name}}", "width": 1000, "height": 700 }],
    "security": { "csp": null }
  },
  "bundle": {
    "active": true,
    "targets": "all",
    "icon": ["icons/32x32.png", "icons/128x128.png", "icons/128x128@2x.png", "icons/icon.icns", "icons/icon.ico"]
  }
}
`,
	".gitignore": `/target/
/gen/
/binaries/
`,
}

// renderSrcTauri returns the src-tauri files for a project, icons included
// (path relative to the project).
func renderSrcTauri(name, workspace, buildDirFromProject string) map[string]string {
	lib := strings.ReplaceAll(segment(name), "-", "_")
	if lib[0] >= '0' && lib[0] <= '9' {
		lib = "app_" + lib
	}
	r := strings.NewReplacer(
		"{{name}}", name,
		"{{lib}}", lib,
		"{{identifier}}", "com."+segment(workspace)+"."+segment(name),
		"{{dist}}", "../"+buildDirFromProject,
	)
	files := map[string]string{}
	for path, body := range srcTauriTemplates {
		files[SrcDir+"/"+path] = r.Replace(body)
	}
	for path, body := range placeholderIcons() {
		files[SrcDir+"/icons/"+path] = body
	}
	return files
}

// segment lowers s to what a bundle identifier and a crate name allow.
func segment(s string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(s) {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' {
			b.WriteRune(c)
		}
	}
	if b.Len() == 0 {
		return "app"
	}
	return b.String()
}
