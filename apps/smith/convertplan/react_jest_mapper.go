package convertplan

import (
	"regexp"
	"strings"
)

// AssetTransformModule is the Jest transformer of @monosmith/jest-preset that
// replaces @nx/react/plugins/jest.
const AssetTransformModule = JestPresetPackage + "/asset-transform"

// reactBabelPresets are the presets @nx/react/babel stood for under Jest:
// @nx/js/babel's preset-env (for the running Node) and preset-typescript,
// then React's automatic JSX runtime.
const reactBabelPresets = `['@babel/preset-env', { targets: { node: 'current' }, loose: true }], ['@babel/preset-typescript', { allowDeclareFields: true }], ['@babel/preset-react', { runtime: 'automatic', development: true }]`

var (
	reactJestPlugin = regexp.MustCompile(`['"]@nx/react/plugins/jest['"]`)
	reactBabel      = regexp.MustCompile(`['"]@nx/react/babel['"]`)
)

// RewriteNxReactJest rewrites a Jest config that uses @nx/react's
// transformer and Babel preset: the transformer becomes the asset transform
// of @monosmith/jest-preset and the preset the Babel presets it stood for. ok
// is false when anything else of Nx is left.
func RewriteNxReactJest(text string) (string, bool) {
	out := reactJestPlugin.ReplaceAllLiteralString(text, "'"+AssetTransformModule+"'")
	out = reactBabel.ReplaceAllLiteralString(out, reactBabelPresets)
	if out == text || strings.Contains(out, "@nx/") || strings.Contains(out, "@nrwl/") {
		return text, false
	}
	return out, true
}
