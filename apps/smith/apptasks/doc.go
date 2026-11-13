// Package apptasks is `smith func` and `smith electron`: the steps that the
// tasks of a function app or an Electron app run, in place of the scripts that
// smith once copied into each app. The logic is in the plugin SDK's
// functionhost package, one implementation for Azure Functions, AWS Lambda,
// Google Cloud and Electron. Hidden: the tasks call them.
package apptasks
