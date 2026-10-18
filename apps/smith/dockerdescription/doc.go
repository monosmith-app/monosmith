// Package dockerdescription is `smith docker description`: it sets the Docker
// Hub description of the apps whose registries name a description_from file,
// without a release. The sync itself is the docker plugin's, shared with the
// release; this slice finds the apps and reports.
package dockerdescription
