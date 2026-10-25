import { describe, expect, it } from 'vitest'
import { task, buildTask } from './task.algorithm.ts'

describe('task', () => {
  it('is a command with options', () => {
    expect(task(['git', '--version'], { cache: true, inputs: ['widget.json'] })).toEqual({ command: ['git', '--version'], cache: true, inputs: ['widget.json'] })
  })
})

describe('buildTask', () => {
  it('builds into dist/, caches, restores the outputs and waits for the dependencies', () => {
    expect(buildTask('libs/core', outDir => ['tool', 'build', outDir], { inputs: ['widget.json'] })).toEqual({
      command:   ['tool', 'build', '../../dist/libs/core'],
      outputs:   ['../../dist/libs/core'],
      dependsOn: ['^build'],
      cache:     true,
      inputs:    ['widget.json'],
    })
  })

  it('lets an option replace a default', () => {
    expect(buildTask('web', () => ['x'], { dependsOn: [], cache: false })).toMatchObject({ dependsOn: [], cache: false })
  })
})
