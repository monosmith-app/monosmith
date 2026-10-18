// Package dockerdescription is `smith docker description`: it sets the Docker
// Hub descriptions (full and short) of the apps that have a docker.io
// registry, without a release. The sync itself is the docker plugin's, shared with the
// release; this slice finds the apps and reports.
package dockerdescription
