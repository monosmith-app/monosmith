import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { installTools, type AdminAnswer, type InstallToolsDeps } from './install-tools.use-case.ts'
import type { ToolPlan } from './tool-plan.contract.ts'

const python: ToolPlan = {
  name:        'python',
  title:       'Python',
  why:         'the python projects install, build and test with Python',
  consequence: 'smith test fails for the python projects',
  download:    'https://www.python.org/downloads/',
  state:       'missing',
  default:     'install:brew',
  options:     [
    { kind: 'install', value: 'install:brew', manager: 'brew', command: ['brew', 'install', 'python@3.13'], label: 'Install with brew', hint: 'runs brew install python@3.13' },
    { kind: 'install', value: 'install:apt', manager: 'apt', command: ['sudo', 'apt-get', 'install', 'python3'], admin: true, label: 'Install with apt', hint: 'needs administrator rights' },
    { kind: 'page', value: 'page', label: 'Use the official page', hint: 'go to the page' },
    { kind: 'myself', value: 'myself', label: 'I will install it myself', hint: 'smith shows the ways' },
    { kind: 'skip', value: 'skip', label: 'Skip it', hint: 'go on without it' },
  ],
}

function fake (plans: ToolPlan[], answers: string[], admin: AdminAnswer | undefined = 'myself') {
  const calls: string[][] = []
  const terminal: string[][] = []
  const pages: string[] = []
  const shown: string[][] = []
  const deps: InstallToolsDeps = {
    run: async (args) => {
      calls.push(args)
      const dry = args.includes('--dry-run')
      const run: SmithRun = ({ code: 0, stdout: dry ? JSON.stringify(plans) : 'Running: brew install python@3.13\nPython is installed (python3 3.13.1).\n', stderr: '' })

      return run
    },
    chooseWay:     async () => answers.shift(),
    chooseAdmin:   async () => admin,
    openPage:      async (url) => { pages.push(url) },
    showCommands:  async (_plan, lines) => { shown.push(lines) },
    runInTerminal: (args) => { terminal.push(args) },
  }

  return { deps, calls, terminal, pages, shown }
}

describe('installTools', () => {
  it('says nothing is missing when smith has no plan', async () => {
    const f = fake([], [])
    expect(await installTools(f.deps)).toEqual({ kind: 'nothing' })
    expect(f.calls).toEqual([['tools', 'install', '--dry-run', '--json']])
  })

  it('reads the plan first and runs only the command of the chosen option', async () => {
    const f = fake([python], ['install:brew'])
    const outcome = await installTools(f.deps)
    expect(f.calls[1]).toEqual(['tools', 'install', 'python', '--manager', 'brew'])
    expect(outcome).toEqual({ kind: 'handled', results: [{ tool: 'python', choice: 'install:brew', lines: ['Running: brew install python@3.13', 'Python is installed (python3 3.13.1).'] }] })
  })

  it('asks a second question before a command that needs administrator rights and runs nothing for "myself"', async () => {
    const f = fake([python], ['install:apt'], 'myself')
    const outcome = await installTools(f.deps)
    expect(f.calls).toHaveLength(1)
    expect(f.terminal).toEqual([])
    expect(outcome).toMatchObject({ results: [{ lines: ['Run this yourself, then open a new terminal: sudo apt-get install python3'] }] })
  })

  it('runs an administrator command in a terminal only after "run"', async () => {
    const f = fake([python], ['install:apt'], 'run')
    await installTools(f.deps)
    expect(f.terminal).toEqual([['tools', 'install', 'python', '--manager', 'apt', '--with-admin']])
    expect(f.calls).toHaveLength(1)
  })

  it('opens the page, shows the commands, or skips, without running smith again', async () => {
    const f = fake([python, python, python], ['page', 'myself', 'skip'])
    const outcome = await installTools(f.deps)
    expect(f.calls).toHaveLength(1)
    expect(f.pages).toEqual(['https://www.python.org/downloads/'])
    expect(f.shown).toEqual([['brew install python@3.13', 'sudo apt-get install python3', 'https://www.python.org/downloads/']])
    expect(outcome).toMatchObject({ results: [{ choice: 'page' }, { choice: 'myself' }, { choice: 'skip', lines: ['Skipped Python: smith test fails for the python projects.'] }] })
  })

  it('stops when the person cancels', async () => {
    const f = fake([python, python], [])
    expect(await installTools(f.deps)).toEqual({ kind: 'handled', results: [] })
  })

  it('refuses with smith\'s words when the plan fails', async () => {
    const deps = fake([], []).deps
    deps.run = async () => ({ code: 1, stdout: '', stderr: '× this folder is not a smith workspace yet\n  → smith adopt --dry-run\n' })
    expect(await installTools(deps)).toMatchObject({ kind: 'refused' })
  })
})
