package doctorfix

import (
	"fmt"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Resolve answers a fix's questions in order. answers are the picked options,
// counted from 1 (the first option, the suggested one, is 1). It returns the
// fix with the picked options' files and steps added after its own, and
// no questions left; when answers run out first it returns the fix as far as
// it got and the next question to ask, so a wizard calls it again after each
// answer. A picked option may itself ask: its questions come after the ones
// already pending.
func Resolve(fix protocol.Fix, answers []int) (protocol.Fix, *protocol.FixChoice, error) {
	pending := append([]protocol.FixChoice(nil), fix.Choices...)
	resolved := fix
	resolved.Choices = nil
	resolved.Files = append([]protocol.FileEdit(nil), fix.Files...)
	resolved.Steps = append([]protocol.FixStep(nil), fix.Steps...)
	for n := 0; len(pending) > 0; n++ {
		q := pending[0]
		if n >= len(answers) {
			resolved.Choices = pending
			return resolved, &q, nil
		}
		pick := answers[n]
		if pick < 1 || pick > len(q.Options) {
			return resolved, nil, fmt.Errorf("%q has options 1 to %d, not %d", q.Question, len(q.Options), pick)
		}
		option := q.Options[pick-1].Fix
		resolved.Files = append(resolved.Files, option.Files...)
		resolved.Steps = append(resolved.Steps, option.Steps...)
		pending = append(pending[1:], option.Choices...)
	}
	return resolved, nil, nil
}
