import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

interface MenuItem { command: string, when?: string, group?: string }
const manifest = JSON.parse(readFileSync(new URL('../../package.json', import.meta.url), 'utf8')) as {
  contributes: { menus: Record<string, MenuItem[]>, commands: Array<{ command: string, icon?: string }> }
}

function inProjectsTitle (command: string): MenuItem | undefined {
  return manifest.contributes.menus['view/title']?.find((item) => item.command === command && item.when?.includes('view == monosmith.projects') === true)
}

describe('the ways to create a project in the Projects view', () => {
  it('has a + button in the title with a project in the workspace or none', () => {
    const add = inProjectsTitle('monosmith.addProject')
    expect(add?.group).toMatch(/^navigation/)
    expect(add?.when).not.toContain('noProjects')
    expect(manifest.contributes.commands.find((c) => c.command === 'monosmith.addProject')?.icon).toBe('$(add)')
  })

  it('has a Compose a Stack entry in the title menu', () => {
    expect(inProjectsTitle('monosmith.composeStack')?.group).toBe('0_create@1')
  })
})
