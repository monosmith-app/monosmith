import { describe, expect, it } from 'vitest'
import { describeFailure, sayProblem } from './say-problem.mapper.ts'

describe('sayProblem', () => {
  it('drops the marks and keeps each way out', () => {
    expect(sayProblem('✖ it broke\n  → do this\n  → or that\n')).toBe('it broke Try: do this Try: or that')
  })
})

describe('describeFailure', () => {
  it('uses standard error first', () => {
    expect(describeFailure({ code: 1, stdout: 'x', stderr: '✖ no smith.toml\n  → run smith new\n' })).toBe('no smith.toml Try: run smith new')
  })

  it('falls back to standard output, then to the exit code', () => {
    expect(describeFailure({ code: 1, stdout: '✖ from stdout\n', stderr: ' \n' })).toBe('from stdout')
    expect(describeFailure({ code: 3, stdout: '', stderr: '' })).toBe('smith stopped with exit code 3')
  })
})
