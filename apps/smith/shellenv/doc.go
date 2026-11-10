// Package shellenv is the port between smith and the machine that tab
// completion needs: the shells smith knows, the file system, the environment
// variables, the operating system and the program that started smith. The real
// adapter is OSEnvironment. The tests of the packages that use it give an
// Environment that works in a temporary folder, so nothing touches the real
// home.
//
// DetectShell finds the shell a person works in: the parent program, then
// PSModulePath (outside Windows), then SHELL, and on Windows PowerShell.
package shellenv
