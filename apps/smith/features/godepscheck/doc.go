// Package godepscheck is `smith go-deps-check`, the go plugin's deps-check
// task: it fails when a module imports a package from a module its go.mod
// does not require. Hidden: the task calls it through {smith}, people run
// smith run deps-check or smith doctor.
package godepscheck
