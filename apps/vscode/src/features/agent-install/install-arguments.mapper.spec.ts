import { describe, expect, it } from 'vitest'
import { installArguments, optOutArguments } from './install-arguments.mapper.ts'

describe('installArguments', () => {
  it('previews the agent for both clients, then writes it with --yes', () => {
    expect(installArguments('agent', false)).toEqual(['agent', 'install', '--claude', '--vscode'])
    expect(installArguments('agent', true)).toEqual(['agent', 'install', '--claude', '--vscode', '--yes'])
  })

  it('installs the MCP server alone for VS Code', () => {
    expect(installArguments('mcp', false)).toEqual(['mcp', 'install', '--vscode'])
    expect(installArguments('mcp', true)).toEqual(['mcp', 'install', '--vscode', '--yes'])
  })
})

describe('optOutArguments', () => {
  it('records the opt-out through smith config, not by editing smith.toml', () => {
    expect(optOutArguments()).toEqual(['config', 'set', 'agent.enabled', 'false'])
  })
})
