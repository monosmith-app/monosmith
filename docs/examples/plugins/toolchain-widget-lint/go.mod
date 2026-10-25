module example.com/smith-plugin-widget-lint

go 1.26.0

require github.com/monosmith-app/monosmith/packages/plugin-sdk v0.0.0

// Inside the MonoSmith repository the SDK is a neighbour. In your own plugin,
// drop this line and `go get` a released version of the SDK instead.
replace github.com/monosmith-app/monosmith/packages/plugin-sdk => ../../../../packages/plugin-sdk
