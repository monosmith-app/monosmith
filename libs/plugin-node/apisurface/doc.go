// Package apisurface lists the names a Node package exports, at the working
// tree or at a git revision, so a release can compare two lists and say
// whether the commits ask for a smaller bump than the API change needs.
//
// It reads the entry points package.json declares (exports, types, typings,
// module, main) and parses the files with a pragmatic scanner, not a compiler.
// It never builds: when the built entry (dist/index.d.ts) is not in the
// repository, it reads the source file the built path comes from
// (src/index.ts). Not detected: wildcard subpath exports, a change in a
// signature or a type (only names are compared), destructured and multiple
// declarators of one const, the implicit exports of an ambient .d.ts file,
// and a regular expression literal that holds an unbalanced brace.
package apisurface
