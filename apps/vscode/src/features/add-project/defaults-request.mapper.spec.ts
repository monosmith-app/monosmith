import { describe, expect, it } from 'vitest'
import { canStartWithDefaults, defaultsRequest } from './defaults-request.mapper.ts'

describe('defaultsRequest', () => {
  it('keeps the pick, follows the siblings, and builds a web app for lcl, dev, uat and prd with prd as the build', () => {
    expect(defaultsRequest({ role: 'app', language: 'node', preset: ['react'], outcomeId: 'web' }, 'web', [])).toEqual({
      role:         'app',
      name:         'web',
      language:     'node',
      with:         ['react'],
      fresh:        false,
      environments: { names: ['lcl', 'dev', 'uat', 'prd'], default: 'prd', output: 'workspace', envFiles: 'commit' },
    })
  })

  it('uses the environments of the workspace when it has them, and the last one builds when prd is not there', () => {
    const request = defaultsRequest({ role: 'app', language: 'node', preset: ['vue'] }, 'site', ['qa', 'stage'])
    expect(request.environments).toMatchObject({ names: ['qa', 'stage'], default: 'stage' })
  })

  it('asks nothing about environments for an app that is not a website, and takes no choice for "decide later"', () => {
    expect(defaultsRequest({ role: 'app', language: 'go', preset: [] }, 'api', [])).toEqual({ role: 'app', name: 'api', language: 'go', with: [], fresh: false })
  })

  it('does not offer the defaults for a function app, whose trigger smith does not choose', () => {
    expect(canStartWithDefaults({ outcomeId: 'function' })).toBe(false)
    expect(canStartWithDefaults({ outcomeId: 'web' })).toBe(true)
  })
})
