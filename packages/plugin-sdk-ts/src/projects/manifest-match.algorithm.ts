/** One element of a pattern: an escaped character, a class, a wildcard, or any other character. */
const PART = /\\(.)|\[([^\]]+)\]|([*?])|(.)/gsu

/**
 * Whether a file's base name is one of a plugin's manifests: equal to a name,
 * or matching a pattern (`*.csproj`). A pattern is Go's `path.Match`: `*`
 * matches any run of characters, `?` one, `[a-z]` and `[^a-z]` a class, `\`
 * escapes. A malformed pattern matches nothing.
 */
export function isManifest (manifests: readonly string[], base: string): boolean {
  return manifests.some(manifest => manifest === base || toRegExp(manifest)?.test(base) === true)
}

function toRegExp (pattern: string): RegExp | undefined {
  let source = '^'
  for (const [, escaped, characters, wildcard, literal] of pattern.matchAll(PART)) {
    if (escaped !== undefined) source += escape(escaped)
    else if (characters !== undefined) source += `[${characters.replaceAll('\\', String.raw`\\`)}]`
    else if (wildcard === '*') source += '[^/]*'
    else if (wildcard === '?') source += '[^/]'
    // A bracket or backslash that did not make an escape or a class is a malformed pattern.
    else if (literal === '[' || literal === '\\') return undefined
    else source += escape(literal)
  }
  try {
    return new RegExp(`${source}$`, 'u')
  } catch {
    return undefined
  }
}

function escape (character: string): string {
  return character.replaceAll(/[.*+?^${}()|[\]\\/]/g, String.raw`\$&`)
}
