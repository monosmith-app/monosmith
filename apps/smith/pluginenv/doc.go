// Package pluginenv decides which environment variables an external plugin
// is started with: a small documented baseline, plus the variables the plugin
// (or the workspace owner) names. Built-in plugins are the smith binary itself
// and do not go through here.
package pluginenv
