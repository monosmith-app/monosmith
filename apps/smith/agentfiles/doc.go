// Package agentfiles is the MonoSmith agent a workspace carries: the Claude
// Code subagent .claude/agents/smith.md and the registration of smith mcp in
// .mcp.json (Claude Code) and .vscode/mcp.json (VS Code). smith writes them in
// smith new, offers them in smith adopt and smith convert, keeps the agent file
// in step in smith upgrade, checks them in smith doctor, and installs them on
// demand with smith agent install and smith mcp install.
//
// The agent file carries a marker with a hash of its text. A file whose hash
// matches is smith's and is rewritten when smith's text changes; any other file
// is the person's and is never overwritten. The MCP files are JSON and have no
// comments, so smith owns only the entry named smith in them.
package agentfiles
