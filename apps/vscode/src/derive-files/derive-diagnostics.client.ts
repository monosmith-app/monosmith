import { join } from 'node:path'
import * as vscode from 'vscode'
import { isHeldBack } from './derive-arguments.mapper.ts'
import type { DeriveReport } from './derive-report.contract.ts'
import { fixFor } from './derive-tree.mapper.ts'

/** The Problems panel entries for derived files that need action. */
export interface DeriveDiagnostics extends vscode.Disposable {
  /**
   * Replaces the entries of a folder: one warning on the first line of each derived file that needs action, with the fix.
   * @param root - the workspace folder
   * @param report - what `smith derive --check` found; a report with no file clears the folder
   */
  update:   (root: string, report: DeriveReport) => void
  /** Removes every entry. */
  clearAll: () => void
}

/** Creates the `smith derive` collection of the Problems panel. */
export function createDeriveDiagnostics (): DeriveDiagnostics {
  const collection = vscode.languages.createDiagnosticCollection('smith derive')
  const byRoot = new Map<string, vscode.Uri[]>()
  const clearRoot = (root: string): void => {
    const previous = byRoot.get(root) ?? []
    for (const uri of previous) collection.delete(uri)
    byRoot.delete(root)
  }

  return {
    update: (root, report) => {
      clearRoot(root)
      const uris: vscode.Uri[] = []
      for (const file of report.files) {
        const uri = vscode.Uri.file(join(root, file.path))
        const diagnostic = new vscode.Diagnostic(
          new vscode.Range(0, 0, 0, 0),
          `This derived file is ${file.status}${file.detail === '' ? '' : `: ${file.detail}`}. ${fixFor(file.status)}`,
          isHeldBack(file.status) ? vscode.DiagnosticSeverity.Warning : vscode.DiagnosticSeverity.Information,
        )
        diagnostic.source = 'smith derive'
        diagnostic.code = file.status
        collection.set(uri, [diagnostic])
        uris.push(uri)
      }
      byRoot.set(root, uris)
    },
    clearAll: () => { collection.clear(); byRoot.clear() },
    dispose:  () => { collection.dispose() },
  }
}
