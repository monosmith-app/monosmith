import { describe, expect, it } from 'vitest'
import { toExplainDocument } from './explain-document.mapper.ts'

describe('toExplainDocument', () => {
  const document = toExplainDocument('web', [
    {
      id:          'web:build',
      description: 'Compile into dist',
      fields:      [
        { name: 'command', value: 'npm run build --silent', source: 'plugin node' },
        { name: 'inputs', value: 'a | b', source: 'plugin node' },
        { name: 'env', value: '—', source: 'plugin node' },
      ],
    },
    { id: 'web:start', fields: [] },
  ])

  it('has a section per task with its description', () => {
    expect(document).toContain('# Explain web')
    expect(document).toContain('## web:build\n\nCompile into dist')
    expect(document).toContain('## web:start')
  })

  it('tabulates the fields with their source and keeps a pipe from breaking the table', () => {
    expect(document).toContain('| command | `npm run build --silent` | plugin node |')
    expect(document).toContain('| inputs | `a \\| b` | plugin node |')
    expect(document).toContain('| env | — | plugin node |')
  })

  it('says so when the project has no tasks', () => {
    expect(toExplainDocument('x', [])).toContain('This project has no tasks.')
  })
})
