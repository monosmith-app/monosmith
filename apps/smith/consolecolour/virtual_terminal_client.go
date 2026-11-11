//go:build !windows

package consolecolour

// Enable does nothing outside Windows: a terminal there renders ANSI codes.
func Enable() {}
