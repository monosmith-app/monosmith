import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { addProject, type AddDeps } from './add-project.use-case.ts'
import type { AddRequest, PlanAnswer } from './add-options.contract.ts'

const OK = (stdout: string): SmithRun => ({ code: 0, stdout, stderr: '' })
const KINDS = OK(JSON.stringify([{ kind: 'node-app', language: 'node', label: 'Node app', description: 'd' }]))
const CHOICES = OK(JSON.stringify([{ axis: 'test', name: 'vitest', description: 'Vitest', default: true }]))
const STACKS = OK(JSON.stringify([{ name: 'node-vitest', description: 'd', language: 'node' }]))
const REQUEST: AddRequest = { role: 'app', name: 'web', language: 'node', with: ['vitest'], fresh: false }

function fake (options: { plan?: SmithRun, apply?: SmithRun, request?: AddRequest, answer?: PlanAnswer, lists?: SmithRun[] }) {
  const ran: string[][] = []
  const applied: string[][] = []
  const shown: Array<{ plan: string, follows: boolean }> = []
  const deps: AddDeps = {
    run: async (args) => {
      ran.push(args)
      if (args[0] === 'kinds') return options.lists?.[0] ?? KINDS
      if (args[0] === 'choices') return options.lists?.[1] ?? CHOICES

      return args[0] === 'stacks' ? options.lists?.[2] ?? STACKS : options.plan ?? OK('Plan: add app web\nDry run: nothing was changed.\n')
    },
    apply: async (args) => {
      applied.push(args)

      return options.apply ?? OK('Done: app web is ready.\n')
    },
    ask:     async () => options.request ?? REQUEST,
    confirm: async (_request, plan, follows) => {
      shown.push({ plan, follows })

      return options.answer ?? 'add'
    },
  }

  return { deps, ran, applied, shown }
}

describe('addProject', () => {
  it('plans with --dry-run, shows the plan, then adds with --yes', async () => {
    const f = fake({})
    expect(await addProject(f.deps)).toEqual({ kind: 'added', name: 'web', lines: ['Done: app web is ready.'] })
    expect(f.ran.at(-1)).toEqual(['add', 'app', 'web', '--lang', 'node', '--with', 'vitest', '--dry-run'])
    expect(f.applied).toEqual([['add', 'app', 'web', '--lang', 'node', '--with', 'vitest', '--yes']])
    expect(f.shown[0]?.plan).toContain('Plan: add app web')
  })

  it('tells the confirmation when the project follows its siblings, and adds with --fresh when asked', async () => {
    const f = fake({ plan: OK('Plan: add app web\n  follows api\n'), answer: 'fresh' })
    await addProject(f.deps)
    expect(f.shown[0]?.follows).toBe(true)
    expect(f.applied[0]).toContain('--fresh')
  })

  it('adds nothing when the person cancels the questions or the plan', async () => {
    const nothing = fake({})
    nothing.deps.ask = async () => undefined
    expect(await addProject(nothing.deps)).toEqual({ kind: 'cancelled' })

    const no = fake({ answer: 'cancel' })
    expect(await addProject(no.deps)).toEqual({ kind: 'cancelled' })
    expect(no.applied).toEqual([])
  })

  it('returns the words of smith when the plan fails, and adds nothing', async () => {
    const f = fake({ plan: { code: 1, stdout: '', stderr: '✖ web exists\n  → pick another name\n' } })
    expect(await addProject(f.deps)).toEqual({ kind: 'refused', message: 'web exists Try: pick another name' })
    expect(f.applied).toEqual([])
  })

  it('returns the words of smith when the add fails', async () => {
    const f = fake({ apply: { code: 1, stdout: '', stderr: '✖ the build failed\n' } })
    expect(await addProject(f.deps)).toEqual({ kind: 'refused', message: 'the build failed' })
  })

  it('refuses with a fix when smith lists something this editor cannot read', async () => {
    const f = fake({ lists: [OK('not json')] })
    expect(await addProject(f.deps)).toMatchObject({ kind: 'refused', message: expect.stringContaining('Update the extension') as string })
  })

  it('refuses when smith cannot list the kinds', async () => {
    const f = fake({ lists: [{ code: 1, stdout: '', stderr: '✖ no smith.toml\n  → run smith new\n' }] })
    expect(await addProject(f.deps)).toEqual({ kind: 'refused', message: 'no smith.toml Try: run smith new' })
  })
})
