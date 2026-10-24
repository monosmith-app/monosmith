import * as vscode from 'vscode'
import { findPathReferences } from './find-path-reference.algorithm.ts'
import type { PathReference } from './parsed-diagnostic.contract.ts'

/** What the links need to know about the documents they are made in. */
export interface OutputLinkSource {
  /** The document selector: the schemes and, for the output panel, the channel. */
  selector: vscode.DocumentSelector
  /** Whether a document is one of the extension's. */
  owns:     (document: vscode.TextDocument) => boolean
  /**
   * The file a reference in a line means, or undefined to leave it without a
   * link.
   * @param document - where the line is
   * @param task - the task the line says it belongs to (`web:build`), if it does
   * @param reference - the place the line names
   */
  resolve:  (document: vscode.TextDocument, task: string | undefined, reference: PathReference) => string | undefined
}

/** `web:build | text`: the prefix the output channel puts before each line of a task. */
const TASK_PREFIX = /^(?<task>[^\s|][^\s:|]*:[^\s|]+) \| /

/**
 * Makes the places in the output clickable: `src/a.ts:12:5` opens the file at
 * that line and column. The paths a task prints are relative to its project,
 * so the link is found with the project the line belongs to.
 */
export function registerOutputLinks (source: OutputLinkSource): vscode.Disposable {
  return vscode.languages.registerDocumentLinkProvider(source.selector, {
    provideDocumentLinks: (document) => {
      if (!source.owns(document)) return []
      const links: vscode.DocumentLink[] = []
      for (let index = 0; index < document.lineCount; index++) {
        const text = document.lineAt(index).text
        const task = TASK_PREFIX.exec(text)?.groups?.task
        for (const reference of findPathReferences(text)) {
          const file = source.resolve(document, task, reference)
          if (file === undefined) continue
          const target = vscode.Uri.file(file).with({ fragment: `L${reference.line}${reference.column === undefined ? '' : `,${reference.column}`}` })
          const link = new vscode.DocumentLink(new vscode.Range(index, reference.start, index, reference.end), target)
          link.tooltip = 'Open the file at that line'
          links.push(link)
        }
      }

      return links
    },
  })
}
