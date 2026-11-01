package legacyjesthelpers

import (
	"regexp"
	"strings"
)

const (
	// Package is the published package that replaces the helper files.
	Package = "@monosmith/jest-preset"

	// ResolverFile and AssetTransformFile are the helper files older converts wrote at the root.
	ResolverFile       = "jest.resolver.cjs"
	AssetTransformFile = "jest.asset-transform.cjs"

	// resolverMarker and assetTransformMarker open the files convert wrote: a
	// file without its marker is the user's own and is never touched.
	resolverMarker       = "// Written by smith convert: what @nx/jest/plugins/resolver did when Jest cannot"
	assetTransformMarker = "// Written by smith convert: what @nx/react/plugins/jest did, without Nx."
)

// legacyPresetValues is the object literal older converts wrote into
// jest.preset.js as the values of @nx/jest/preset, character for character: a
// preset whose values still match it is replaced by the package, one the user
// edited is left for them.
const legacyPresetValues = `{
  testMatch:            ['**/?(*.)+(spec|test).?([mc])[jt]s?(x)'],
  resolver:             require.resolve('./jest.resolver.cjs'),
  moduleFileExtensions: ['ts', 'js', 'mts', 'mjs', 'cts', 'cjs', 'html'],
  coverageReporters:    ['html'],
  moduleNameMapper:     { '^(\\.{1,2}/.*)\\.js$': '$1' },
  transform:            {
    '^.+\\.(ts|js|mts|mjs|cts|cjs|html)$': ['ts-jest', { tsconfig: '<rootDir>/tsconfig.spec.json' }],
  },
  testEnvironment:          'jsdom',
  modulePathIgnorePatterns: ['<rootDir>/dist/', '<rootDir>/out-tsc/'],
  testEnvironmentOptions:   {
    customExportConditions: ['node', 'require', 'default'],
  },
}`

// assetTransformReference is how a project's Jest config named the transformer
// at the root: '<rootDir>/../../jest.asset-transform.cjs', in either quote.
var assetTransformReference = regexp.MustCompile(`(['"])<rootDir>/(?:\.\./)*` + regexp.QuoteMeta(AssetTransformFile) + `(['"])`)

// bom is the byte order mark an editor may put in front of a file.
var bom = string(rune(0xFEFF))

// OwnsResolver reports whether a jest.resolver.cjs is the one convert wrote.
func OwnsResolver(text string) bool {
	return strings.HasPrefix(strings.TrimPrefix(text, bom), resolverMarker)
}

// OwnsAssetTransform reports whether a jest.asset-transform.cjs is the one convert wrote.
func OwnsAssetTransform(text string) bool {
	return strings.HasPrefix(strings.TrimPrefix(text, bom), assetTransformMarker)
}

// RewritePreset returns a jest.preset file whose written-out values (still
// exactly as convert wrote them) are the package instead, and whether it
// found them.
func RewritePreset(text string) (string, bool) {
	return replaceLines(text, func(lf string) (string, bool) {
		if !strings.Contains(lf, legacyPresetValues) {
			return lf, false
		}
		return strings.Replace(lf, legacyPresetValues, "require('"+Package+"')", 1), true
	})
}

// RewriteConfig returns a Jest config whose reference to the root transformer
// names the package's transform instead, and whether it had one.
func RewriteConfig(text string) (string, bool) {
	return replaceLines(text, func(lf string) (string, bool) {
		out := assetTransformReference.ReplaceAllStringFunc(lf, func(match string) string {
			return match[:1] + Package + "/asset-transform" + match[len(match)-1:]
		})
		return out, out != lf
	})
}

// replaceLines runs an edit over text with its line endings normalised to LF
// and gives the result the file's own line endings back.
func replaceLines(text string, edit func(string) (string, bool)) (string, bool) {
	crlf := strings.Contains(text, "\r\n")
	out, changed := edit(strings.ReplaceAll(text, "\r\n", "\n"))
	if !changed {
		return text, false
	}
	if crlf {
		out = strings.ReplaceAll(out, "\n", "\r\n")
	}
	return out, true
}

// NamesHelper reports whether a Jest config still names a helper file.
func NamesHelper(text string) bool {
	return strings.Contains(text, ResolverFile) || strings.Contains(text, AssetTransformFile)
}
