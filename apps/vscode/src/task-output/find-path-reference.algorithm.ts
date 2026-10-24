import type { PathReference } from './parsed-diagnostic.contract.ts'

const REFERENCE = /(?<![\w:@~/\\.-])(?<file>(?:[A-Z]:)?[\w@./\\~-]*[\w@-]\.[A-Z0-9]+)(?::(?<line>\d+)(?::(?<col>\d+))?|\((?<line2>\d+),(?<col2>\d+)\))/gi

/**
 * Finds the places in files a line of output names: `path:line`,
 * `path:line:col` and `path(line,col)`. A path without a line is not a place.
 */
export function findPathReferences (line: string): PathReference[] {
  const found: PathReference[] = []
  for (const match of line.matchAll(REFERENCE)) {
    const groups = match.groups
    const file = groups?.file
    // A path that follows a colon is the tail of a url (`host:3000/x.js:4`), not a place.
    if (groups === undefined || file === undefined || match.index === undefined || line[match.index - 1] === ':') continue
    const column = groups.col ?? groups.col2
    found.push({
      start: match.index,
      end:   match.index + match[0].length,
      file,
      line:  Number(groups.line ?? groups.line2),
      ...(column !== undefined && { column: Number(column) }),
    })
  }

  return found
}
