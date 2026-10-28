/**
 * The MCP registration of smith as `.mcp.json` words it (Claude Code, and most
 * clients), for a person who adds it to a client smith does not write for.
 */
export function mcpConfigText (): string {
  return `${JSON.stringify({ mcpServers: { smith: { command: 'smith', args: ['mcp'] } } }, undefined, 2)}\n`
}
