import type { FolderRef } from '../workspace-folders/index.ts'
import { detectRootProject } from './detect-root-project.policy.ts'
import { detectRepository, type RepositoryProbe } from './detect-repository.policy.ts'
import type { RepositoryOffer } from './repository-offer.contract.ts'

/**
 * The folders to offer MonoSmith to: the ones that are repositories smith can
 * bring in. A folder that already runs smith (`skip`) is never offered.
 */
export function findRepositoryOffers (folders: readonly FolderRef[], skip: ReadonlySet<string>, probe: RepositoryProbe): RepositoryOffer[] {
  const offers: RepositoryOffer[] = []
  for (const folder of folders) {
    if (skip.has(folder.root)) continue
    const found = detectRepository(folder.root, probe)
    if (found !== undefined) {
      offers.push({ ...folder, ...found })
      continue
    }
    const project = detectRootProject(folder.root, probe)
    if (project !== undefined) offers.push({ ...folder, kind: 'repository', marker: project.manifest, project })
  }

  return offers
}
