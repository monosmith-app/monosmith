import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import { GUIDED_COMMANDS } from './guided-commands.config.ts'
import { runGuidedCommand, type GuidedDeps } from './run-guided-command.use-case.ts'

const OK = (stdout: string): SmithRun => ({ code: 0, stdout, stderr: '' })
const spec = (id: string) => GUIDED_COMMANDS.find((command) => command.id === id)!

function fake (commit: boolean | undefined, confirmed = true) {
  const ran: string[][] = []
  const deps: GuidedDeps = {
    run: async (args) => {
      ran.push(args)

      return OK('')
    },
    ask:       async () => 'architecture-value',
    show:      async () => undefined,
    confirm:   async () => confirmed,
    askCommit: async () => commit,
  }

  return { deps, ran }
}

describe('the commit box of a guided command', () => {
  it('marks every command that edits a manifest and takes --commit', () => {
    for (const id of ['environmentsAdd', 'environmentsRemove', 'tauriAdd', 'configSet', 'configUnset', 'pluginEnable', 'pluginDisable', 'architecture']) {
      expect(spec(id).commit, id).toBe(true)
    }
    expect(spec('cacheStatus').commit).toBeUndefined()
  })

  it('adds --commit only when the box is ticked', async () => {
    const ticked = fake(true)
    await runGuidedCommand(spec('configUnset'), ticked.deps, { setting: 'ci.devcontainer' })
    expect(ticked.ran.at(-1)).toEqual(['config', 'unset', 'ci.devcontainer', '--commit'])
    const unticked = fake(false)
    await runGuidedCommand(spec('configUnset'), unticked.deps, { setting: 'ci.devcontainer' })
    expect(unticked.ran.at(-1)).toEqual(['config', 'unset', 'ci.devcontainer'])
  })

  it('writes nothing when the person cancels the box or the confirmation', async () => {
    const boxCancelled = fake(undefined)
    expect(await runGuidedCommand(spec('configUnset'), boxCancelled.deps, { setting: 'x.y' })).toEqual({ kind: 'cancelled' })
    expect(boxCancelled.ran).toEqual([])
    const noConfirm = fake(true, false)
    expect(await runGuidedCommand(spec('configUnset'), noConfirm.deps, { setting: 'x.y' })).toEqual({ kind: 'cancelled' })
    expect(noConfirm.ran).toEqual([])
  })
})
