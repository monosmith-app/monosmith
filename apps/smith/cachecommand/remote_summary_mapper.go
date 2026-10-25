package cachecommand

import (
	"net/url"

	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// RemoteSummary describes the remote cache without anything secret: the
// scheme and host, never a path, a query or credentials in the address
// (url.URL.Host leaves out any user:password@).
type RemoteSummary struct {
	Configured bool   `json:"configured"`
	Scheme     string `json:"scheme,omitempty"`
	Host       string `json:"host,omitempty"`
	// Active is true when the remote is used: it needs SMITH_CACHE_KEY.
	Active   bool `json:"active"`
	ReadOnly bool `json:"readOnly"`
}

// summarizeRemote reduces the remote cache settings to what is safe to print.
func summarizeRemote(plan session.RemotePlan) RemoteSummary {
	if plan.URL == "" {
		return RemoteSummary{}
	}
	out := RemoteSummary{Configured: true, Active: plan.Active, ReadOnly: plan.ReadOnly}
	if u, err := url.Parse(plan.URL); err == nil {
		out.Scheme, out.Host = u.Scheme, u.Host
	}
	return out
}
