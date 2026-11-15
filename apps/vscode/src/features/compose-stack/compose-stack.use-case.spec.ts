import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import type { ComposeAnswer } from './compose-stack.contract.ts'
import { composeStack, type ComposeDeps } from './compose-stack.use-case.ts'

const ITEMS = [
  { id: 'node/react', kind: 'choice', name: 'react', title: 'React', language: 'node', layer: 'framework' },
  { id: 'node/electron', kind: 'choice', name: 'electron', title: 'Electron', language: 'node', layer: 'shell' },
  { id: 'node/azure-functions', kind: 'choice', name: 'azure-functions', title: 'Azure Functions', language: 'node', layer: 'host' },
  { id: 'node/http-trigger', kind: 'choice', name: 'http-trigger', title: 'HTTP trigger', language: 'node', layer: 'host' },
  { id: 'node/eslint', kind: 'choice', name: 'eslint', title: 'ESLint', language: 'node', layer: 'lint' },
  { id: 'node/jest', kind: 'choice', name: 'jest', title: 'Jest', language: 'node', layer: 'test' },
  { id: 'plugin/slices', kind: 'plugin', name: 'slices', title: 'slices', language: '', layer: 'architecture', route: 'smith plugin enable slices' },
]

function ok (body: unknown): SmithRun {
  return { code: 0, stdout: JSON.stringify(body), stderr: '' }
}

/** A smith that answers compose --json from a table keyed by the --with value. */
function smith (checks: Record<string, unknown>): (args: string[]) => Promise<SmithRun> {
  return async (args) => {
    const withValue = args[args.indexOf('--with') + 1] ?? ''
    if (!args.includes('--with')) return ok({ layers: [], items: ITEMS })

    return ok({ layers: [], items: ITEMS, ...(args.includes('--name') ? { plan: PLAN } : { check: checks[withValue] }) })
  }
}

const COMPLETE = { language: 'node', complete: true, selection: ['electron', 'eslint', 'vitest'], added: ['eslint', 'vitest'] }
const PLAN = {
  command:    'smith add app desk --lang node --with electron,eslint,vitest',
  argv:       ['smith', 'add', 'app', 'desk', '--lang', 'node', '--with', 'electron,eslint,vitest'],
  files:      [],
  tools:      [],
  notes:      [],
  recipeArgv: ['smith', 'recipe', 'new', '--name', 'x'],
  recipeNote: '',
}

function deps (over: Partial<ComposeDeps> & { picks: string[][], reply?: ComposeAnswer }): { deps: ComposeDeps, applied: string[][], told: string[], messages: string[], stacks: string[], installs: string[] } {
  const applied: string[][] = []
  const told: string[] = []
  const messages: string[] = []
  const stacks: string[] = []
  const installs: string[] = []
  const rounds = [...over.picks]

  return {
    applied,
    told,
    messages,
    stacks,
    installs,
    deps: {
      run:   smith({ 'electron': COMPLETE, 'react,electron': { language: 'node', complete: false, conflict: 'react and electron are both on the framework axis: keep one' } }),
      apply: async (args) => {
        applied.push(args)

        return { code: 0, stdout: 'Done\n', stderr: '' }
      },
      pick: async (_items, _ticked, message) => {
        messages.push(message)

        return rounds.shift()
      },
      answer:  async () => undefined,
      askName: async (stack) => {
        stacks.push(stack)

        return 'desk'
      },
      installTools: async () => { installs.push('tools') },
      tell:         (message) => { told.push(message) },
      confirm:      async () => over.reply ?? 'run',
      ...over,
    },
  }
}

