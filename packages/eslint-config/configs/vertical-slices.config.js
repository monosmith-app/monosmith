import { readdirSync, readFileSync, statSync } from 'node:fs'
import { basename, dirname, join, relative, resolve, sep } from 'node:path'

/** The roles a production file may play, as its last suffix before the extension says. */
export const ROLES = ['handler', 'use-case', 'algorithm', 'policy', 'model', 'contract', 'mapper', 'validator', 'repository', 'client', 'store', 'middleware', 'error', 'config', 'enum']

/**
 * The front-end roles a `.tsx` file may play without any option: a file that holds JSX
 * is a UI file by construction, so asking each React workspace to opt in is ceremony.
 * The non-JSX front-end roles (`.style`, `.mock`, `.fixture`, a `.hook.ts`) stay opt-in
 * through `roles`, because a `.ts` file says nothing about being UI.
 */
export const TSX_ROLES = ['route', 'component', 'section', 'hook', 'content']

const EXTENSION = String.raw`\.(?:ts|tsx|mts|cts)`
const KEBAB = /^[a-z0-9]+(?:-[a-z0-9]+)*$/
/**
 * The name pattern a production file must match, for a given role list.
 *
 * Built per rule invocation rather than once at module load, because the role
 * list is configurable - see the `roles` option on {@link verticalSlices}. The
 * cost is one `RegExp` per linted file, which is nothing next to parsing it.
 *
 * @param roles - Every role a production file may end in.
 * @returns The matching pattern.
 */
