// Package fixwizard is the dialog that walks a doctor fix: the questions the
// fix asks (numbered options), a review of what it will change (files as
// diffs, the commands it runs, the projects it covers), the progress while it
// applies, and, when a step fails, that step's output with Retry and Close.
//
// It holds the state of the walk and draws it; it applies nothing. The shell
// starts doctorfix.Apply when the wizard asks for it (Outcome.Apply) and
// reports back with Progress and Fail. Nothing here needs a terminal: the
// plain, non-TTY and screen-reader modes never open it, `smith doctor --fix`
// asks the same questions as numbered prompts.
package fixwizard
