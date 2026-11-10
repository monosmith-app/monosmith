// Package prompt answers `smith prompt`: the one-line status a shell prompt
// shows (Oh My Posh, Starship, powerlevel10k). It must answer in about 50 ms,
// so it reads only what smith doctor saved in .smith/prompt.json and asks git
// which files changed; it never opens a session, starts a plugin or runs doctor.
package prompt
