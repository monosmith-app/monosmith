import { describe, expect, it } from 'vitest'
import { mcpConfigText } from './mcp-config.mapper.ts'

describe('mcpConfigText', () => {
  it('registers smith mcp the way .mcp.json does', () => {
    expect(JSON.parse(mcpConfigText())).toEqual({ mcpServers: { smith: { command: 'smith', args: ['mcp'] } } })
  })
})
