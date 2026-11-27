import { describe, expect, it } from 'vitest'
import { settingRowAnswers } from './setting-row-answers.mapper.ts'

describe('settingRowAnswers', () => {
  it('answers which setting a change or a reset is about', () => {
    expect(settingRowAnswers('configSet', { root: '/r', setting: 'ci.devcontainer' })).toEqual({ setting: 'ci.devcontainer' })
    expect(settingRowAnswers('configUnset', { root: '/r', setting: 'ci.devcontainer' })).toEqual({ setting: 'ci.devcontainer' })
  })

  it('answers which plugin to switch', () => {
    expect(settingRowAnswers('pluginEnable', { plugin: 'go' })).toEqual({ plugin: 'go' })
    expect(settingRowAnswers('pluginDisable', { plugin: 'go' })).toEqual({ plugin: 'go' })
  })

  it('answers nothing for another command, another row or no row', () => {
    expect(settingRowAnswers('cacheClear', { setting: 'ci.devcontainer' })).toEqual({})
    expect(settingRowAnswers('configSet', { kind: 'project', project: 'web' })).toEqual({})
    expect(settingRowAnswers('configSet', undefined)).toEqual({})
    expect(settingRowAnswers('configSet', { setting: '' })).toEqual({})
  })
})
