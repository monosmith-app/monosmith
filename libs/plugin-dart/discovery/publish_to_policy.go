package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// PubDev is what a package that publishes to pub.dev reports as PublishTo.
const PubDev = "pub.dev"

// publishTo is where a package publishes: pub.dev unless its pubspec's
// publish_to says otherwise. Only a package publishes; anything else says
// nothing.
func publishTo(p Pubspec, kind protocol.ProjectKind) string {
	if kind != protocol.ProjectPackage {
		return ""
	}
	if p.PublishTo == "" {
		return PubDev
	}
	return p.PublishTo
}
