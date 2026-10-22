package convertplan

import (
	"path"
	"regexp"
	"strings"
)

// AssetTransformFile is the workspace-root Jest transformer that replaces
// @nx/react/plugins/jest.
const AssetTransformFile = "jest.asset-transform.cjs"

// AssetTransform is what @nx/react/plugins/jest did, without Nx: an
// imported SVG becomes a React component named the way SVGR names it, a JS
// or TS file goes to the configured JS transformer, and any other file is
// its file name.
const AssetTransform = `// Written by smith convert: what @nx/react/plugins/jest did, without Nx.
// An imported SVG becomes a React component (named as SVGR names it), a JS or
// TS file goes to @swc/jest or babel-jest, and any other file is its name.
const path = require('node:path')

function className (name) {
  const camel = name
    .replaceAll(/([^a-z0-9])+(.)?/gi, (_, __, chr) => (chr ? chr.toUpperCase() : ''))
    .replaceAll(/[^a-z\d]/gi, '')

  return camel.charAt(0).toUpperCase() + camel.slice(1)
}

function jsTransformer (transforms) {
  const swc = (transforms ?? []).some((t) => Array.isArray(t) && t.some((v) => typeof v === 'string' && v.includes('@swc/jest')))
  const transformer = swc ? require('@swc/jest') : require('babel-jest').default

  return transformer.createTransformer()
}

module.exports = {
  process (src, filename, options) {
    const assetFilename = JSON.stringify(path.basename(filename))
    if (filename.endsWith('.svg')) {
      const componentName = 'Svg' + className(path.parse(filename).name)

      return {
        code: ` + "`" + `const React = require('react');
module.exports = {
  __esModule: true,
  default: ${assetFilename},
  ReactComponent: function ${componentName}(props) {
    return React.createElement('svg', Object.assign({}, props, { children: ${assetFilename} }));
  },
};` + "`" + `,
      }
    }
    const isScript = ['.ts', '.tsx', '.js', '.jsx', '.mjs', '.cjs'].includes(path.extname(filename))

    return isScript
      ? jsTransformer(options.config?.transform).process(src, filename, options)
      : { code: ` + "`module.exports = ${assetFilename};`" + ` }
  },
}
`

// reactBabelPresets are the presets @nx/react/babel stood for under Jest:
// @nx/js/babel's preset-env (for the running Node) and preset-typescript,
// then React's automatic JSX runtime.
const reactBabelPresets = `['@babel/preset-env', { targets: { node: 'current' }, loose: true }], ['@babel/preset-typescript', { allowDeclareFields: true }], ['@babel/preset-react', { runtime: 'automatic', development: true }]`

var (
	reactJestPlugin = regexp.MustCompile(`['"]@nx/react/plugins/jest['"]`)
	reactBabel      = regexp.MustCompile(`['"]@nx/react/babel['"]`)
)

// RewriteNxReactJest rewrites a Jest config that uses @nx/react's
// transformer and Babel preset (dir is its workspace-relative directory):
// the transformer becomes the workspace's AssetTransformFile and the preset
// the Babel presets it stood for. ok is false when anything else of Nx is
// left.
func RewriteNxReactJest(text, dir string) (string, bool) {
	up := strings.Repeat("../", strings.Count(path.Clean(dir), "/")+1)
	if dir == "." {
		up = ""
	}
	out := reactJestPlugin.ReplaceAllLiteralString(text, "'<rootDir>/"+up+AssetTransformFile+"'")
	out = reactBabel.ReplaceAllLiteralString(out, reactBabelPresets)
	if out == text || strings.Contains(out, "@nx/") || strings.Contains(out, "@nrwl/") {
		return text, false
	}
	return out, true
}
