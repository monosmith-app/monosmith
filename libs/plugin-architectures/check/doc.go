// Package check applies an architecture's rules to a project: it places every
// source file and every import of the project's own code in the layer, context
// or module it belongs to, and reports the imports the architecture forbids.
package check
