import { describe, expect, it } from 'vitest'
import { nameTask, toSmithArguments } from './smith-arguments.mapper.ts'

describe('toSmithArguments', () => {
  it('runs a task in every project that has it', () => {
    expect(toSmithArguments({ task: 'build' })).toEqual(['run', 'build'])
  })

  it('adds the project, the flags and the parallelism', () => {
    expect(toSmithArguments({ task: 'test', project: 'kind:app', affected: true, bail: true, noCache: true, parallel: 4 }))
      .toEqual(['run', 'test', '-p', 'kind:app', '--affected', '--bail', '--no-cache', '--parallel', '4'])
  })

  it('leaves out an empty project and a parallelism of zero', () => {
    expect(toSmithArguments({ task: 'lint', project: '', parallel: 0 })).toEqual(['run', 'lint'])
  })

  it('refuses a value that could be an option', () => {
    expect(toSmithArguments({ task: '--json' })).toBeUndefined()
    expect(toSmithArguments({ task: 'build', project: '--all' })).toBeUndefined()
    expect(toSmithArguments({ task: 'two words' })).toBeUndefined()
    expect(toSmithArguments({ task: '' })).toBeUndefined()
  })
})

describe('nameTask', () => {
  it('names a task by what it runs and where', () => {
    expect(nameTask({ task: 'build' })).toBe('build')
    expect(nameTask({ task: 'build', project: 'web' })).toBe('web:build')
    expect(nameTask({ task: 'test', affected: true })).toBe('test (affected)')
  })
})
