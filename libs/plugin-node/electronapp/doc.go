// Package electronapp reads the Electron setup of a Node project: the app
// smith makes (Electron Forge driven by its scripts), an adopted Electron Forge
// setup (with or without its Vite plugin), or an adopted electron-builder
// setup. It names the folder the tool packs into, gives the tasks that pack
// the app their link step into the shipped folder dist/<project dir>, and
// makes the presented fix that moves the tool's folder to the work folder
// dist/electron/<project dir> (ADR 0092).
package electronapp
