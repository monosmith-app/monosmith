package cachecommand

import "github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"

// StatusReport is `smith cache status --json`.
type StatusReport struct {
	cachemaintenance.Status
	// UsedPercent is how full the cache is against its limit; absent when
	// there is no limit.
	UsedPercent *float64      `json:"usedPercent,omitempty"`
	Remote      RemoteSummary `json:"remote"`
}
