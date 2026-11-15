// Package declarationspecifiers gives the relative specifiers in a library's
// generated declarations the file they name (`./hooks` becomes `./hooks.js`,
// or `./hooks/index.js` for a folder), which nodenext consumers of an ES
// module package need to follow the types. A converted build runs it after
// tsc as `smith declaration-specifiers <dir>`, replacing the script smith
// convert used to leave at the workspace root.
package declarationspecifiers
