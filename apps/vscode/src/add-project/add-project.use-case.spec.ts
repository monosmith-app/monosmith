import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { addProject, type AddDeps } from './add-project.use-case.ts'
import type { AddOptions, AddRequest, PlanAnswer } from './add-options.contract.ts'

const OK = (stdout: string): SmithRun => ({ code: 0, stdout, stderr: '' })
const KINDS = OK(JSON.stringify([{ kind: 'node-app', language: 'node', label: 'Node app', description: 'd' }]))
const CHOICES = OK(JSON.stringify([{ axis: 'test', name: 'vitest', description: 'Vitest', default: true }]))
const STACKS = OK(JSON.stringify([{ name: 'node-vitest', description: 'd', language: 'node' }]))
const COMPOSE = OK(JSON.stringify({ outcomes: [{ id: 'web', title: 'A website or web app', kind: 'app', languages: ['node'], words: ['website'], items: [] }], items: [{ kind: 'language', name: 'node', title: 'TypeScript / JavaScript' }] }))
const REQUEST: AddRequest = { role: 'app', name: 'web', language: 'node', with: ['vitest'], fresh: false }

function fake (options: { plan?: SmithRun, apply?: SmithRun, request?: AddRequest, answer?: PlanAnswer, lists?: SmithRun[] }) {
  const ran: string[][] = []
  const applied: string[][] = []
  const shown: Array<{ plan: string, follows: boolean, choices: string }> = []
  const offered: AddOptions[] = []
  const deps: AddDeps = {
    run: async (args) => {
      ran.push(args)
      if (args[0] === 'kinds') return options.lists?.[0] ?? KINDS
      if (args[0] === 'choices') return options.lists?.[1] ?? CHOICES

      if (args[0] === 'compose') return options.lists?.[3] ?? COMPOSE

      return args[0] === 'stacks' ? options.lists?.[2] ?? STACKS : options.plan ?? OK('Plan: add app web\nDry run: nothing was changed.\n')
    },
    apply: async (args) => {
      applied.push(args)

      return options.apply ?? OK('Done: app web is ready.\n')
    },
    ask: async (asked) => {
      offered.push(asked)

      return options.request ?? REQUEST
    },
    confirm: async (_request, plan, follows, choices) => {
      shown.push({ plan, follows, choices })

      return options.answer ?? 'add'
    },
  }

  return { deps, ran, applied, shown, offered }
}

