import { readdirSync, readFileSync } from 'node:fs'
import { join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'
import ts from 'typescript'
import { describe, expect, it } from 'vitest'

/**
 * ADR 0089: the user-facing text of the extension lives in the catalogue (libs/words), never in code.
 * This guard fails when a non-test source file passes raw wording to a prompt, or when package.json
 * holds a label that is not a %key%.
 *
 * Allow-list: none. A file (path under src, with /) or a package.json path that needs an exception
 * goes in the two sets below, with the reason beside it.
 */
const ALLOWED_RAW_TEXT_FILES = new Set<string>()
const ALLOWED_RAW_MANIFEST_PATHS = new Set<string>()

const SRC = fileURLToPath(new URL('..', import.meta.url))
const PACKAGE_JSON = fileURLToPath(new URL('../../package.json', import.meta.url))
const NLS_JSON = fileURLToPath(new URL('../../package.nls.json', import.meta.url))

const PROMPTS = new Set(['showInformationMessage', 'showErrorMessage', 'showWarningMessage', 'showQuickPick', 'showInputBox'])
const RAW_WORDING = /^[A-Z].{11}/
const SHOWN_NAMES = new Set(['title', 'description', 'markdownDescription', 'label', 'contents', 'displayName', 'category', 'shortTitle', 'tooltip'])
const KEY_REFERENCE = /^%[\w.-]+%$/

/** Every non-test .ts file under src, except the module wordgen writes. */
function sourceFiles (dir: string): string[] {
  const found: string[] = []
  const entries = readdirSync(dir, { withFileTypes: true })
  for (const entry of entries) {
    const path = join(dir, entry.name)
    if (entry.isDirectory()) found.push(...sourceFiles(path))
    else if (entry.name.endsWith('.ts') && !entry.name.endsWith('.spec.ts') && entry.name !== 'smith-words.config.ts') found.push(path)
  }

  return found
}

/** The literal wording of a node when it is a string or a template, else undefined. */
function wordingOf (node: ts.Node): string | undefined {
  if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) return node.text

  return ts.isTemplateExpression(node) ? node.head.text : undefined
}

/** The name of the function a call invokes: `showQuickPick` for `vscode.window.showQuickPick(...)`. */
function calleeName (call: ts.CallExpression): string {
  if (ts.isPropertyAccessExpression(call.expression)) return call.expression.name.text

  return ts.isIdentifier(call.expression) ? call.expression.text : ''
}

/** The raw wordings (12+ characters, a capital first) passed to a prompt function in a source file. */
function rawWordings (path: string): string[] {
  const file = ts.createSourceFile(path, readFileSync(path, 'utf8'), ts.ScriptTarget.Latest, true)
  const found: string[] = []
  const inside = (node: ts.Node): void => {
    const text = wordingOf(node)
    const line = file.getLineAndCharacterOfPosition(node.getStart(file)).line + 1
    if (text !== undefined && RAW_WORDING.test(text)) found.push(`${relative(SRC, path)}:${line} ${JSON.stringify(text)}`)
    ts.forEachChild(node, inside)
  }
  const visit = (node: ts.Node): void => {
    if (ts.isCallExpression(node) && PROMPTS.has(calleeName(node))) {
      for (const argument of node.arguments) inside(argument)
    }
    ts.forEachChild(node, visit)
  }
  visit(file)

  return found
}

/** The path and value of each string in package.json that shows to people and is not a %key%. */
function rawManifestTexts (value: unknown, path = ''): string[] {
  if (Array.isArray(value)) return value.flatMap((item, index) => rawManifestTexts(item, `${path}[${index}]`))
  if (value === null || typeof value !== 'object') return []
  const found: string[] = []
  for (const [name, child] of Object.entries(value)) {
    const here = path === '' ? name : `${path}.${name}`
    // A view has a name that shows; a problem matcher or the package has a name that is an id.
    const shows = SHOWN_NAMES.has(name) || (name === 'name' && path.startsWith('contributes.views'))
    if (typeof child !== 'string') found.push(...rawManifestTexts(child, here))
    else if (shows && !KEY_REFERENCE.test(child) && !ALLOWED_RAW_MANIFEST_PATHS.has(here)) found.push(`${here} ${JSON.stringify(child)}`)
  }

  return found
}

describe('raw text in the extension', () => {
  it('passes no raw wording to a message, a quick pick or an input box', () => {
    const files = sourceFiles(SRC).filter((path) => !ALLOWED_RAW_TEXT_FILES.has(relative(SRC, path).replaceAll('\\', '/')))
    const raw = files.flatMap((path) => rawWordings(path))
    expect(raw).toEqual([])
  })

  it('writes every text that package.json shows as a %key%', () => {
    const manifest: unknown = JSON.parse(readFileSync(PACKAGE_JSON, 'utf8'))
    expect(rawManifestTexts(manifest)).toEqual([])
  })

  it('has an English text in package.nls.json for every %key% of package.json', () => {
    const nls = JSON.parse(readFileSync(NLS_JSON, 'utf8')) as Record<string, string>
    const missing: string[] = []
    let count = 0
    for (const match of readFileSync(PACKAGE_JSON, 'utf8').matchAll(/"%([\w.-]+)%"/g)) {
      count++
      if (nls[match[1] ?? ''] === undefined) missing.push(match[1] ?? '')
    }
    expect(missing).toEqual([])
    expect(count).toBeGreaterThan(0)
  })
})
