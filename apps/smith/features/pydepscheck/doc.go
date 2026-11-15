// Package pydepscheck is `smith py-deps-check`, the python plugin's
// deps-check task: it fails when a package imports a module that none of
// the distributions its pyproject.toml declares provides. Hidden: the task
// calls it through {smith}, people run smith run deps-check or smith doctor.
package pydepscheck
