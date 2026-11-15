// Package pluginmarket finds plugins by name in a marketplace index: a JSON
// file (the repository monosmith-app/plugins by default) that lists, for each
// plugin, its description, roles, kinds, languages and the versions with the
// URL and SHA-256 of their packed archive (decision 0040). It searches the
// index and turns "name" or "name@version" into an archive and a digest for
// plugininstall, which then checks, installs and never trusts it. Nothing here
// downloads or starts a plugin.
package pluginmarket
