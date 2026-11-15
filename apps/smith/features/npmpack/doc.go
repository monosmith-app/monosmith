// Package npmpack is `smith npm-pack`, a step of the node plugin's package
// task: an app that builds into dist/<project dir> is packed from there, once
// its package.json is staged beside the build. Hidden: tasks call it through
// {smith}, people call smith ci pack.
package npmpack
