import { describe, expect, it } from 'vitest'
import { DEFAULT_CHOICES, describeChoices, toCreateArguments } from './new-workspace-arguments.mapper.ts'
import { parseNewQuestions } from './new-questions.mapper.ts'

describe('toCreateArguments', () => {
  it('gives every question a flag, so smith asks nothing and decides nothing in silence', () => {
    expect(toCreateArguments({ name: 'ws-j', ...DEFAULT_CHOICES })).toEqual([
      'new', 'ws-j', '--here', '--yes', '--ci', 'github', '--language', 'later', '--architecture', 'vertical-slices', '--agent', '--devcontainer',
    ])
  })

  it('passes the Node stack only for a Node workspace, and several CI systems as a list', () => {
    const node = toCreateArguments({ name: 'a', ...DEFAULT_CHOICES, language: 'node', nodeStack: 'node-vitest', ci: ['github', 'azure'] })
    expect(node).toEqual(expect.arrayContaining(['--ci', 'github,azure', '--default-stack', 'node=node-vitest']))
    expect(toCreateArguments({ name: 'a', ...DEFAULT_CHOICES, language: 'go', nodeStack: 'node-vitest' })).not.toContain('--default-stack')
  })

  it('writes the opposite flags for a no, and no commit only when the person said no', () => {
    const args = toCreateArguments({ name: 'a', ...DEFAULT_CHOICES, agent: false, devcontainer: false, commit: false })
    expect(args).toEqual(expect.arrayContaining(['--no-agent', '--no-devcontainer', '--no-commit']))
    expect(toCreateArguments({ name: 'a', ...DEFAULT_CHOICES })).not.toContain('--no-commit')
  })

  it('adds --dry-run for the preview that lists the files and writes nothing', () => {
    expect(toCreateArguments({ name: 'a', ...DEFAULT_CHOICES }, true).at(-1)).toBe('--dry-run')
  })
})

const QUESTIONS = parseNewQuestions(JSON.stringify([
  { key: 'ci', flag: 'ci', options: [{ value: 'github', label: 'GitHub Actions' }] },
  { key: 'language', flag: 'language', options: [{ value: 'later', label: 'Decide later' }, { value: 'node', label: 'TypeScript / JavaScript' }] },
  { key: 'node-stack', flag: 'default-stack', options: [{ value: 'node-vitest', label: 'Vitest' }] },
  { key: 'architecture', flag: 'architecture', options: [{ value: 'vertical-slices', label: 'Vertical slices' }] },
]))

describe('describeChoices', () => {
  it('says the name comes from the folder, and what the other choices mean', () => {
    const lines = describeChoices({ name: 'ws-j', ...DEFAULT_CHOICES }, 'ws-j', QUESTIONS)
    expect(lines[0]).toBe('Name: ws-j (the name of this folder)')
    expect(lines).toContain('CI: GitHub Actions')
    expect(lines).toContain('Main language: Decide later')
    expect(lines).toContain('First commit of smith\'s own files: yes, smith makes it')
  })

  it('does not say the name is the folder\'s when the person typed another, and names the Node stack', () => {
    const lines = describeChoices({ name: 'acme', ...DEFAULT_CHOICES, language: 'node', nodeStack: 'node-vitest', commit: false }, 'ws-j', QUESTIONS)
    expect(lines[0]).toBe('Name: acme')
    expect(lines).toContain('Tests and lint for Node projects: Vitest')
    expect(lines.at(-1)).toContain('stay uncommitted')
  })
})
