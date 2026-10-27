// Package releaseassets holds what every language plugin that builds an app's
// release archives does the same way: filling the file name pattern, writing
// a zip with the program and its README and licence, and the SHA-256 lines of
// a checksums file. What differs per language (how the program is built for
// a platform) stays in the plugin.
package releaseassets
