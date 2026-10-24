import type { FolderState } from './folder-state.contract.ts'

/**
 * A few words on a folder that is not ready, for its row in a multi-root view.
 * Empty for a folder that is.
 */
export function describeFolderNote (folder: Pick<FolderState, 'phase' | 'error'>): string {
  switch (folder.phase) {
    case 'starting': {
      return 'starting smith…'
    }
    case 'missing': {
      return 'smith not found'
    }
    case 'untrusted': {
      return 'workspace not trusted'
    }
    case 'failed': {
      return folder.error === undefined ? 'smith failed' : `smith failed: ${folder.error.split('\n', 1)[0]}`
    }
    case 'ready': {
      return ''
    }
  }
}
