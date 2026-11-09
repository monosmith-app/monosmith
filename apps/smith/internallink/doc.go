// Package internallink lets a project depend on a library of the same
// workspace: it tells a workspace library from a registry package by name,
// refuses a link that cannot work (itself, an app, another language, a
// cycle), asks the user wherever there is a real choice, and has the language
// plugin write the dependency in its own form.
package internallink
