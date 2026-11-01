import { describe, expect, it } from 'vitest'
import { readFileWhy, readSettingWhy, readTaskWhy } from './read-why.use-case.ts'

describe('readSettingWhy', () => {
  it('asks for JSON and puts the setting after --', async () => {
    const calls: string[][] = []
    const page = await readSettingWhy(async (args) => {
      calls.push(args)

      return { code: 0, stdout: JSON.stringify({ path: 'ci.devcontainer', source: 'default', default: 'false', explanation: 'x', change: 'smith config set ci.devcontainer <value>' }), stderr: '' }
    }, 'ci.devcontainer')
    expect(calls).toEqual([['why', '--json', '--', 'ci.devcontainer']])
    expect(page).toContain('# Why: ci.devcontainer')
  })
})

describe('readFileWhy', () => {
  it('asks for JSON and puts the file after --', async () => {
    const calls: string[][] = []
    const page = await readFileWhy(async (args) => {
      calls.push(args)

      return { code: 0, stdout: JSON.stringify({ path: 'AGENTS.md', known: true, summary: 'smith owns the marked lines.', change: ['run smith agent file'] }), stderr: '' }
    }, 'AGENTS.md')
    expect(calls).toEqual([['why', '--json', '--', 'AGENTS.md']])
    expect(page).toContain('# Why: AGENTS.md')
  })

  it('throws smith words when it fails', async () => {
    await expect(readFileWhy(async () => ({ code: 1, stdout: '', stderr: '✖ not a workspace\n  → run smith new\n' }), 'a')).rejects.toThrow('not a workspace Try: run smith new')
  })

  it('throws a fix when the answer is not JSON', async () => {
    await expect(readFileWhy(async () => ({ code: 0, stdout: 'text', stderr: '' }), 'a')).rejects.toThrow('Update the extension')
  })
})

describe('readTaskWhy', () => {
  it('asks why-miss for project:task', async () => {
    const calls: string[][] = []
    await readTaskWhy(async (args) => {
      calls.push(args)

      return { code: 0, stdout: JSON.stringify({ task: 'api:build', status: 'hit', key: 'k' }), stderr: '' }
    }, 'api', 'build')
    expect(calls).toEqual([['why-miss', 'api:build', '--json']])
  })
})
