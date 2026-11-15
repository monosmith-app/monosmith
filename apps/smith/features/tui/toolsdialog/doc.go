// Package toolsdialog is the "Install the missing tools" dialog of the
// interactive shell (ADR 0077). It asks one question per missing tool, with the
// exact command of each option and the marked default, then a second question
// when the command needs administrator rights, shows the output of the install
// as it streams, and ends with a result per tool. The model runs nothing: it
// asks its host to run a smith command through Outcome, so the choice always
// comes before the install.
package toolsdialog
