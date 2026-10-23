import { describe, expect, it } from 'vitest'
import type { SmithConnection } from '../smith-client/index.ts'
import { readWorkspace } from './read-workspace.use-case.ts'

const stopListening = (): void => {}
const listenToNothing = (): (() => void) => stopListening

function connection (answers: Record<string, unknown>): SmithConnection {
  return {
    request: async <T>(method: string): Promise<T> => {
      const answer = answers[method]
      if (answer === undefined) throw new Error(`no ${method}`)

      return answer as T
    },
    onEvent: listenToNothing,
    close:   stopListening,
  }
}

describe('readWorkspace', () => {
  it('reads the projects and what is affected', async () => {
    const snapshot = await readWorkspace(connection({
      projects: [{ name: 'api', dir: 'apps/api', kind: 'app', language: 'go', tasks: ['build', 'test'], deps: ['core'] }],
      affected: { projects: ['api'] },
    }))
    expect(snapshot.projects).toEqual([
      { name: 'api', dir: 'apps/api', kind: 'app', language: 'go', tasks: ['build', 'test'], deps: ['core'] },
    ])
    expect(snapshot.affected.has('api')).toBe(true)
  })

  it('still shows the projects when affected cannot be had', async () => {
    const snapshot = await readWorkspace(connection({ projects: [{ name: 'web', root: 'apps/web', targets: ['build'] }] }))
    expect(snapshot.projects[0]).toMatchObject({ name: 'web', dir: 'apps/web', tasks: ['build'] })
    expect(snapshot.affected.size).toBe(0)
  })
})
