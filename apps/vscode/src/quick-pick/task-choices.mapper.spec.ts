import { describe, expect, it } from 'vitest'
import type { WorkspaceProject, WorkspaceSnapshot } from '../workspace/index.ts'
import { folderChoices, projectChoices, taskChoices, taskNameChoices } from './task-choices.mapper.ts'

const project = (name: string, tasks: string[]): WorkspaceProject => ({ name, dir: name, kind: 'app', language: 'ts', tasks, deps: [] })
const snapshot: WorkspaceSnapshot = {
  projects: [project('web', ['build', 'lint']), project('api', ['build', 'test']), project('docs', [])],
  affected: new Set(['web']),
}

describe('projectChoices', () => {
  it('puts the affected projects first, then by name, and leaves out those without tasks', () => {
    expect(projectChoices(snapshot).map((choice) => choice.value.name)).toEqual(['web', 'api'])
  })

  it('marks an affected project and lists its tasks', () => {
    const [first] = projectChoices(snapshot)
    expect(first?.label).toContain('web')
    expect(first?.description).toContain('affected')
    expect(first?.detail).toBe('build, lint')
  })
})

describe('taskChoices', () => {
  it('offers all tasks first, as undefined, then each task', () => {
    const choices = taskChoices(project('api', ['build', 'test']))
    expect(choices.map((choice) => choice.value)).toEqual([undefined, 'build', 'test'])
  })
})

describe('taskNameChoices', () => {
  it('lists each task name once with how many projects have it', () => {
    const choices = taskNameChoices(snapshot, false)
    expect(choices.map((choice) => [choice.value, choice.description])).toEqual([
      ['build', '2 projects'],
      ['lint', '1 project'],
      ['test', '1 project'],
    ])
  })

  it('counts only affected projects when asked', () => {
    expect(taskNameChoices(snapshot, true).map((choice) => choice.value)).toEqual(['build', 'lint'])
  })
})

describe('folderChoices', () => {
  it('names a folder and counts its projects', () => {
    expect(folderChoices([{ root: '/a', name: 'a', snapshot }])).toEqual([{ label: 'a', description: '3 projects', value: '/a' }])
  })
})
