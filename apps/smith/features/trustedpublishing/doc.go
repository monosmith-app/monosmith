// Package trustedpublishing offers the registry sign-ins that need no
// long-lived key (ADR 0109). When a GitHub workspace publishes to NuGet,
// RubyGems or the Visual Studio Marketplace and smith.toml does not yet say
// how, smith ci generate asks: trusted publishing (the workflow signs in with
// its OIDC token) or the API key secret. The answer goes to smith.toml, a "no"
// as false, so smith does not ask again. A run without a person (--yes, CI)
// does not ask: it prints one note that names the setting.
package trustedpublishing
