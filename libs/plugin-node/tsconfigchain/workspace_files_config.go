package tsconfigchain

// BaseFile is the content of the root tsconfig.base.json a new workspace gets.
const BaseFile = `{
  "extends": "` + Preset + `"
}
`

// NodeRoot are the root files of a Node project's chain: the base. A stack adds
// them to its template set; a file that exists is never overwritten.
func NodeRoot() map[string]string {
	return map[string]string{Base: BaseFile}
}

// BrowserFile is the content of the root tsconfig.browser.json: the base, for
// code that a bundler builds (a browser app, a VS Code extension, an Astro
// site). It sets what differs from Node: the DOM libraries, the bundler's
// module resolution, no declarations, no rewrite of import extensions (the
// bundler resolves them: a .vue import must stay) and no emit. A project names its own "types".
const BrowserFile = `{
  "extends": "./` + Base + `",
  "compilerOptions": {
    "lib": ["es2023", "dom", "dom.iterable"],
    "module": "esnext",
    "moduleResolution": "bundler",
    "declaration": false,
    "rewriteRelativeImportExtensions": false,
    "noEmit": true,
    "allowImportingTsExtensions": true,
    "types": []
  }
}
`

// BrowserRoot are the root files of a bundled project's chain: the base and the
// browser file that extends it.
func BrowserRoot() map[string]string {
	return map[string]string{Base: BaseFile, Browser: BrowserFile}
}
