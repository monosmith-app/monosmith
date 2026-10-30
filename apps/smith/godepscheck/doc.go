// Package godepscheck is `smith go-deps-check`, the go plugin's deps-check
// task: it fails when a module's go.mod does not match what the module
// imports (go mod tidy -diff, or a read-only comparison). Hidden: the task
// calls it through {smith}, people run smith run deps-check or smith doctor.
package godepscheck
