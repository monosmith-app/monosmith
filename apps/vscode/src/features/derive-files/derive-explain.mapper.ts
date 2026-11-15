import { say } from '../words/index.ts'
/**
 * Lays any text smith printed out as a Markdown page with a title, for a plan or a list the person reads once.
 * @param title - the heading
 * @param text - what smith printed
 */
export function toTextDocument (title: string, text: string): string {
  return [`# ${title}`, '', '```text', text.trimEnd(), '```', ''].join('\n')
}

/**
 * Lays the text of `smith derive --explain` out as a Markdown page. The text
 * keeps its lines; a diff in it gets its colours from the `diff` fence.
 * @param file - the file the text explains
 * @param text - what smith printed
 */
export function toExplainDocument (file: string, text: string): string {
  return [say('vscode.derive-files.derived-file', { file }), '', '```diff', text.trimEnd(), '```', '', say('vscode.derive-files.apply-adopt-and-eject-are-in'), ''].join('\n')
}
