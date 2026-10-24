// Package manifestfixes writes the fixes of the Node plugin's package.json
// checks: declare the packages a project imports, install the packages a
// config imports, install the tools a script starts or remove the script. A
// fix edits package.json in place (keys keep their order) and, where
// dependencies change, runs npm; none is safe, because each changes what the
// team declared, so the host asks.
package manifestfixes
