import type { FolderRef } from '../workspace-folders/index.ts'
import { detectRepository, type RepositoryProbe } from './detect-repository.policy.ts'
import type { RepositoryOffer } from './repository-offer.contract.ts'

/**
 * The folders to offer MonoSmith to: the ones that are repositories smith can
 * bring in. A folder that already runs smith (`skip`) is never offered.
 */
export function findRepositoryOffers (folders: readonly FolderRef[], skip: ReadonlySet<string>, probe: RepositoryProbe): RepositoryOffer[] {
  const offers: RepositoryOffer[] = []
  for (const folder of folders) {
    const found = skip.has(folder.root) ? undefined : detectRepository(folder.root, probe)
    if (found !== undefined) offers.push({ ...folder, ...found })
  }

  return offers
}
