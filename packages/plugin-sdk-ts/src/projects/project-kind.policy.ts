import type { ProjectKind } from '../protocol/index.ts'

const KINDS: ReadonlySet<string> = new Set(['app', 'lib', 'package'])
const DIRECTORY_KIND: Readonly<Record<string, ProjectKind>> = { apps: 'app', libs: 'lib', packages: 'package' }

/**
 * A project's kind when its manifest may declare none: the declared kind when
 * it is valid; otherwise the directory convention (`apps/` is an app, `libs/`
 * a lib, `packages/` a package); and finally `fallback`.
 */
export function resolveProjectKind (declared: string | undefined, root: string, fallback: ProjectKind): ProjectKind {
  const kind = (declared ?? '').trim()
  if (KINDS.has(kind)) return kind as ProjectKind
  const top = root.split('/', 1)[0] ?? ''

  return Object.hasOwn(DIRECTORY_KIND, top) ? DIRECTORY_KIND[top] : fallback
}

/** The conventional parent directory of a kind: `apps`, `libs` or `packages`. */
export function projectKindDirectory (kind: ProjectKind): string {
  switch (kind) {
    case 'app': { return 'apps'
    }
    case 'lib': { return 'libs'
    }
    default: { return 'packages'
    }
  }
}
