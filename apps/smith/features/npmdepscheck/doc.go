// Package npmdepscheck is `smith npm-deps-check`, the node plugin's
// deps-check task: it fails when a package imports a package its
// package.json does not declare. Hidden: the task calls it through {smith},
// people run smith run deps-check or smith doctor.
package npmdepscheck
