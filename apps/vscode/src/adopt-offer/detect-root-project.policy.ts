import { posix, win32 } from 'node:path'
import { deriveProjectName } from './derive-project-name.mapper.ts'
import type { RepositoryProbe } from './detect-repository.policy.ts'
import type { RootProject } from './repository-offer.contract.ts'

/** The manifests adopt can move, by ecosystem. */
const MANIFESTS: ReadonlyArray<{ file: string, ecosystem: RootProject['ecosystem'] }> = [
  { file: 'package.json', ecosystem: 'node' },
  { file: 'go.mod', ecosystem: 'go' },
  { file: 'pyproject.toml', ecosystem: 'python' },
  { file: 'Cargo.toml', ecosystem: 'rust' },
]

/** Manifests of other languages: they only make a root with another manifest ambiguous. */
const OTHER_MANIFESTS = ['pom.xml', 'pubspec.yaml', 'build.gradle', 'build.gradle.kts']

/**
 * Finds the one project a plain repository keeps at its root, the way
 * `smith adopt` does. A root with two manifests (a Go module with a
 * package.json) is ambiguous and a workspace (a package.json with
 * `workspaces`, a go.work, a Cargo `[workspace]`) is not one project: both give
 * undefined, and `smith adopt` explains them. Reads one small manifest.
 * @param root - the folder
 * @param probe - what the detection may look at
 * @returns the project, or undefined
 */
export function detectRootProject (root: string, { exists, read }: RepositoryProbe): RootProject | undefined {
  const path = root.includes('\\') ? win32 : posix
  const at = (file: string): string => path.join(root, file)
  // A folder with a smith.toml is a workspace already, readable or not.
  if (exists(at('smith.toml'))) return undefined
  const found = MANIFESTS.filter(({ file }) => exists(at(file)))
  const first = found[0]
  if (first === undefined || found.length > 1 || OTHER_MANIFESTS.some((file) => exists(at(file)))) return undefined
  const text = read(at(first.file)) ?? ''
  const folder = path.basename(root)
  const project = (raw: string, app: boolean): RootProject => ({ manifest: first.file, ecosystem: first.ecosystem, name: deriveProjectName(raw, folder), app })

  switch (first.ecosystem) {
    case 'node': {
      const manifest = parseJson(text)
      if (manifest === undefined || hasWorkspaces(manifest.workspaces)) return undefined
      const scripts = manifest.scripts as Record<string, unknown> | undefined

      return project(typeof manifest.name === 'string' ? manifest.name : '', typeof scripts?.start === 'string' || manifest.bin !== undefined)
    }
    case 'go': {
      return exists(at('go.work')) ? undefined : project(/^module\s+"?([^\s"]+)/m.exec(text)?.[1] ?? '', exists(at('main.go')))
    }
    case 'python': {
      return project(tomlName(text), text.includes('[project.scripts]') || text.includes('[tool.poetry.scripts]') || exists(at('main.py')) || exists(at('__main__.py')))
    }
    case 'rust': {
      return text.includes('[workspace]') ? undefined : project(tomlName(text), exists(path.join(root, 'src', 'main.rs')) || text.includes('[[bin]]'))
    }
  }
}

function tomlName (text: string): string {
  return /^name\s*=\s*["']([^"']+)["']/m.exec(text)?.[1] ?? ''
}

function parseJson (text: string): Record<string, unknown> | undefined {
  try {
    const parsed: unknown = JSON.parse(text)

    return typeof parsed === 'object' && parsed !== null ? parsed as Record<string, unknown> : undefined
  } catch {
    return undefined
  }
}

function hasWorkspaces (workspaces: unknown): boolean {
  return Array.isArray(workspaces) ? workspaces.length > 0 : typeof workspaces === 'object' && workspaces !== null
}
