/**
 * The version number in the text smith prints for `--version`
 * (`smith version 1.56.0`), or undefined when there is none.
 */
export function parseVersion (text: string): string | undefined {
  return /\d+\.\d+\.\d+(?:-[0-9A-Z.-]+)?/i.exec(text)?.[0]
}
