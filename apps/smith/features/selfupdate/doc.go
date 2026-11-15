// Package selfupdate is smith update: it asks the registry for the newest smith
// on the person's channel, shows what would run (the command that matches how
// smith was installed) and runs it after the person chooses. It never updates
// without a choice: --yes is the choice for a script.
package selfupdate
