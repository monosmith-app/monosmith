import type { FolderState } from './folder-state.contract.ts'
import { say } from '../words/index.ts'

/**
 * A few words on a folder that is not ready, for its row in a multi-root view.
 * Empty for a folder that is.
 */
export function describeFolderNote (folder: Pick<FolderState, 'phase' | 'error'>): string {
  switch (folder.phase) {
    case 'starting': {
      return say('vscode.folder-session.starting-smith')
    }
    case 'missing': {
      return say('vscode.folder-session.smith-not-found')
    }
    case 'untrusted': {
      return say('vscode.folder-session.workspace-not-trusted')
    }
    case 'failed': {
      return folder.error === undefined ? say('vscode.folder-session.smith-failed') : say('vscode.folder-session.smith-failed-2', { item: folder.error.split('\n', 1)[0] })
    }
    case 'ready': {
      return ''
    }
  }
}
