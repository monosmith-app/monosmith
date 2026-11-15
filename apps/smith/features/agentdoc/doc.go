// Package agentdoc is `smith agent file`: it writes AGENTS.md, and a CLAUDE.md
// that points to it, from the rules of the workspace: the architecture in
// force, the naming of files, how to add a project, the commands that mean
// done, the derived files not to edit by hand, and the choices of each
// project. The generated text sits between region markers, so the text a
// person writes in the same file survives a new run. The text is the same for
// the same workspace. See ADR 0048.
package agentdoc
