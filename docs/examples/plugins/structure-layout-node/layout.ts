import { existsSync, readdirSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { contributeTo, task } from '@monosmith/plugin-sdk'
import type { SmithPlugin } from '@monosmith/plugin-sdk'

const JUNK_DRAWERS = new Set(['util', 'utils', 'helper', 'helpers', 'common', 'shared', 'misc'])
const SKIPPED = new Set(['dist', 'node_modules', '.git'])

/** The entry point, which is also the program the layout-check task runs. */
const SELF = fileURLToPath(new URL('plugin.ts', import.meta.url))

/** The folders below dir named for nothing, as slash-separated paths relative to dir. */
export function findJunkDrawers (dir: string, prefix = ''): string[] {
  const found: string[] = []
  const entries = readdirSync(join(dir, prefix), { withFileTypes: true })
  for (const entry of entries) {
    if (!entry.isDirectory() || SKIPPED.has(entry.name)) continue
    const relative = prefix === '' ? entry.name : `${prefix}/${entry.name}`
    if (JUNK_DRAWERS.has(entry.name.toLowerCase())) found.push(relative)
    found.push(...findJunkDrawers(dir, relative))
  }

  return found
}

/** The plugin. It implements contribute (tasks) and doctor (findings): the SDK advertises exactly those. */
export function createLayoutPlugin (): SmithPlugin {
  let workspaceRoot = ''

  return {
    describe: (params) => {
      workspaceRoot = params.workspaceRoot

      return {
        plugin:    { name: 'widget-layout', version: '0.1.0' },
        kind:      'convention',
        roles:     ['structure'],
        // The languages it applies to. A convention that names none applies to every language.
        languages: ['widget'],
      }
    },

    // A layout-check task per project, and the project's lint (when it has one) waits for it.
    contribute: ({ projects }) => contributeTo(projects, () => ({
      tasks: {
        // No inputs listed: every file of the project is one, so a new or renamed folder is a cache miss.
        'layout-check': task(['node', SELF, 'layout-check'], {
          cache:       true,
          description: 'fail on folders named util, helpers, common, shared or misc',
        }),
      },
      // dependsOn only touches a task the project has once every earlier plugin has contributed;
      // for a project without lint it is ignored. Do not decide from project.tasks here: it holds
      // only what the language plugin discovered, not what a toolchain plugin adds.
      dependsOn: { lint: ['layout-check'] },
    })),

    // Every project needs a README.md.
    doctor: ({ projects }) => ({
      findings: projects.map((project) => {
        const check = `${project.name} has a README.md`
        const ok = existsSync(join(workspaceRoot, ...project.root.split('/'), 'README.md'))
        if (ok) return { check, ok }

        return {
          check,
          ok,
          detail:  `${project.root}/README.md does not exist`,
          remedy:  `write ${project.root}/README.md: what the project is for`,
          warning: true, // reported, but does not fail smith doctor
        }
      }),
    }),
  }
}
