import { describe, expect, it } from 'vitest'
import type { WorkspaceProject } from '../workspace/index.ts'
import { describePlan, toRunPlan } from './run-plan.mapper.ts'

const project = (name: string, tasks: string[]): WorkspaceProject => ({ name, dir: name, kind: 'app', language: 'ts', tasks, deps: [] })
const projects = [project('api', ['build', 'test']), project('web', ['build', 'lint']), project('docs', [])]
const off = { bail: false, noCache: false }

describe('toRunPlan', () => {
  it('runs one task of one project', () => {
    expect(toRunPlan({ kind: 'task', project: 'api', task: 'build' }, projects, off)).toEqual({ tasks: ['build'], projects: ['api'] })
  })

  it('runs every task of a project', () => {
    expect(toRunPlan({ kind: 'projects', projects: ['api'] }, projects, off)).toEqual({ tasks: ['build', 'test'], projects: ['api'] })
  })

  it('runs the tasks of a group once each, in the projects that have tasks', () => {
    expect(toRunPlan({ kind: 'projects', projects: ['api', 'web', 'docs'] }, projects, off))
      .toEqual({ tasks: ['build', 'test', 'lint'], projects: ['api', 'web'] })
  })

  it('has nothing to run for a project without tasks', () => {
    expect(toRunPlan({ kind: 'projects', projects: ['docs'] }, projects, off)).toBeUndefined()
  })

  it('runs a task in the affected projects', () => {
    expect(toRunPlan({ kind: 'affected', task: 'test' }, projects, off)).toEqual({ tasks: ['test'], affected: true })
  })

  it('runs a task everywhere', () => {
    expect(toRunPlan({ kind: 'all', task: 'build' }, projects, off)).toEqual({ tasks: ['build'] })
  })

  it('carries bail and no-cache', () => {
    expect(toRunPlan({ kind: 'all', task: 'build' }, projects, { bail: true, noCache: true })).toEqual({ tasks: ['build'], bail: true, noCache: true })
  })
})

describe('describePlan', () => {
  it('names a single task like its id', () => {
    expect(describePlan({ tasks: ['build'], projects: ['api'] })).toBe('api:build')
  })

  it('says where an affected run goes', () => {
    expect(describePlan({ tasks: ['test'], affected: true })).toBe('test in affected projects')
  })

  it('counts the tasks of a larger run', () => {
    expect(describePlan({ tasks: ['build', 'test'], projects: ['api'] })).toBe('2 tasks in api')
  })

  it('does not list a long set of projects', () => {
    const many = Array.from({ length: 30 }, (_, index) => `project-${index}`)
    expect(describePlan({ tasks: ['build'], projects: many })).toBe('build in 30 projects')
  })
})
