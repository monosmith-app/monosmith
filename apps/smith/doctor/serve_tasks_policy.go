package doctor

import (
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
)

// BadServe lists the tasks whose serve settings cannot work: serve names a
// task that does not exist, serve_url or serve_timeout without serve, a
// serve_url that is not http(s), a serve_timeout that is not a duration.
func BadServe(g *projectgraph.Graph) []string {
	var out []string
	for _, p := range g.Projects {
		names := make([]string, 0, len(p.Tasks))
		for n := range p.Tasks {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			t := p.Tasks[n]
			id := p.Name + ":" + n
			if t.Serve == "" {
				if t.ServeURL != "" {
					out = append(out, id+" sets serve_url without serve")
				}
				if t.ServeTimeout != "" {
					out = append(out, id+" sets serve_timeout without serve")
				}
				continue
			}
			if _, _, err := taskgraph.ServeTarget(g, p.Name, t.Serve); err != nil {
				out = append(out, fmt.Sprintf("%s: %v", id, err))
			}
			if t.ServeURL != "" {
				if u, err := url.Parse(t.ServeURL); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
					out = append(out, fmt.Sprintf("%s: serve_url %q is not an http(s) URL", id, t.ServeURL))
				}
			}
			if t.ServeTimeout != "" {
				if d, err := time.ParseDuration(t.ServeTimeout); err != nil || d <= 0 {
					out = append(out, fmt.Sprintf("%s: serve_timeout %q is not a duration like 60s", id, t.ServeTimeout))
				}
			}
		}
	}
	return out
}
