/**
 * Lays the text of `smith derive --explain` out as a Markdown page. The text
 * keeps its lines; a diff in it gets its colours from the `diff` fence.
 * @param file - the file the text explains
 * @param text - what smith printed
 */
export function toExplainDocument (file: string, text: string): string {
  return [`# Derived file: ${file}`, '', '```diff', text.trimEnd(), '```', '', 'Apply, Adopt and Eject are in the context menu of the file in the Health view.', ''].join('\n')
}
