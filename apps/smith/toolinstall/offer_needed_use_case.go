package toolinstall

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// OfferOptions say what a command that made a project offers about its tools.
type OfferOptions struct {
	// Out is where the findings and the questions are printed.
	Out io.Writer
	// In is where the answers are read from.
	In io.Reader
	// Ask is true when a person is at a terminal and did not pass --yes: smith
	// then presents the install choice of each tool. When false smith installs
	// nothing and prints the one command that does.
	Ask bool
	// Projects are the projects the command made or touched; nil means all.
	Projects []string
	// ExtraLanguages and ExtraChoices add the tools of a language or choice no
	// project uses yet.
	ExtraLanguages, ExtraChoices []string
}

// OfferText words the tools that are missing or different as one block: a
// line for each with what stops working, then the single command that offers
// the ways to get all of them. It is empty when nothing is lacking.
func OfferText(list []Status) string {
	lines, names := lackingLines(list)
	if len(names) == 0 {
		return ""
	}
	return lines + fmt.Sprintf("→ Run smith tools install %s to choose how to get %s.\n", strings.Join(names, " "), pronoun(len(names)))
}

// lackingLines is a "!" line for each tool that is lacking, and their names.
func lackingLines(list []Status) (string, []string) {
	var b strings.Builder
	var names []string
	for _, st := range list {
		if st.Lacking() {
			names = append(names, st.Name)
			fmt.Fprintf(&b, "! %s %s: %s.\n", st.Title, st.Problem(), st.Consequence)
		}
	}
	return b.String(), names
}

func pronoun(n int) string {
	if n == 1 {
		return "it"
	}
	return "them"
}

// Offer checks the tools the touched projects need and deals with what is
// missing or different, so the first `smith doctor` does not have to. With
// o.Ask it presents the install choice of each tool (the choices of
// `smith tools install`: a package manager, the official page, "I will install
// it myself", skip or keep mine); nothing is installed before the person
// chooses and a command that needs administrator rights is never run in
// silence. Without o.Ask it installs nothing and prints OfferText.
//
// Parameters: ctx bounds the detection and the installs; s is the open session;
// o says where to print and read, and which projects to check.
// Returns: the tools that were lacking before the offer, in the order the
// plugins declare them, with the state each ended in.
func Offer(ctx context.Context, s *session.Session, o OfferOptions) []Status {
	needs := OnlyProjects(WorkspaceNeeds(s, o.ExtraLanguages, o.ExtraChoices), o.Projects)
	return offerFor(ctx, DetectAll(ctx, ExecDetector{}, needs), o, ExecRunner{}, ExecDetector{})
}

// offerFor is Offer for tools already detected; tests give a fake runner and detector.
func offerFor(ctx context.Context, list []Status, o OfferOptions, runner Runner, detector Detector) []Status {
	var lacking []Status
	for _, st := range list {
		if st.Lacking() {
			lacking = append(lacking, st)
		}
	}
	if len(lacking) == 0 {
		return nil
	}
	if !o.Ask {
		fmt.Fprint(o.Out, OfferText(lacking))
		return lacking
	}
	lines, _ := lackingLines(lacking)
	fmt.Fprint(o.Out, lines)
	fmt.Fprintln(o.Out, "Choose how to get each one now. Nothing is installed before you choose.")
	f := Flow{In: bufio.NewReader(o.In), Out: o.Out, Interactive: true, Runner: runner, Detector: detector}
	outcomes, err := f.Run(ctx, PlansFor(lacking, false))
	if err != nil {
		fmt.Fprintf(o.Out, "→ smith could not ask: %v. Run smith tools install to choose.\n", err)
		return lacking
	}
	for i, out := range outcomes {
		lacking[i].State = out.State
	}
	return lacking
}
