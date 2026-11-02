package conformance

import (
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// checkLink checks what every plugin with the link capability does the same: a
// project whose folder is not there is refused with an error, never given a
// manifest.
func checkLink(c *host.Client, r *Report) {
	if !c.Info.Capabilities.Link {
		return
	}
	language := ""
	if len(c.Info.Languages) > 0 {
		language = c.Info.Languages[0]
	}
	gone := protocol.Project{Name: "conformance-gone", Root: "conformance/no-such-folder", Language: language, Kind: protocol.ProjectLib}
	lib := protocol.Project{Name: "conformance-lib", Root: "conformance/no-such-lib", Language: language, Kind: protocol.ProjectLib, Exports: []string{"conformance-lib"}}
	var res protocol.LinkResult
	err := c.Call(protocol.MethodLink, protocol.LinkParams{Project: gone, Library: lib, Quiet: true}, &res, time.Minute)
	r.add("a link in a folder that does not exist is refused", err != nil, "it must answer with an error that names the manifest, and write nothing")
}
