/**
 * The last `count` non-empty lines of a text, so a crashed smith's final words
 * can be shown without keeping everything it ever said.
 */
export function lastLines (text: string, count: number): string {
  return text
    .split(/\r?\n/)
    .filter((line) => line.trim() !== '')
    .slice(-count)
    .join('\n')
}
