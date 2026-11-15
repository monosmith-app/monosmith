import { describe, expect, it } from 'vitest'
import { toOfferedTasks } from './offered-tasks.mapper.ts'

describe('toOfferedTasks', () => {
  const offered = toOfferedTasks(['test', 'build', 'test', 'lint'])

  it('offers each task name once, with its affected twin', () => {
    expect(offered.map((task) => task.name)).toEqual(['build', 'build (affected)', 'lint', 'lint (affected)', 'test', 'test (affected)'])
    expect(offered[1]?.definition).toEqual({ type: 'smith', task: 'build', affected: true })
  })

  it('puts build and test in their groups, and only the plain ones', () => {
    expect(offered.map((task) => task.group)).toEqual(['build', undefined, undefined, undefined, 'test', undefined])
  })

  it('offers nothing for a folder with no tasks', () => {
    expect(toOfferedTasks([])).toEqual([])
  })
})
