// Package mcpserver is `smith mcp`: the workspace served to an AI assistant
// over the Model Context Protocol (JSON-RPC on stdio). Each tool is a smith
// command line: the server runs smith itself with it, so a tool does exactly
// what the command does, and a client can show the user that command.
package mcpserver