describe('composeStack', () => {
  it('sends lint and test choices alone to smith to check: no choice that writes the project is asked for', async () => {
    const checked: string[] = []
    const f = deps({
      picks: [['node/eslint', 'node/jest']],
      run:   async (args) => {
        if (args.includes('--with') && !args.includes('--name')) checked.push(args[args.indexOf('--with') + 1] ?? '')

        return await smith({ 'eslint,jest': { ...COMPLETE, selection: ['eslint', 'jest'] } })(args)
      },
    })
    await composeStack(f.deps)
    expect(checked).toEqual(['eslint,jest'])
    expect(f.messages).toEqual([''])
  })

  it('does not say a choice that writes the project is needed when only a plugin is ticked', async () => {
    const f = deps({ picks: [['plugin/slices']] })
    await composeStack(f.deps)
    expect(f.messages[1]).toBe('Tick at least one choice. To start from the defaults of the stack, pick the defaults in Add project.')
    expect(f.messages[1]).not.toContain('writes the project')
  })

  it('runs the exact add line the plan names, with --yes, and never commits unless asked', async () => {
    const f = deps({ picks: [['node/electron']] })
    const outcome = await composeStack(f.deps)
    expect(outcome).toMatchObject({ kind: 'added', name: 'desk', committed: false })
    expect(f.applied).toEqual([['add', 'app', 'desk', '--lang', 'node', '--with', 'electron,eslint,vitest', '--yes']])
  })

  it('asks the name before the plan and names the stack in the question', async () => {
    const f = deps({ picks: [['node/electron']] })
    await composeStack(f.deps)
    expect(f.stacks).toEqual(['Electron (node)'])
  })

  it('offers the tools flow when a tool is missing, and shows the plan again with the states read after it', async () => {
    const states = ['missing', 'installed']
    const asked: boolean[] = []
    const answers: ComposeAnswer[] = ['tools', 'run']
    const f = deps({
      picks: [['node/electron']],
      run:   async (args) => {
        return args.includes('--name') ? ok({ layers: [], items: ITEMS, plan: { ...PLAN, tools: [{ choice: 'tauri', name: 'tauri-linux-libs', title: 'the Tauri system libraries', binary: 'pkg-config', state: states.shift() ?? 'installed', command: 'smith tools install tauri-linux-libs' }] } }) : (await smith({ electron: COMPLETE })(args))
      },
      confirm: async (_text, _recipe, missing) => {
        asked.push(missing)

        return answers.shift() ?? 'cancel'
      },
    })
    expect(await composeStack(f.deps)).toMatchObject({ kind: 'added' })
    expect(asked).toEqual([true, false])
    expect(f.installs).toEqual(['tools'])
    expect(f.applied).toHaveLength(1)
  })

  it('adds --commit only for the commit answer', async () => {
    const f = deps({ picks: [['node/electron']], reply: 'run-commit' })
    await composeStack(f.deps)
    expect(f.applied[0]).toContain('--commit')
  })

  it('explains a conflict, takes the new picks, and writes nothing in between', async () => {
    const f = deps({ picks: [['node/react', 'node/electron'], ['node/electron']] })
    const outcome = await composeStack(f.deps)
    expect(f.messages[1]).toContain('Conflict: react and electron are both on the framework axis')
    expect(outcome.kind).toBe('added')
    expect(f.applied).toHaveLength(1)
  })

  it('swaps a second pick on a single-choice axis for the first one and says so, without a conflict', async () => {
    const withAxis = ITEMS.map((item) => item.kind === 'choice' && (item.name === 'react' || item.name === 'electron') ? { ...item, axis: 'framework' } : item)
    const base = smith({ electron: COMPLETE, react: { language: 'node', complete: false, missing: 'tick a shell' } })
    const f = deps({
      picks: [['node/react'], ['node/react', 'node/electron']],
      run:   async (args) => {
        const run = await base(args)

        return run.stdout.includes('"items"') ? ok({ ...JSON.parse(run.stdout) as object, items: withAxis }) : run
      },
    })
    const outcome = await composeStack(f.deps)
    expect(f.told.join(' ')).toContain('Electron replaces React')
    expect(f.messages.join(' ')).not.toContain('Conflict')
    expect(outcome.kind).toBe('added')
    expect(f.applied[0]).toContain('electron,eslint,vitest')
  })

  it('asks an open question and ticks the answer the person gave, not the suggested one by itself', async () => {
    const question = { language: 'node', complete: false, question: { choice: 'azure-functions', axis: 'trigger', options: [{ name: 'http-trigger', suggested: true }] } }
    const f = deps({
      picks: [['node/azure-functions']],
      run:   smith({ 'azure-functions': question, 'azure-functions,http-trigger': COMPLETE }),
    })
    let asked = 0
    f.deps.answer = async (q) => {
      asked += 1

      return q.options[0]
    }
    // After the answer the loop shows the picker again with the answer ticked; accept it as it is.
    let rounds = 0
    f.deps.pick = async (_items, ticked) => {
      rounds += 1

      return rounds === 1 ? ['node/azure-functions'] : [...ticked]
    }
    const outcome = await composeStack(f.deps)
    expect(asked).toBe(1)
    expect(outcome.kind).toBe('added')
  })

  it('does not tick a workspace plugin: it tells the command that turns it on', async () => {
    const f = deps({ picks: [['plugin/slices'], ['node/electron']] })
    await composeStack(f.deps)
    expect(f.told[0]).toContain('smith plugin enable slices')
  })

  it('copies the command without running anything', async () => {
    const f = deps({ picks: [['node/electron']], reply: 'copy' })
    expect(await composeStack(f.deps)).toEqual({ kind: 'copied', command: PLAN.command })
    expect(f.applied).toEqual([])
  })

  it('saves the plugins as a recipe with the recipe line of the plan', async () => {
    const f = deps({ picks: [['node/electron']], reply: 'recipe' })
    expect(await composeStack(f.deps)).toMatchObject({ kind: 'recipe' })
    expect(f.applied).toEqual([['recipe', 'new', '--name', 'x']])
  })

  it('says a project that was written but fails its doctor was written, with the facts and smith steps', async () => {
    const stdout = 'Creating app desk in apps/desk/: a Tauri app\nWrote 32 file(s)\nThe new project desk fails smith doctor:\n  × pkg-config does not find webkit2gtk-4.1\n      → run: sudo apt-get install libwebkit2gtk-4.1-dev\n'
    const stderr = 'smith: the new project desk was written but fails smith doctor; follow the steps above, then run smith doctor\n'
    const f = deps({
      picks: [['node/electron']],
      reply: 'run-commit',
      apply: async () => ({ code: 1, stdout, stderr }),
    })
    const outcome = await composeStack(f.deps)
    expect(outcome).toMatchObject({ kind: 'written', name: 'desk', operation: 'add app desk', failure: { files: 32, committed: false, reasons: ['pkg-config does not find webkit2gtk-4.1'] } })
    expect(outcome.kind === 'written' ? outcome.lines.join('\n') : '').toContain('sudo apt-get install libwebkit2gtk-4.1-dev')
  })

  it('never adds --commit-anyway by itself: the person decides after the facts', async () => {
    const f = deps({ picks: [['node/electron']], reply: 'run-commit', apply: async (args) => ({ code: args.includes('--commit-anyway') ? 0 : 1, stdout: '', stderr: '× the build failed\n' }) })
    expect(await composeStack(f.deps)).toMatchObject({ kind: 'refused' })
  })

  it('cancels when the person closes the list, and refuses when smith fails', async () => {
    expect(await composeStack(deps({ picks: [] }).deps)).toEqual({ kind: 'cancelled' })
    const broken = deps({ picks: [['node/electron']], run: async () => ({ code: 1, stdout: '', stderr: 'no workspace' }) })
    expect(await composeStack(broken.deps)).toMatchObject({ kind: 'refused' })
  })
})
