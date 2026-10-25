import { describe, expect, it } from 'vitest'
import { encodeResult } from '../codec/index.ts'
import { actionStep, ciStep, credential } from './ci-contribution.algorithm.ts'

describe('CI builders', () => {
  it('keeps the inputs of an action in the order given', () => {
    expect(actionStep('actions/setup-node@v6')).toEqual({ uses: 'actions/setup-node@v6' })
    expect(actionStep('actions/setup-go@v7', { 'go-version': '1.26', 'cache': 'false' }).inputs).toEqual([{ name: 'go-version', value: '1.26' }, { name: 'cache', value: 'false' }])
    expect(actionStep('x', [['b', '1'], ['a', '2']]).inputs?.map(i => i.name)).toEqual(['b', 'a'])
  })

  it('blocks unless told not to', () => {
    expect(ciStep('lint', ['eslint', '.'])).toEqual({ name: 'lint', command: ['eslint', '.'], blocking: true })
    expect(ciStep('audit', ['npm', 'audit'], { blocking: false, dir: 'web', cacheFiles: ['package-lock.json'] }))
      .toEqual({ name: 'audit', command: ['npm', 'audit'], blocking: false, dir: 'web', cacheFiles: ['package-lock.json'] })
  })

  it('requires a credential unless told not to', () => {
    expect(credential('NPM_TOKEN', 'publishes to npm')).toEqual({ env: 'NPM_TOKEN', description: 'publishes to npm', required: true })
    expect(credential('X', 'y', { required: false, setup: true })).toEqual({ env: 'X', description: 'y', required: false, setup: true })
  })

  it('makes a result the codec accepts', () => {
    const json = encodeResult('ci', {
      toolchains:  [{ name: 'node', version: '24', github: actionStep('actions/setup-node@v6', { 'node-version': '24' }) }],
      setup:       [ciStep('install', ['npm', 'ci'])],
      credentials: [credential('NPM_TOKEN', 'publishes to npm')],
    })
    expect(json).toBe('{"toolchains":[{"name":"node","version":"24","github":{"uses":"actions/setup-node@v6","inputs":[{"name":"node-version","value":"24"}]}}],"setup":[{"name":"install","command":["npm","ci"],"blocking":true}],"credentials":[{"env":"NPM_TOKEN","description":"publishes to npm","required":true}]}')
  })
})
