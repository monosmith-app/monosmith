// Package pubrelease is `smith release pub`, which the generated pub
// workflow runs on a release's pub tag: it fetches the release's archives
// from its GitHub Release, builds the pub.dev package from them and
// publishes it, through the OIDC token pub.dev's automated publishing
// accepts.
package pubrelease
