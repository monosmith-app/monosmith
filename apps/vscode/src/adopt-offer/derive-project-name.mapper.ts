const MAX_LENGTH = 50

/**
 * Turns a manifest name into a project folder name, the way `smith adopt`
 * does: the npm scope or the module path in front is dropped, camelCase and
 * separators become single hyphens, and the result is lowercase kebab-case
 * that starts with a letter.
 * @param raw - the manifest name (`@acme/WebApp`, `github.com/acme/tool/v2`)
 * @param fallback - tried when nothing usable is left, usually the folder name
 * @returns the name, or `project` as the last resort
 */
export function deriveProjectName (raw: string, fallback: string): string {
  return kebab(lastSegment(raw)) || kebab(lastSegment(fallback)) || 'project'
}

/** The part after the scope or the last slash of a module path; a trailing major version (`/v2`) is skipped. */
function lastSegment (raw: string): string {
  const parts = raw.trim().replaceAll(/^\/+|\/+$/g, '').split('/')
  while (parts.length > 1 && /^v\d+$/.test(parts.at(-1) ?? '')) parts.pop()

  return parts.at(-1) ?? ''
}

function kebab (text: string): string {
  const name = text
    .replaceAll(/([a-z0-9])([A-Z])/g, '$1-$2')
    .toLowerCase()
    .replaceAll(/[^a-z0-9]+/g, '-')
    .replaceAll(/^-+|-+$/g, '')

  return /^[a-z]/.test(name) ? name.slice(0, MAX_LENGTH).replace(/-+$/, '') : ''
}
