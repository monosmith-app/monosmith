// Package manifestoffer handles a source whose manifest is not at its top,
// for smith import and smith adopt (ADR 0107). It finds the source files the
// language plugins recognise in a folder without a manifest (Sources), asks
// whether to write the manifest of the language's default stack (Settle,
// --manifest answers), and has the plugin write it from its scaffold
// templates (Write, generate with ManifestOnly). It also finds the projects
// whose manifest is below the top of a source (FindNested), so smith import
// can offer the one to import (--path answers).
package manifestoffer
