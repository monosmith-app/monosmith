package convert

import (
	"context"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/toolinstall"
)

// OfferConvertedTools checks the tools the converted projects need and deals
// with what is missing or different, as `smith add` does for a new project: a
// choice in a terminal, the command to run otherwise. A workspace that cannot
// be opened offers nothing here: the conversion already checked that it loads.
//
// Parameters: ctx bounds the detection and the installs; root is the
// workspace; version is the running smith's version; out is where it prints;
// in is where it reads the answers; ask is true when a person can be asked.
func OfferConvertedTools(ctx context.Context, root, version string, out io.Writer, in io.Reader, ask bool) {
	s, err := session.Open(session.Options{Dir: root, Version: version, Log: io.Discard, AllowCycles: true})
	if err != nil {
		return
	}
	defer s.Close()
	toolinstall.Offer(ctx, s, toolinstall.OfferOptions{Out: out, In: in, Ask: ask})
}
