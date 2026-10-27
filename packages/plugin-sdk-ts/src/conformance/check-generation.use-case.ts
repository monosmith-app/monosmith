import { existsSync } from 'node:fs'
import { join } from 'node:path'
import type { PluginClient } from '../host/index.ts'
import { effectiveRoles } from '../plugin-roles/index.ts'
import { projectKindDirectory } from '../projects/index.ts'
import type { GenerateParams, Project, ProjectKind } from '../protocol/index.ts'
import { checkKindChange } from './check-kind-change.use-case.ts'
import type { ConformanceReport } from './conformance-report.model.ts'
import { findManifests } from './find-manifests.algorithm.ts'
import { goList, goQuote } from './go-list.mapper.ts'

/** The stack name generated projects are asked to record. */
export const CONFORMANCE_STACK = 'conformance'

const LONG_MS = 20 * 60 * 1000

function article (kind: ProjectKind): string {
  return kind === 'app' ? 'an app' : `a ${kind}`
}

/**
 * Asks the plugin for one project of each kind, checks the files it reports
 * exist, and, for a language plugin, that its own `discover` returns what it
 * made, with the kind and the stack it was asked for, tasks with commands, and
 * that the dry runs of `version/write`, `remove`, `publish` and `published`
 * answer, and, for a plugin that changes kinds, what `project/kind` edits.
 */
export async function checkGeneration (client: PluginClient, workspace: string, report: ConformanceReport): Promise<void> {
  const language = effectiveRoles(client.info).includes('language')
  const made: GenerateParams[] = []
  for (const kind of ['app', 'lib', 'package'] as const) {
    const name = `c${kind}`
    const params: GenerateParams = {
      kind,
      name,
      directory: `${projectKindDirectory(kind)}/${name}`,
      stack:     CONFORMANCE_STACK,
      workspace: { name: 'conformance', scope: '@conformance', modulePrefix: 'example.com/conformance' },
    }
    let files: string[]
    try {
      const generated = await client.call('generate', params, { timeoutMs: LONG_MS })
      files = generated.files
    } catch (error) {
      report.add(`generates ${article(kind)}`, false, messageOf(error))
      continue
    }
    const missing = files.filter(file => !existsSync(join(workspace, ...file.split('/'))))
    report.add(`generates ${article(kind)}`, files.length > 0 && missing.length === 0, `files reported but not written: ${goList(missing)} (of ${goList(files)})`)
    made.push(params)
  }
  if (!language || made.length === 0) return

  let projects: Project[]
  try {
    const discovered = await client.call('discover', { manifests: findManifests(workspace, client.info.manifests ?? []) }, { timeoutMs: LONG_MS })
    projects = discovered.projects
  } catch (error) {
    report.add('discovers what it generated', false, messageOf(error))

    return
  }
  const found = new Map(projects.map(project => [project.root, project]))
  for (const params of made) {
    const project = found.get(params.directory)
    report.add(`discovers the generated ${params.kind}`, project !== undefined, `no project at ${params.directory} among ${projects.length} discovered`)
    if (project === undefined) continue
    report.add(`the generated ${params.kind} keeps its kind`, project.kind === params.kind, `discovered as ${goQuote(project.kind)}`)
    report.add(`the generated ${params.kind} records its stack`, project.stack === CONFORMANCE_STACK, `discovered with stack ${goQuote(project.stack ?? '')}; generate must record params.stack in the manifest`)
    const empty = Object.entries(project.tasks ?? {}).filter(([, task]) => task.command.length === 0).map(([name]) => name).sort((a, b) => a.localeCompare(b))
    report.add(`the generated ${params.kind}'s tasks have commands`, empty.length === 0, `tasks without a command: ${goList(empty)}`)
    if (client.supports('version/write')) {
      await attempt(report, `writes the ${params.kind}'s version (dry run)`, () => client.call('version/write', { project, version: '1.2.3', dryRun: true }, { timeoutMs: 60_000 }))
    }
    if (client.supports('remove')) {
      await attempt(report, `would remove the ${params.kind} (dry run)`, () => client.call('remove', { project, dryRun: true }, { timeoutMs: 60_000 }))
    }
    if (client.supports('project/kind')) {
      await checkKindChange(client, workspace, project, projects, report)
    }
    if (params.kind === 'package' && client.supports('publish')) {
      await attempt(report, 'publishes the package (dry run)', () => client.call('publish', { project, version: '1.2.3', dryRun: true }, { timeoutMs: 5 * 60_000 }))
    }
    if (params.kind === 'package' && client.supports('published')) {
      await attempt(report, 'says what the registry has of the package', () => client.call('published', { project }, { timeoutMs: 60_000 }), ' (an unreachable registry is a result, not an error)')
    }
  }
}

async function attempt (report: ConformanceReport, check: string, call: () => Promise<unknown>, hint = ''): Promise<void> {
  try {
    await call()
    report.add(check, true, '')
  } catch (error) {
    report.add(check, false, `${messageOf(error)}${hint}`)
  }
}

function messageOf (thrown: unknown): string {
  return thrown instanceof Error ? thrown.message : String(thrown)
}
