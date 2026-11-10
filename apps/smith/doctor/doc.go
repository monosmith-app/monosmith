// Package doctor checks a workspace's health: smith's own invariants plus
// every plugin's checks. It changes no workspace file (a normal run only saves a
// small summary in .smith/prompt.json for smith prompt); every failed finding names its
// remedy, and the JSON output is a stable contract for editors and CI.
package doctor
