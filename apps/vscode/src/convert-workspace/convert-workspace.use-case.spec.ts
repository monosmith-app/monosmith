import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { toConvertArguments } from './convert-arguments.mapper.ts'
import { convertWorkspace, type ConvertDeps } from './convert-workspace.use-case.ts'

const OK = (stdout: string): SmithRun => ({ code: 0, stdout, stderr: '' })

const QUESTIONS = JSON.stringify([
  { key: 'level', prompt: 'How far?', default: 'conservative', flag: 'level', options: [{ value: 'conservative', label: 'conservative' }, { value: 'full', label: 'full' }] },
  { key: 'lint', prompt: 'Lint with:', default: 'keep', flag: 'lint', when: { level: 'full' }, options: [{ value: 'keep', label: 'keep' }, { value: 'smith', label: 'smith' }] },
  { key: 'test', prompt: 'Run tests with:', default: 'keep', flag: 'test', when: { level: 'full' }, options: [{ value: 'keep', label: 'keep' }, { value: 'jest', label: 'jest' }] },
  { key: 'build', prompt: 'Build with:', default: 'smith', flag: 'build', when: { level: 'full' }, options: [{ value: 'smith', label: 'smith' }, { value: 'keep', label: 'keep' }] },
  { key: 'ci', prompt: 'CI pipeline:', default: 'smith', flag: 'ci', when: { level: 'full' }, options: [{ value: 'smith', label: 'smith' }, { value: 'keep', label: 'keep' }] },
])

function fake (answers: Map<string, string | undefined> = new Map(), over: Partial<ConvertDeps> = {}) {
  const ran: string[][] = []
  const applied: string[][] = []
  const asked: string[] = []
  const deps: ConvertDeps = {
    run: async (args) => {
      ran.push(args)

      return OK(args.includes('--questions') ? QUESTIONS : 'would delete jest.config.js\n')
    },
    apply: async (args) => {
      applied.push(args)

      return OK('converted\n')
    },
    pick: async (question, current) => {
      asked.push(question.key)
      const answer = answers.has(question.key) ? answers.get(question.key) : current[0]

      return answer === undefined ? undefined : [answer]
    },
    askCommit: async () => false,
    confirm:   async () => true,
    ...over,
  }

  return { deps, ran, applied, asked }
}

describe('convertWorkspace', () => {
  it('asks only the level when it is conservative, because the four choices apply to the full level', async () => {
    const f = fake()
    expect(await convertWorkspace(f.deps)).toEqual({ kind: 'converted', lines: ['converted'], committed: false })
    expect(f.asked).toEqual(['level'])
    expect(f.ran[1]).toEqual(['convert', '--level', 'conservative', '--dry-run'])
    expect(f.applied[0]).toEqual(['convert', '--level', 'conservative', '--yes'])
  })

  it('asks the four choices at the full level and passes each as a flag', async () => {
    const f = fake(new Map([['level', 'full'], ['lint', 'smith'], ['test', 'jest']]))
    await convertWorkspace(f.deps)
    expect(f.asked).toEqual(['level', 'lint', 'test', 'build', 'ci'])
    expect(f.applied[0]).toEqual(['convert', '--level', 'full', '--lint', 'smith', '--test', 'jest', '--build', 'smith', '--ci', 'smith', '--yes'])
  })

  it('shows the plan before anything is written and writes nothing when the person says no', async () => {
    const f = fake(new Map(), { confirm: async () => false })
    expect(await convertWorkspace(f.deps)).toEqual({ kind: 'cancelled' })
    expect(f.applied).toEqual([])
  })

  it('commits only when the box is ticked', async () => {
    const f = fake(new Map(), { askCommit: async () => true })
    expect(await convertWorkspace(f.deps)).toMatchObject({ kind: 'converted', committed: true })
    expect(f.applied[0]?.at(-1)).toBe('--commit')
  })

  it('writes nothing when the person cancels a question or the commit box', async () => {
    for (const f of [fake(new Map([['level', undefined]])), fake(new Map(), { askCommit: async () => undefined })]) {
      expect(await convertWorkspace(f.deps)).toEqual({ kind: 'cancelled' })
      expect(f.applied).toEqual([])
    }
  })

  it('stops with a message when smith cannot list its questions', async () => {
    const old = fake(new Map(), { run: async () => ({ code: 2, stdout: '', stderr: 'unknown flag: --questions\n' }) })
    expect(await convertWorkspace(old.deps)).toMatchObject({ kind: 'refused', message: expect.stringContaining('Update smith') })
    const odd = fake(new Map(), { run: async () => OK('not json') })
    expect(await convertWorkspace(odd.deps)).toMatchObject({ kind: 'refused' })
  })

  it('refuses, with the words of smith, a plan it cannot make', async () => {
    const f = fake(new Map(), { run: async (args) => args.includes('--questions') ? OK(QUESTIONS) : { code: 1, stdout: '', stderr: '× the working tree has uncommitted changes\n' } })
    expect(await convertWorkspace(f.deps)).toMatchObject({ kind: 'refused', message: expect.stringContaining('uncommitted changes') })
    expect(f.applied).toEqual([])
  })
})

describe('toConvertArguments', () => {
  it('names no choice for a conservative conversion', () => {
    expect(toConvertArguments({ level: 'conservative', lint: 'smith' }, false)).toEqual(['convert', '--level', 'conservative', '--dry-run'])
  })

  it('takes smith\'s defaults for a choice not answered at the full level', () => {
    expect(toConvertArguments({ level: 'full' }, true)).toEqual(['convert', '--level', 'full', '--lint', 'keep', '--test', 'keep', '--build', 'smith', '--ci', 'smith', '--yes'])
  })
})