function productionPattern (roles) {
  return new RegExp(String.raw`^[a-z0-9]+(?:-[a-z0-9]+)*\.(?:${roles.join('|')})${EXTENSION}$`)
}
const TEST = new RegExp(String.raw`\.(?:spec|test)${EXTENSION}$`)
/** The words that name no outcome, as parts of a file name (split on `-`, `.` and `_`). */
export const BANNED_WORDS = ['util', 'utils', 'helper', 'helpers', 'common', 'shared', 'manager', 'processor', 'data', 'misc']
const SOURCE = new RegExp(`${EXTENSION}$`)
// `content.config` is Astro's: a framework-mandated name that must sit at the root of src, so it is an entry like main.
const ENTRY = new RegExp(String.raw`^(?:index|main|content\.config)${EXTENSION}$`)
const BARREL = new RegExp(`^index${EXTENSION}$`)
// A slice's index named as a file. Node's own ESM resolution (TypeScript's
// `nodenext`) cannot import a directory, so `'../billing/index.ts'` (or the
// `.js` it is emitted as) is the same door as `'../billing'`.
const INDEX_TAIL = /[\\/]index\.(?:ts|tsx|mts|cts|js|jsx|mjs|cjs)$/
const EXTENSIONS = ['.ts', '.tsx', '.mts', '.cts', '.js', '.jsx', '.mjs', '.cjs']
const ASSET = /\.(?:svg|png|jpe?g|gif|webp|avif|ico|css|scss|sass|less|json|woff2?|ttf|otf|eot|md|mdx|txt|ya?ml|csv|html|mp[34]|webm|pdf|wasm)(?:\?.*)?$/
const RELATIVE_IMPORT = /(?:from|import\()\s*['"](\.[^'"]*)['"]/g

/** The folders directly under `src` that group slices instead of being one, unless the `groups` option says otherwise. */
export const GROUPS = ['features', 'integrations', 'shared']

/** Folders that hold support files (fixtures, mocks, test data): never a slice. `testing` is also exempt from the role rule. */
const FIXTURES = 'fixtures'
const TESTING = 'testing'
const SUPPORT = new Set([FIXTURES, TESTING])

/**
 * Where a file sits: its project's `src`, and the folders between `src` and the file.
 *
 * @param filename - The file being linted.
 * @returns Its place (`dirs` is empty at the root of `src`), or `null` outside any `src`.
 */
function locate (filename) {
  const parts = filename.split(sep)
  const src = parts.lastIndexOf('src')

  return src === -1 ? null : { src: parts.slice(0, src + 1).join(sep), dirs: parts.slice(src + 1, -1) }
}

/**
 * Whether a directory is a group: a folder directly under `src` whose children are the slices.
 *
 * @param src - A project's `src` directory.
 * @param directory - The directory to test.
 * @param groups - The group names.
 * @returns True for a group.
 */
function isGroup (src, directory, groups) {
  return dirname(directory) === src && groups.includes(basename(directory))
}

/**
 * Whether a directory directly holds production files, which makes it a slice: it is not a group
 * or a support folder, and a barrel alone counts.
 *
 * @param src - A project's `src` directory.
 * @param directory - The directory to test.
 * @param groups - The group names.
 * @returns True for a slice.
 */
function isSlice (src, directory, groups) {
  if (isGroup(src, directory, groups) || SUPPORT.has(basename(directory))) return false
  try {
    return readdirSync(directory, { withFileTypes: true }).some(entry => entry.isFile() && SOURCE.test(entry.name) && !TEST.test(entry.name))
  } catch {
    return false
  }
}

/**
 * The stats of a path, or `null` when nothing is there.
 *
 * @param candidate - A path.
 * @returns The stats, or `null`.
 */
function exists (candidate) {
  try {
    return statSync(candidate)
  } catch {
    return null
  }
}

/**
 * Where an import specifier points: a directory (a slice's barrel) or a file.
 *
 * @param from - The directory of the importing file.
 * @param specifier - A relative module specifier.
 * @returns The target `directory` and, for a file, its `file` path.
 */
function resolveTarget (from, specifier) {
  const path = resolve(from, specifier)
  if (INDEX_TAIL.test(path)) return { directory: dirname(path), file: null }
  if (exists(path)?.isDirectory()) return { directory: path, file: null }

  return basename(path).includes('.') || EXTENSIONS.some(extension => exists(path + extension)?.isFile()) ? { directory: dirname(path), file: path } : { directory: path, file: null }
}

/**
 * The directory a file counts as living in: a support folder (`fixtures`, `testing`) belongs to its parent.
 *
 * @param directory - The directory of a file.
 * @returns The directory with any support folder and what is under it cut off.
 */
function ownerOf (directory) {
  const parts = directory.split(sep)
  const at = parts.findIndex(part => SUPPORT.has(part))

  return at === -1 ? directory : parts.slice(0, at).join(sep)
}

/** Each project's slice graph, built once per lint run: slice path to the slices it imports. */
const graphs = new Map()

/**
 * A slice's name in the graph: its path under `src`.
 *
 * @param src - A project's `src` directory.
 * @param directory - A slice directory.
 * @returns The path with `/` separators.
 */
function sliceName (src, directory) {
  return relative(src, directory).split(sep).join('/')
}

/**
 * The slices each slice's production files import, by path under `src`. Every directory that holds
 * production files is a node; an import of a slice's barrel (or of a file in it) is an edge.
 *
 * @param src - A project's `src` directory.
 * @param groups - The group names.
 * @returns The graph.
 */
function graphOf (src, groups) {
  const key = `${src}\0${groups.join(',')}`
  let graph = graphs.get(key)
  if (graph !== undefined) return graph

  graph = new Map()
  const walk = (directory) => {
    let entries
    try {
      entries = readdirSync(directory, { withFileTypes: true })
    } catch {
      return
    }
    if (isSlice(src, directory, groups)) {
      const edges = new Set()
      for (const entry of entries) {
        if (!entry.isFile() || !SOURCE.test(entry.name) || TEST.test(entry.name)) continue
        const text = readFileSync(join(directory, entry.name), 'utf8')
        for (const [, specifier] of text.matchAll(RELATIVE_IMPORT)) {
          if (ASSET.test(specifier)) continue
          const { directory: target } = resolveTarget(directory, specifier)
          if (target !== directory && target.startsWith(src + sep) && isSlice(src, target, groups)) edges.add(sliceName(src, target))
        }
      }
      graph.set(sliceName(src, directory), edges)
    }
    for (const entry of entries) if (entry.isDirectory() && !SUPPORT.has(entry.name)) walk(join(directory, entry.name))
  }
  walk(src)
  graphs.set(key, graph)

  return graph
}

/**
 * A path back from one slice to another through the graph.
 *
 * @param graph - A project's slice graph.
 * @param from - Where to start.
 * @param to - Where to arrive.
 * @param seen - Slices already walked.
 * @returns The slices on the way, or `null` when there is no way.
 */
function pathBetween (graph, from, to, seen = new Set()) {
  if (from === to) return [to]
  if (seen.has(from)) return null
  seen.add(from)
  const nexts = graph.get(from) ?? []
  for (const next of nexts) {
    const rest = pathBetween(graph, next, to, seen)
    if (rest !== null) return [from, ...rest]
  }

  return null
}

/**
 * The outermost slice, from `src` down to a directory, that the importer does not itself live in:
 * the one whose barrel is the public door to the target.
 *
 * @param src - A project's `src` directory.
 * @param directory - The target directory.
 * @param owner - The importer's own directory.
 * @param groups - The group names.
 * @returns The slice directory, or `null` when the target is in no such slice.
 */
function outermostSlice (src, directory, owner, groups) {
  let current = src
  for (const part of relative(src, directory).split(sep)) {
    if (part === '') continue
    current = join(current, part)
    if (owner === current || owner.startsWith(current + sep)) continue
    if (isSlice(src, current, groups)) return current
  }

  return null
}

const SOURCES = {
  ImportDeclaration:      node => node.source,
  ExportNamedDeclaration: node => node.source,
  ExportAllDeclaration:   node => node.source,
  ImportExpression:       node => (node.source.type === 'Literal' ? node.source : null),
}

/**
 * Visitors that call `check` with every module specifier in the file.
 *
 * @param check - Called with the specifier and its node.
 * @returns ESLint visitors.
 */
function onSpecifiers (check) {
  return Object.fromEntries(Object.entries(SOURCES).map(([type, pick]) => [type, (node) => {
    const source = pick(node)
    if (source !== null && source !== undefined && typeof source.value === 'string') check(source.value, source)
  }]))
}

/** The rules, as a plugin. */
export const plugin = {
  meta:  { name: 'monosmith-vertical-slices' },
  rules: {
    'file-role': {
      meta: {
        type:     'problem',
        messages: {
          case:   'Name files and folders in kebab-case: "{{name}}".',
          role:   'A production file ends in its role - .{{roles}} - so "{{name}}" says what it is. If this is a front-end file (a .ts file under a UI project), the front-end roles .route, .component, .section, .hook, .style, .content, .mock and .fixture exist too: allow them with the verticalSlices.roles option.',
          banned: '"{{name}}" has the word "{{word}}" in its name, and that names no outcome. Name the file for what it does.',
          root:   'Only index and main live at the root of src; "{{name}}" belongs in a subfeature.',
          group:  '"{{group}}" is a group: it holds slice folders, and at most an index barrel. "{{name}}" belongs in a slice inside it.',
        },
        schema: [
          {
            type:                 'object',
            properties:           { roles: { type: 'array', items: { type: 'string' } }, groups: { type: 'array', items: { type: 'string' } } },
            additionalProperties: false,
          },
        ],
      },
      create (context) {
        const place = locate(context.filename)
        if (place === null) return {}
        const name = basename(context.filename)
        // Extra roles are APPENDED to the defaults rather than replacing them:
        // a workspace adding `.component.tsx` still wants `.use-case.ts`, and
        // making the option replace the list would mean restating fourteen
        // entries to add one.
        // A file that imports Angular is UI by construction, as one that holds JSX is: its components and routes
        // are classes in `.ts` files, so without this every Angular app would need to opt in.
        const angular = /\bfrom\s+['"]@angular\//.test(context.sourceCode.text)
        const roles = [...ROLES, ...(angular ? TSX_ROLES : []), ...(context.options[0]?.roles ?? [])]
        const production = productionPattern(roles)
        // Unlike `roles`, `groups` REPLACES the default list: it names this workspace's folders, not a vocabulary to extend.
        const groups = context.options[0]?.groups ?? GROUPS

        return {
          Program (node) {
            const folders = place.dirs
            for (const folder of folders) if (!SUPPORT.has(folder) && !KEBAB.test(folder)) context.report({ node, messageId: 'case', data: { name: folder } })
            if (!ENTRY.test(name) && !KEBAB.test(name.split('.', 1)[0])) context.report({ node, messageId: 'case', data: { name } })
            // A name that already lacks its role gets that message alone; the banned word comes once the name is otherwise right.
            const banned = () => {
              if (ENTRY.test(name) || folders.some(folder => SUPPORT.has(folder))) return
              const word = name.toLowerCase().split(/[-._]/).find(part => BANNED_WORDS.includes(part))
              if (word !== undefined) context.report({ node, messageId: 'banned', data: { name, word } })
            }
            if (TEST.test(name)) return banned()
            // Mocks, fixtures and test data take any role: they are support, not production.
            if (folders.includes(TESTING)) return
            if (folders.length === 0) {
              if (!ENTRY.test(name)) context.report({ node, messageId: 'root', data: { name } })
            } else if (folders.length === 1 && groups.includes(folders[0])) {
              if (!BARREL.test(name)) context.report({ node, messageId: 'group', data: { name, group: folders[0] } })
            } else if (!BARREL.test(name) && !production.test(name)) {
              context.report({ node, messageId: 'role', data: { name, roles: roles.join(' .') } })

              return
            }

            banned()
          },
        }
      },
    },

    'no-deep-import': {
      meta: {
        type:     'problem',
        messages: {
          deep: 'Reach "{{slice}}" through its index - import from \'{{door}}\' (or \'{{door}}/index.ts\' under nodenext), not from a file inside it or a slice nested in it.',
          self: 'Import the files of your own subfeature directly, never through its own index.',
        },
        schema: [{ type: 'object', properties: { groups: { type: 'array', items: { type: 'string' } } }, additionalProperties: false }],
      },
      create (context) {
        const place = locate(context.filename)
        if (place === null || place.dirs.length === 0) return {}
        const groups = context.options[0]?.groups ?? GROUPS
        const own = dirname(context.filename)
        const owner = ownerOf(own)

        return onSpecifiers((specifier, node) => {
          // Static assets (an svg, a stylesheet) are not slice code, so they have no barrel to go through.
          if (!specifier.startsWith('.') || ASSET.test(specifier)) return
          const { directory, file } = resolveTarget(own, specifier)
          // The root of `src` holds entries, not a slice: the file-role rule already reports what does not belong there.
          if (directory === place.src || (directory !== own && !directory.startsWith(place.src + sep))) return
          if (file === null && directory === own) return void context.report({ node, messageId: 'self' })
          // A file is reachable only from its own folder (a support folder belongs to the folder above it).
          const home = ownerOf(directory)
          // Files of `src/testing` (shared mocks and test setup) are support, free to import.
          if (home === place.src || (file !== null && home === owner)) return
          const slice = outermostSlice(place.src, home, owner, groups)
          if (file === null && (slice === null || slice === directory)) return
          const door = relative(own, slice ?? home).split(sep).join('/')
          context.report({ node, messageId: 'deep', data: { slice: basename(slice ?? home), door: door.startsWith('.') ? door : `./${door}` } })
        })
      },
    },

    'no-slice-cycle': {
      meta: {
        type:     'problem',
        messages: {
          cycle: 'Subfeatures import each other: {{path}}. Move what both need into a slice they can both depend on.',
        },
        schema: [{ type: 'object', properties: { groups: { type: 'array', items: { type: 'string' } } }, additionalProperties: false }],
      },
      create (context) {
        const place = locate(context.filename)
        if (place === null || TEST.test(context.filename)) return {}
        const groups = context.options[0]?.groups ?? GROUPS
        const own = dirname(context.filename)
        if (place.dirs.some(folder => SUPPORT.has(folder)) || !isSlice(place.src, own, groups)) return {}
        const here = sliceName(place.src, own)

        return onSpecifiers((specifier, node) => {
          if (!specifier.startsWith('.') || ASSET.test(specifier)) return
          const { directory } = resolveTarget(own, specifier)
          if (directory === own || !directory.startsWith(place.src + sep) || !isSlice(place.src, directory, groups)) return
          const back = pathBetween(graphOf(place.src, groups), sliceName(place.src, directory), here)
          if (back !== null) context.report({ node, messageId: 'cycle', data: { path: [here, ...back].join(' -> ') } })
        })
      },
    },
  },
}

/**
 * Vertical feature slices, enforced: a project's `src` holds subfeatures, each
 * reached only through its `index`, holding flat, kebab-case, role-suffixed
 * files - and no two subfeatures importing each other.
 *
 * @remarks
 * **On by default** (smith's architecture is vertical feature slices), for every
 * project's `src` under `apps/`, `libs/` and `packages/`. A workspace that is
 * not organised this way opts out with `monosmith({ verticalSlices: false })`, or
 * passes the globs of the projects that follow it.
 *
 * Three rules:
 *
 * - `vertical-slices/file-role` - kebab-case names; a production file ends in
 *   its role (`.use-case.ts`, `.policy.ts`, `.algorithm.ts`... see `ROLES`); a
 *   group folder holds only slices and an optional `index`; only `index` and
 *   `main` sit at the root of `src`.
 * - `vertical-slices/no-deep-import` - another slice is reached through
 *   its `index` (`'../billing'`, or `'../billing/index.ts'` where Node's ESM
 *   resolution needs a file), never a file inside it (`'../billing/fee.policy'`)
 *   or a slice nested in it, and a slice never imports its own `index`.
 * - `vertical-slices/no-slice-cycle` - no two slices importing each other,
 *   **type-only imports included**. This is the gap `import-x/no-cycle` leaves:
 *   it sees cycles between *files*, and a cycle between two folders usually runs
 *   through different files on each side - a contract in one, a use case in the
 *   other - so no file-level cycle exists to report. Type imports count because
 *   they tie the slices together as surely as values do, even though they are
 *   erased at run time.
 *
 * Tests (`.spec` or `.test`) are exempt from TWO of the three rules, and the
 * split is deliberate rather than an oversight:
 *
 * - `file-role` exempts them, because a test is named for the file it tests
 *   and so has no role suffix by design.
 * - `no-slice-cycle` exempts them, because it describes the PRODUCTION
 *   dependency graph and a spec is not in the shipped bundle.
 * - `no-deep-import` does NOT. It is about respecting a sibling's public API,
 *   and a test reaching past an index couples to that sibling's internals
 *   exactly as production code would: rename a file in one slice and another
 *   slice's test breaks. The rule does not care who wrote the import.
 *
 * This used to be documented as a blanket "tests may reach wherever they need
 * to", which was simply false - a test importing `'../template/x.algorithm'`
 * has always been reported.
 *
 * THE ROLE LIST IS EXTENSIBLE
 *
 * {@link ROLES} is the back-end vocabulary, and it is not the only one. The ADR
 * behind these rules also allows `.service` (with a recorded exception), and
 * its front-end amendment adds `.route`, `.component`,
 * `.hook`, `.section`, `.style`, `.content`, `.mock` and `.fixture`. A React
 * app could not opt in at all while the list was fixed - every component it
 * has would report.
 *
 * So `roles` appends to the defaults rather than replacing them: a workspace
 * adding `.component.tsx` still wants `.use-case.ts`, and a replacing option
 * would mean restating fourteen entries to add one.
 *
 * GROUPS AND NESTING
 *
 * A folder directly under `src` named in `groups` (default `features`,
 * `integrations`, `shared`) is a GROUP, not a slice: its children are the
 * slices, and it holds no production file except an optional `index` barrel.
 * Unlike `roles`, the `groups` option REPLACES the default list, because it
 * names this workspace's own folders. A slice directly under `src` works as
 * before.
 *
 * A slice may hold subfeature folders, to any depth. Every folder that holds
 * production files is a slice: flat, role-suffixed files and its own `index`
 * as the only door. A folder that holds only a barrel and child slices is
 * valid. `fixtures` and `testing` folders are support, never slices, and
 * `testing` is exempt from the role rule.
 *
 * A file may import: its own folder's files relatively; a child slice through
 * its barrel (`'./child'`); any other slice through its barrel (`'../other'`,
 * `'../../features/other'`). Reaching a slice's files, or a slice nested in
 * another slice, is a deep import. Importing the own barrel is reported. The
 * slice graph has one node per slice folder, an edge per import of its barrel,
 * so a parent and its child importing each other is a cycle.
 *
 * @param options - Globs, or `{ files, roles, groups }`. An array (or nothing) is
 * treated as `files`, so both earlier forms keep working.
 * @param options.files - Globs of the source this applies to. Default: every
 * project's `src` under `apps/`, `libs/` and `packages/`.
 * @param options.roles - Extra role suffixes, APPENDED to {@link ROLES} (and {@link TSX_ROLES}).
 * @param options.groups - Group folder names under `src`, REPLACING {@link GROUPS}.
 * @returns The flat config blocks.
 */
export default function verticalSlices (options = {}) {
  // `true` never reaches here - index.js maps it to `undefined` - but an array
  // does, and so does nothing at all. Both predate the options object and both
  // still work.
  const settings = Array.isArray(options) ? { files: options } : options
  const files = settings.files ?? [
    'apps/*/src/**/*.{ts,mts,cts,tsx}',
    'libs/*/src/**/*.{ts,mts,cts,tsx}',
    'packages/*/src/**/*.{ts,mts,cts,tsx}',
  ]
  const roles = settings.roles ?? []
  const grouped = settings.groups === undefined ? {} : { groups: settings.groups }
  const rolesOption = roles.length > 0 ? { roles, ...grouped } : grouped

  return [
    {
      name:    'monosmith/vertical-slices',
      files,
      plugins: { 'vertical-slices': plugin },
      rules:   {
        // Only `file-role` takes the roles: the other two are about imports,
        // which a file's suffix has no bearing on. All three take the groups.
        'vertical-slices/file-role':      Object.keys(rolesOption).length > 0 ? ['error', rolesOption] : 'error',
        'vertical-slices/no-deep-import': Object.keys(grouped).length > 0 ? ['error', grouped] : 'error',
        'vertical-slices/no-slice-cycle': Object.keys(grouped).length > 0 ? ['error', grouped] : 'error',
      },
    },
    {
      // JSX files take the front-end roles with no option (see TSX_ROLES). Nested
      // `files` arrays are an AND, so this narrows the user's globs to `.tsx`.
      name:    'monosmith/vertical-slices-tsx',
      files:   files.map(glob => [glob, '**/*.tsx']),
      plugins: { 'vertical-slices': plugin },
      rules:   { 'vertical-slices/file-role': ['error', { roles: [...TSX_ROLES, ...roles], ...grouped }] },
    },
  ]
}