describe('addProject', () => {
  it('offers the outcomes that smith compose --json lists, not a list of its own', async () => {
    const f = fake({})
    await addProject(f.deps)
    expect(f.ran).toContainEqual(['compose', '--json'])
    expect(f.offered[0]?.outcomes.map((outcome) => outcome.id)).toEqual(['web'])
    expect(f.offered[0]?.languages[0]?.title).toBe('TypeScript / JavaScript')
  })

  it('still asks, with no outcomes, when an older smith cannot print them', async () => {
    const f = fake({ lists: [OK(KINDS.stdout), OK(CHOICES.stdout), OK(STACKS.stdout), { code: 1, stdout: '', stderr: 'unknown flag\n' }] })
    expect(await addProject(f.deps)).toMatchObject({ kind: 'added' })
    expect(f.offered[0]?.outcomes).toEqual([])
  })

  it('hands over to the composer when the person chose Compose a stack, and adds nothing', async () => {
    const f = fake({})
    f.deps.ask = async () => 'compose'
    expect(await addProject(f.deps)).toEqual({ kind: 'composing' })
    expect(f.applied).toEqual([])
  })

  it('plans with --dry-run, shows the plan, then adds with --yes', async () => {
    const f = fake({})
    expect(await addProject(f.deps)).toEqual({ kind: 'added', name: 'web', role: 'app', operation: 'add app web', lines: ['Done: app web is ready.'], committed: false })
    expect(f.ran.at(-1)).toEqual(['add', 'app', 'web', '--lang', 'node', '--with', 'vitest', '--dry-run'])
    expect(f.applied).toEqual([['add', 'app', 'web', '--lang', 'node', '--with', 'vitest', '--yes']])
    expect(f.shown[0]?.plan).toContain('Plan: add app web')
  })

  it('leaves the closing "Dry run" line out of the plan the person reads', async () => {
    const f = fake({})
    await addProject(f.deps)
    expect(f.shown[0]?.plan).not.toContain('Dry run')
  })

  it('does not say it can be undone when smith printed "Not committed" for a commit that was asked', async () => {
    const f = fake({ answer: 'add-commit', apply: OK('Done: app web is ready.\nNot committed. Your work tree had other changes.\n') })
    expect(await addProject(f.deps)).toMatchObject({ kind: 'added', committed: false })
  })

  it('tells the confirmation what was chosen, in words', async () => {
    const f = fake({})
    await addProject(f.deps)
    expect(f.shown[0]?.choices).toBe('Your choices: vitest (default).')
  })

  it('says the project was made when its first check could not run for a missing tool', async () => {
    const stdout = 'Creating app api-cs in apps/api-cs/: a console app\nWrote 12 file(s):\nVerifying: smith run build test -p api-cs\ndotnet is not installed or not on PATH. Install it from https://dot.net/download, or run smith tools install.\n'
    const f = fake({ apply: { code: 1, stdout, stderr: 'api-cs was created and its files are kept, but a tool it needs is missing.\n' }, request: { ...REQUEST, name: 'api-cs', language: 'dotnet' } })
    expect(await addProject(f.deps)).toEqual({
      kind:  'unchecked',
      name:  'api-cs',
      lines: stdout.trim().split('\n'),
      tool:  { title: 'the .NET SDK', files: 12, url: 'https://dot.net/download', dir: 'apps/api-cs', check: ['run', 'build', 'test', '-p', 'api-cs'] },
    })
  })

  it('says a project that was written but fails its doctor was written, and keeps smith steps', async () => {
    const stdout = 'Wrote 32 file(s)\nThe new project web fails smith doctor:\n  ✖ pkg-config does not find webkit2gtk-4.1\n'
    const stderr = 'smith: the new project web was written but fails smith doctor; follow the steps above, then run smith doctor\n'
    const f = fake({ apply: { code: 1, stdout, stderr }, answer: 'add-commit' })
    const outcome = await addProject(f.deps)
    expect(outcome).toMatchObject({ kind: 'failing', name: 'web', operation: 'add app web', failure: { files: 32, committed: false } })
    expect(outcome.kind === 'failing' ? outcome.lines : []).toContain('The new project web fails smith doctor:')
  })

  it('refuses, with smith words, a failure that is not a missing tool', async () => {
    const f = fake({ apply: { code: 1, stdout: '', stderr: '✖ the new project web fails its first build\n' } })
    expect(await addProject(f.deps)).toMatchObject({ kind: 'refused' })
  })

  it('tells the confirmation when the project follows its siblings, and adds with --fresh when asked', async () => {
    const f = fake({ plan: OK('Plan: add app web\n  follows api\n'), answer: 'fresh' })
    await addProject(f.deps)
    expect(f.shown[0]?.follows).toBe(true)
    expect(f.applied[0]).toContain('--fresh')
  })

  it('adds --commit only when the person ticked the box', async () => {
    const f = fake({ answer: 'add-commit' })
    expect(await addProject(f.deps)).toMatchObject({ kind: 'added', committed: true })
    expect(f.applied[0]?.at(-1)).toBe('--commit')
    const fresh = fake({ plan: OK('Plan: add app web\n  follows api\n'), answer: 'fresh-commit' })
    await addProject(fresh.deps)
    expect(fresh.applied[0]).toEqual(expect.arrayContaining(['--fresh', '--commit']))
    const plain = fake({})
    await addProject(plain.deps)
    expect(plain.applied[0]).not.toContain('--commit')
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
