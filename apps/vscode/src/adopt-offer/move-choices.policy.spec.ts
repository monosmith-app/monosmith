import { describe, expect, it } from 'vitest'
import { checkFolder, destinationFor, moveChoices } from './move-choices.policy.ts'

const project = (app: boolean) => ({ manifest: 'package.json', ecosystem: 'node' as const, name: 'web', app })

describe('moveChoices', () => {
  it('suggests an app folder for a program and a package folder for the rest', () => {
    expect(moveChoices(project(true)).find((choice) => choice.suggested)?.dir).toBe('apps/web')
    expect(moveChoices(project(false)).find((choice) => choice.suggested)?.dir).toBe('packages/web')
  })

  it('lists the five answers in smith order', () => {
    expect(moveChoices(project(true)).map((choice) => choice.key)).toEqual(['apps', 'packages', 'libs', 'other', 'none'])
  })
})

describe('destinationFor', () => {
  const [apps, , libs, other] = moveChoices(project(true))

  it('joins the folder and a valid name', () => {
    expect(destinationFor(apps, 'web-ui')).toEqual({ dir: 'apps/web-ui' })
    expect(destinationFor(libs, 'core')).toEqual({ dir: 'libs/core' })
  })

  it('refuses a name that is not kebab-case', () => {
    expect(destinationFor(apps, 'Web App')).toHaveProperty('problem')
    expect(destinationFor(apps, 'web_app')).toHaveProperty('problem')
  })

  it('takes the whole typed text for another folder', () => {
    expect(destinationFor(other, 'tools/web')).toEqual({ dir: 'tools/web' })
  })
})

describe('checkFolder', () => {
  it.each(['.', '', '../x', '/abs/x', String.raw`C:\x`, '.github/x', 'a b'])('refuses %j', (folder) => {
    expect(checkFolder(folder)).toHaveProperty('problem')
  })

  it('cleans backslashes and a leading dot segment', () => {
    expect(checkFolder(String.raw`tools\web`)).toEqual({ dir: 'tools/web' })
    expect(checkFolder('./tools/web/')).toEqual({ dir: 'tools/web' })
  })
})
