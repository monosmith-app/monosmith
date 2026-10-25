// Package nxworkspace reads an Nx workspace - nx.json, project.json files and
// the "nx" fields of package.json files - into a model a conversion can map.
// It asks the workspace's own Nx for the resolved project graph
// (`nx graph --file`), so the dependencies of an Nx workspace must be installed.
package nxworkspace
