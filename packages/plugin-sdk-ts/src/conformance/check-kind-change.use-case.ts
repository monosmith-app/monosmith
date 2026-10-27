import { existsSync, readFileSync } from 'node:fs'
import { join, posix } from 'node:path'
import type { PluginClient } from '../host/index.ts'
import { projectKindDirectory } from '../projects/index.ts'
import type { Project, ProjectKind } from '../protocol/index.ts'
import type { ConformanceReport } from './conformance-report.model.ts'
import { findManifests } from './find-manifests.algorithm.ts'
import { goList } from './go-list.mapper.ts'

/** The kind a generated project is asked to become: a lib or an app is published as a package, a package goes back to a lib. */
function kindChangeTarget (kind: ProjectKind): ProjectKind {
  if (kind === 'app') return 'lib'

  return kind === 'lib' ? 'package' : 'lib'
}

function article (kind: ProjectKind): string {
  return kind === 'app' ? 'an app' : `a ${kind}`
}

/** The content of every manifest of the workspace, by path. */
function snapshot (workspace: string, names: readonly string[]): Map<string, string> {
  const content = new Map<string, string>()
  for (const manifest of findManifests(workspace, names)) {
    const text = readFileSync(join(workspace, ...manifest.split('/')), 'utf8')
    content.set(manifest, text)
  }

  return content
}

function same (a: Map<string, string>, b: Map<string, string>): boolean {
  return a.size === b.size && [...a].every(([path, text]) => b.get(path) === text)
}

/**
 * Asks a plugin that declares `changeKind` what making a generated project
 * another kind edits, and checks it answers with a refusal or with edits to
 * files that exist, and writes nothing itself: the host applies the edits.
 */
export async function checkKindChange (client: PluginClient, workspace: string, project: Project, all: readonly Project[], report: ConformanceReport): Promise<void> {
  const to = kindChangeTarget(project.kind)
  const newRoot = `${projectKindDirectory(to)}/${posix.basename(project.root)}`
  const check = `answers what making the ${project.kind} ${article(to)} edits (dry run)`
  const names = client.info.manifests ?? []
  const before = snapshot(workspace, names)
  let answer
  try {
    answer = await client.call('project/kind', { project, to, newRoot, projects: [...all] }, { timeoutMs: 5 * 60_000 })
  } catch (error) {
    report.add(check, false, error instanceof Error ? error.message : String(error))

    return
  }
  const files = answer.files ?? []
  const missing = files.map(file => file.path).filter(path => !existsSync(join(workspace, ...path.split('/'))))
  if ((answer.refused ?? '') === '' && files.length === 0) {
    report.add(check, false, 'it answered with neither a refusal nor an edit: a changed kind is recorded in a file')
  } else if (missing.length > 0) {
    report.add(check, false, `edits files that do not exist (paths are those before the move): ${goList(missing)}`)
  } else {
    const untouched = same(before, snapshot(workspace, names))
    report.add(check, untouched, 'it wrote to the workspace: it must only answer with the new content, and let the host write it')
  }
}
