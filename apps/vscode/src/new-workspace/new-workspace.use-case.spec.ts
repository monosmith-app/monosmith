import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import type { NewWorkspaceChoices } from './new-workspace.contract.ts'
import { createWorkspace, type NewWorkspaceDeps, type SetupWay } from './new-workspace.use-case.ts'

const OK = (stdout: string): SmithRun => ({ code: 0, stdout, stderr: '' })

const QUESTIONS = JSON.stringify([
  { key: 'ci', prompt: 'Which CI?', default: 'github', flag: 'ci', many: true, options: [{ value: 'github', label: 'GitHub Actions' }] },
  { key: 'language', prompt: 'Which language?', default: 'later', flag: 'language', options: [{ value: 'later', label: 'Decide later' }, { value: 'node', label: 'TypeScript / JavaScript' }] },
  { key: 'architecture', prompt: 'Which layout?', default: 'vertical-slices', flag: 'architecture', options: [{ value: 'vertical-slices', label: 'Vertical slices' }] },
])

function fake (over: Partial<NewWorkspaceDeps> & { way?: SetupWay | undefined, name?: string | undefined } = {}) {
  const ran: string[][] = []
  const applied: string[][] = []
  const shown: Array<{ summary: readonly string[], plan: string }> = []
  const offered: Array<readonly string[]> = []
  const named: string[] = []
  const deps: NewWorkspaceDeps = {
    folderName: 'ws-j',
    run:        async (args) => {
      if (args.includes('--questions')) return OK(QUESTIONS)
      ran.push(args)

      return OK('Plan: create smith.toml\n')
    },
    apply: async (args) => {
      applied.push(args)

      return OK('Created ws-j\n')
    },
    askWay: async (defaults) => {
      offered.push(defaults)

      return 'way' in over ? over.way : 'defaults'
    },
    askName: async (folder) => {
      named.push(folder)

      return 'name' in over ? over.name : folder
    },
    askCustom: async (start): Promise<NewWorkspaceChoices> => ({ ...start, language: 'node', commit: false }),
    confirm:   async (summary, plan) => {
      shown.push({ summary, plan })

      return true
    },
    ...over,
  }

  return { deps, ran, applied, shown, offered, named }
}

describe('createWorkspace', () => {
  it('lists what the defaults choose before it asks anything else, and asks only the name', async () => {
    const f = fake()
    await createWorkspace(f.deps)
    expect(f.offered[0]).toEqual(expect.arrayContaining(['CI: GitHub Actions', 'Main language: Decide later', 'Code layout smith checks: Vertical slices']))
    expect(f.named).toEqual(['ws-j'])
  })

  it('previews with --dry-run, shows the summary and the files, and only then writes', async () => {
    const f = fake()
    expect(await createWorkspace(f.deps)).toEqual({ kind: 'created', name: 'ws-j', lines: ['Created ws-j'], committed: true })
    expect(f.ran[0]?.at(-1)).toBe('--dry-run')
    expect(f.applied[0]).not.toContain('--dry-run')
    expect(f.shown[0]?.summary[0]).toBe('Name: ws-j (the name of this folder)')
    expect(f.shown[0]?.plan).toContain('smith.toml')
  })

  it('asks every question when the person customises, and passes the answers as flags', async () => {
    const f = fake({ way: 'customise' })
    expect(await createWorkspace(f.deps)).toMatchObject({ kind: 'created', committed: false })
    expect(f.applied[0]).toEqual(expect.arrayContaining(['--language', 'node', '--no-commit']))
  })

  it('hands over to smith in a terminal when the person asks for its own questions, and writes nothing', async () => {
    const f = fake({ way: 'terminal' })
    expect(await createWorkspace(f.deps)).toEqual({ kind: 'terminal' })
    expect(f.ran).toEqual([])
    expect(f.applied).toEqual([])
  })

  it('writes nothing when the person cancels at the way, the name or the summary', async () => {
    for (const f of [fake({ way: undefined }), fake({ name: undefined }), fake({ confirm: async () => false })]) {
      expect(await createWorkspace(f.deps)).toEqual({ kind: 'cancelled' })
      expect(f.applied).toEqual([])
    }
  })

  it('refuses, with the words of smith, a folder it cannot use, before anything is written', async () => {
    const f = fake({ run: async () => ({ code: 1, stdout: '', stderr: '✖ this folder has files\n  → run smith adopt\n' }) })
    expect(await createWorkspace(f.deps)).toEqual({ kind: 'refused', message: 'this folder has files Try: run smith adopt' })
    expect(f.applied).toEqual([])
  })

  it('reads the questions of smith before the way is asked, and passes them to the custom questions', async () => {
    const asked: string[][] = []
    let seen: readonly string[] = []
    const f = fake({
      way: 'customise',
      run: async (args) => {
        asked.push(args)

        return OK(args.includes('--questions') ? QUESTIONS : 'Plan\n')
      },
      askCustom: async (start, questions) => {
        seen = questions.map((question) => question.key)

        return start
      },
    })
    await createWorkspace(f.deps)
    expect(asked[0]).toEqual(['new', '--questions', '--json'])
    expect(seen).toEqual(['ci', 'language', 'architecture'])
  })

  it('labels the summary with the words smith printed', async () => {
    const f = fake()
    await createWorkspace(f.deps)
    expect(f.offered[0]).toContain('Main language: Decide later')
  })

  it('stops with a message when smith cannot list its questions', async () => {
    const old = fake({ run: async () => ({ code: 2, stdout: '', stderr: 'unknown flag: --questions\n' }) })
    expect(await createWorkspace(old.deps)).toMatchObject({ kind: 'refused' })
    const odd = fake({ run: async () => OK('not json') })
    expect(await createWorkspace(odd.deps)).toMatchObject({ kind: 'refused', message: expect.stringContaining('Update smith') })
    expect(old.applied).toEqual([])
  })
})
