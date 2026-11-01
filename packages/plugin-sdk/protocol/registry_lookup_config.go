package protocol

import (
	"context"
	"errors"
	"net"
	"time"
)

// RegistryLookupTimeout is how long a plugin waits for a registry to answer one
// lookup (the published or latest version of a package). When it runs out, the
// plugin answers the lookup with Unreachable and keeps running: the host never
// has to kill the plugin for a registry that hangs.
const RegistryLookupTimeout = 8 * time.Second

// RegistryUnreachableDetail is the reason a lookup that ran out of time puts in
// PublishedResult.Detail.
const RegistryUnreachableDetail = "skipped: registry unreachable (no answer within 8s)"

// LookupFailure is the Detail for a failed registry lookup: the fixed
// RegistryUnreachableDetail when the registry did not answer in time, else the
// error's own text.
func LookupFailure(err error) string {
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return RegistryUnreachableDetail
	}
	return err.Error()
}
