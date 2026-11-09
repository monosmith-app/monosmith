// Package report is `smith report`: one self-contained HTML file with the
// workspace map (an inline SVG dependency graph), the doctor results, the
// derived-file drift, the choices of each project and the release plan of the
// current commit. The file makes no request and runs no script, so it can be
// attached to a pull request. See ADR 0048.
package report
