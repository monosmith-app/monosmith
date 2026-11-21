package cipipeline

import (
	"bufio"
	"io"
	"path"

	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/apps/smith/features/trustedpublishing"
)

// SignInAsk says whether OfferSignIns may ask, and where.
type SignInAsk struct {
	// Ask is true when a person answers in the terminal.
	Ask bool
	// In is the terminal input; nil when Ask is false.
	In *bufio.Reader
	// Out takes the questions, or the notes of a run that does not ask.
	Out io.Writer
}

// OfferSignIns offers trusted publishing for each registry the pipelines
// publish to with a long-lived secret that smith.toml does not decide yet
// (trustedpublishing.Offer), and drops the secrets the answers replace from
// c, so the pipelines are generated with them.
//
// Parameters: s is the open session, whose configuration takes the answers;
// c are the collected contributions; a says whether to ask.
// Returns: an error when smith.toml cannot be written.
func OfferSignIns(s *session.Session, c *Contributions, a SignInAsk) error {
	secrets := make([]string, len(c.Credentials))
	for i, cred := range c.Credentials {
		secrets[i] = cred.Env
	}
	if len(trustedpublishing.Pending(s.Config, secrets)) == 0 {
		return nil
	}
	place := trustedpublishing.Place{Workflow: path.Base(GitHubPath), Branches: []string{s.Config.Release.Branch}}
	for _, ch := range s.Config.Release.Channels {
		place.Branches = append(place.Branches, ch.Branch)
	}
	if repo, err := release.Repository(s); err == nil {
		place.Owner, place.Repository = repo.Owner, repo.Name
	}
	if _, err := trustedpublishing.Offer(trustedpublishing.Question{Root: s.Root, Config: &s.Config, Secrets: secrets, Place: place, Ask: a.Ask, In: a.In, Out: a.Out}); err != nil {
		return err
	}
	c.ApplySignIns(s.Config.Release)
	return nil
}
