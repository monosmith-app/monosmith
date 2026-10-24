import * as vscode from 'vscode'

const SCHEME = 'monosmith-insight'

/** Shows the pages explain and graph write, as read-only documents. */
export interface InsightDocuments extends vscode.Disposable {
  /**
   * Opens a Markdown page, rendered; the source is one click away.
   * @param name - the file name the tab shows, such as `web-explain.md`
   * @param text - the page
   */
  show: (name: string, text: string) => Promise<void>
}

/** Registers the read-only document scheme the pages are served under. */
export function createInsightDocuments (): InsightDocuments {
  const pages = new Map<string, string>()
  const changed = new vscode.EventEmitter<vscode.Uri>()
  const registration = vscode.workspace.registerTextDocumentContentProvider(SCHEME, {
    onDidChange:                changed.event,
    provideTextDocumentContent: (uri) => pages.get(uri.path) ?? '',
  })

  return {
    show: async (name, text) => {
      const uri = vscode.Uri.from({ scheme: SCHEME, path: `/${name}` })
      pages.set(uri.path, text)
      changed.fire(uri)
      try {
        await vscode.commands.executeCommand('markdown.showPreview', uri)
      } catch {
        await vscode.window.showTextDocument(await vscode.workspace.openTextDocument(uri), { preview: true })
      }
    },
    dispose: () => { registration.dispose(); changed.dispose() },
  }
}
