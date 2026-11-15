import { spawnSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import { pathToFileURL } from 'node:url'

/**
 * The opt-in vertical-slice rules, run through the real `eslint` binary for the
 * same reasons as `config.spec.ts`: a flat config's shape proves nothing, and
 * this package is ESM while Jest runs these specs as CJS.
 *
 * One workspace, organised as slices, with each violation planted once and one
 * slice that must lint clean. The rules are opt-in, so `config.spec.ts` - whose
 * fixtures sit flat at the root of `src` - is the proof that nothing changes
 * for a workspace that does not ask for them.
 */

const packageRoot = join(__dirname, '..')
const eslintBin = join(packageRoot, '..', '..', 'node_modules', '.bin', 'eslint')

const FIXTURES: Record<string, string> = {
  // A self-contained tsconfig at the WORKSPACE ROOT, so the TypeScript resolver
  // that `import-x/no-cycle` uses stops walking here.
  //
  // Without it the walk continues into the OS temp directory, which these
  // workspaces are created inside. Anything a previous tool left there - a
  // generated workspace's own tsconfig.json, say - is picked up instead, and if
  // it `extends` a file that is not there the resolver throws and EVERY
  // assertion in the suite fails with "eslint produced no JSON". CI never sees
  // it because CI's temp directory is empty; a developer's is not.
  'tsconfig.json':
    '{\n  "compilerOptions": {\n    "target": "es2021",\n    "module": "commonjs",\n    "moduleResolution": "node",\n    "strict": true,\n    "noEmit": true\n  },\n  "include": []\n}\n',
  'packages/app/tsconfig.json':
    '{\n  "compilerOptions": {\n    "target": "es2021",\n    "module": "commonjs",\n    "moduleResolution": "node",\n    "strict": true,\n    "noEmit": true\n  },\n  "include": ["src/**/*.ts", "src/**/*.tsx"]\n}\n',
  // A project whose slices sit in groups: every slice is in one, and a slice directly under src would be reported.
  'packages/grouped/tsconfig.json':
    '{\n  "compilerOptions": {\n    "target": "es2021",\n    "module": "commonjs",\n    "moduleResolution": "node",\n    "strict": true,\n    "noEmit": true\n  },\n  "include": ["src/**/*.ts"]\n}\n',
  'packages/app/src/index.ts':                    "export { fee } from './billing'\nexport { add } from './arithmetic'\n",
  // A clean slice: flat, role-suffixed, reached through its index.
  'packages/app/src/arithmetic/index.ts':         "export { add } from './add.algorithm'\n",
  'packages/app/src/arithmetic/add.algorithm.ts': 'export function add (a: number, b: number): number {\n  return a + b\n}\n',
  // billing and invoicing import each other - through different files on each
  // side, a type on one and a value on the other, so no FILE forms a cycle.
  'packages/app/src/billing/index.ts':            "export { fee } from './fee.policy'\n",
  'packages/app/src/billing/fee.policy.ts':
    "import type { Invoice } from '../invoicing'\n\nexport function fee (invoice: Invoice): number {\n  return invoice.amount / 100\n}\n",
  'packages/app/src/invoicing/index.ts':            "export type { Invoice } from './invoice.contract'\nexport { total } from './total.use-case'\n",
  'packages/app/src/invoicing/invoice.contract.ts': 'export interface Invoice {\n  amount: number\n}\n',
  'packages/app/src/invoicing/total.use-case.ts':
    "import { fee } from '../billing'\nimport type { Invoice } from './invoice.contract'\n\nexport function total (invoice: Invoice): number {\n  return invoice.amount + fee(invoice)\n}\n",
  // A JSX file takes the front-end roles with no option; the same name on a `.ts` file does not.
  'packages/app/src/ui/index.ts':                  "export { Greeting } from './greeting.component'\n",
  'packages/app/src/ui/greeting.component.tsx':    "export function Greeting (): string {\n  return 'hi'\n}\n",
  'packages/app/src/ui/farewell.component.ts':     "export function Farewell (): string {\n  return 'bye'\n}\n",
  'packages/app/src/ui/untyped.tsx':               'export const Untyped = 1\n',
  // A Chakra recipe needs no option; stories are a JSX role, also in a nested stories/ slice reaching its parent through '..'.
  'packages/app/src/ui/button.recipe.ts':          'export const buttonRecipe = { base: {} }\n',
  'packages/app/src/ui/greeting.stories.tsx':      "import { Greeting } from './greeting.component'\n\nexport default { component: Greeting }\n",
  'packages/app/src/ui/stories/forms.stories.tsx': "import { Greeting } from '..'\n\nexport default { component: Greeting }\n",
  'packages/app/src/reports/index.ts':             "export { report } from './report.use-case'\n",
  'packages/app/src/reports/report.use-case.ts':
    "import { add } from '../arithmetic/add.algorithm'\n\nexport function report (): number {\n  return add(1, 2)\n}\n",
  // Under nodenext a slice's index is named as a file: the same door.
  'packages/app/src/reports/total.use-case.ts':
    "import { add } from '../arithmetic/index.ts'\n\nexport function total (): number {\n  return add(1, 2)\n}\n",
  'packages/app/src/reports/emitted.use-case.ts':
    "import { add } from '../arithmetic/index.js'\n\nexport function emitted (): number {\n  return add(1, 2)\n}\n",
  'packages/app/src/reports/own-index.use-case.ts':
    "import { report } from './index.ts'\n\nexport function again (): number {\n  return report()\n}\n",
  'packages/app/src/reports/selfish.use-case.ts':
    "import { report } from '.'\n\nexport function twice (): number {\n  return report() * 2\n}\n",
  'packages/app/src/reports/with-auth.middleware.ts':     'export const withAuth = 1\n',
  'packages/app/src/reports/reportHelper.ts':             'export const helper = 1\n',
  'packages/app/src/reports/deeper/inner.use-case.ts':    'export const inner = 1\n',
  // A banned word in a file name, in a production file and in a test, and one inside a fixtures folder.
  'packages/app/src/reports/date-utils.algorithm.ts':     'export const dateUtils = 1\n',
  'packages/app/src/reports/shared-state.store.spec.ts':  'export const sharedState = 1\n',
  'packages/app/src/reports/fixtures/helper.contract.ts': 'export const fixture = 1\n',
  'packages/app/src/stray.ts':                            'export const stray = 1\n',
  // A folder nested in a slice is a slice of its own now, and `testing` is support.
  'packages/app/src/reports/testing/quote-data.ts':       'export const quote = 1\n',
  'packages/app/src/reports/testing/camelCase.ts':        'export const camel = 1\n',
  // GROUPS: `features` and `integrations` hold slices, nested to any depth.
  'packages/grouped/src/features/index.ts':               "export { run } from './work-items'\nexport { ledger } from './ledger'\n",
  'packages/grouped/src/features/stray.use-case.ts':      'export const stray = 1\n',
  'packages/grouped/src/features/work-items/index.ts':    "export { clean } from './application/cleaner'\nexport { prune } from './cleaner'\n",
  'packages/grouped/src/features/work-items/run.use-case.ts':
    "import { clean } from './application/cleaner'\n\nexport function run (): number {\n  return clean()\n}\n",
  'packages/grouped/src/features/work-items/reach.use-case.ts':
    "import { clean } from './application/cleaner/cleaner.use-case'\n\nexport function reach (): number {\n  return clean()\n}\n",
  'packages/grouped/src/features/work-items/cleaner/index.ts': "export { prune } from './cleaner.use-case'\n",
  'packages/grouped/src/features/work-items/cleaner/cleaner.use-case.ts':
    "import { pull } from '../../../integrations/git'\n\nexport function prune (): number {\n  return pull()\n}\n",
  'packages/grouped/src/features/work-items/application/cleaner/index.ts': "export { clean } from './cleaner.use-case'\n",
  'packages/grouped/src/features/work-items/application/cleaner/cleaner.use-case.ts':
    "import { prune } from '../../cleaner'\nimport { pull } from '../../../../integrations/git'\n\nexport function clean (): number {\n  return prune() + pull()\n}\n",
  'packages/grouped/src/features/work-items/application/cleaner/deep.use-case.ts':
    "import { push } from '../../../../integrations/git/git.client'\n\nexport function deep (): number {\n  return push()\n}\n",
  'packages/grouped/src/features/work-items/application/cleaner/nested.use-case.ts':
    "import { prune } from '../../../work-items/cleaner/cleaner.use-case'\n\nexport function nested (): number {\n  return prune()\n}\n",
  'packages/grouped/src/integrations/git/index.ts':      "export { pull } from './git.client'\n",
  'packages/grouped/src/integrations/git/git.client.ts': 'export const pull = (): number => 1\nexport const push = (): number => 2\n',
  // A parent and its child importing each other.
  'packages/grouped/src/features/ledger/index.ts':       "export { ledger } from './ledger.policy'\n",
  'packages/grouped/src/features/ledger/ledger.policy.ts':
    "import { entry } from './entries'\n\nexport function ledger (): number {\n  return entry()\n}\n",
  'packages/grouped/src/features/ledger/entries/index.ts': "export { entry } from './entry.policy'\n",
  'packages/grouped/src/features/ledger/entries/entry.policy.ts':
    "import type { ledger } from '..'\n\nexport type Ledger = typeof ledger\nexport function entry (): number {\n  return 1\n}\n",
  // Tests name the file they test, and are exempt from the role and cycle rules.
  'packages/app/src/arithmetic/add.algorithm.spec.ts':
    "import { add } from './add.algorithm'\n\ndescribe('add', () => {\n  it('adds', () => {\n    expect(add(1, 2)).toBe(3)\n  })\n})\n",
  // A test reaching PAST a sibling's index. NOT exempt, deliberately — see the
  // assertion below for why the three rules treat tests differently.
  'packages/app/src/billing/fee.policy.spec.ts':
    "import { add } from '../arithmetic/add.algorithm'\n\ndescribe('fee', () => {\n  it('adds', () => {\n    expect(add(1, 2)).toBe(3)\n  })\n})\n",
}

let workspace: string
let reported: Record<string, string[]>

/** The messages of the last run, by file, beside the rule ids `lintAll` returns. */
const texts: Record<string, string[]> = {}

function lintAll (directory: string): Record<string, string[]> {
  const result = spawnSync(eslintBin, ['.', '--format', 'json', '--no-error-on-unmatched-pattern'], {
    cwd:      directory,
    encoding: 'utf8',
    shell:    process.platform === 'win32',
  })
  const stdout = result.stdout?.trim()
  if (!stdout?.startsWith('[')) throw new Error(`eslint produced no JSON.\nstdout: ${result.stdout}\nstderr: ${result.stderr}`)
  const parsed = JSON.parse(stdout) as { filePath: string, messages: { ruleId: string | null, message: string }[] }[]
  const byFile: Record<string, string[]> = {}
  for (const file of parsed) {
    const relative = file.filePath.slice(directory.length + 1).replaceAll('\\', '/')
    byFile[relative] = file.messages.map(message => message.ruleId ?? 'FATAL')
    texts[relative] = file.messages.map(message => message.message)
  }

  return byFile
}

/** The vertical-slice rules reported for one file. */
function slicesFor (filename: string): string[] {
  return (reported[filename] ?? []).filter(rule => rule.startsWith('vertical-slices/') || rule === 'FATAL').toSorted((a, b) => a.localeCompare(b))
}

beforeAll(() => {
  workspace = mkdtempSync(join(tmpdir(), 'monosmith-eslint-slices-'))
  const entry = pathToFileURL(join(packageRoot, 'index.js')).href
  writeFileSync(join(workspace, 'eslint.config.mjs'), `import monosmith from ${JSON.stringify(entry)}\nexport default monosmith({ verticalSlices: true })\n`)
  for (const [filename, contents] of Object.entries(FIXTURES)) {
    const target = join(workspace, filename)
    mkdirSync(dirname(target), { recursive: true })
    writeFileSync(target, contents)
  }
  reported = lintAll(workspace)
})

afterAll(() => {
  rmSync(workspace, { recursive: true, force: true })
})

describe('monosmith({ verticalSlices: { roles } })', () => {
  // The default role list is the BACK-END vocabulary. The ADR these rules come
  // from also allows `.service` and `.middleware`, and its front-end amendment
  // adds `.route`, `.component`, `.hook`, `.section`, `.style`, `.content`,
  // `.mock` and `.fixture`. While the list was fixed, a React app could not opt
  // in at all - every component it has would report.
  let extended: Record<string, string[]>
  let scoped: string

  beforeAll(() => {
    scoped = mkdtempSync(join(tmpdir(), 'monosmith-eslint-slices-roles-'))
    const entry = pathToFileURL(join(packageRoot, 'index.js')).href
    writeFileSync(
      join(scoped, 'eslint.config.mjs'),
      `import monosmith from ${JSON.stringify(entry)}\nexport default monosmith({ verticalSlices: { roles: ['component', 'hook'] } })\n`,
    )
    const files: Record<string, string> = {
      'tsconfig.json':                                FIXTURES['tsconfig.json'],
      'packages/app/tsconfig.json':                   FIXTURES['packages/app/tsconfig.json'],
      'packages/app/src/index.ts':                    "export { Button } from './buttons'\n",
      'packages/app/src/buttons/index.ts':            "export { Button } from './button.component'\n",
      // An added role.
      'packages/app/src/buttons/button.component.ts': 'export const Button = 1\n',
      // Another added role.
      'packages/app/src/buttons/use-press.hook.ts':   'export const usePress = 1\n',
      // A DEFAULT role, which must still pass - the option appends, it does not
      // replace, or adding one role would cost you the other fourteen.
      'packages/app/src/buttons/press.use-case.ts':   'export const press = 1\n',
      // Still not a role.
      'packages/app/src/buttons/helper.ts':           'export const helper = 1\n',
    }
    for (const [filename, contents] of Object.entries(files)) {
      const target = join(scoped, filename)
      mkdirSync(dirname(target), { recursive: true })
      writeFileSync(target, contents)
    }
    extended = lintAll(scoped)
  })

  afterAll(() => {
    rmSync(scoped, { recursive: true, force: true })
  })

  const slicesIn = (filename: string): string[] =>
    (extended[filename] ?? [])
      .filter(rule => rule.startsWith('vertical-slices/') || rule === 'FATAL')
      .toSorted((a, b) => a.localeCompare(b))

  it('accepts the roles it was given', () => {
    expect(slicesIn('packages/app/src/buttons/button.component.ts')).toEqual([])
    expect(slicesIn('packages/app/src/buttons/use-press.hook.ts')).toEqual([])
  })

  it('still accepts the defaults, because the option APPENDS', () => {
    expect(slicesIn('packages/app/src/buttons/press.use-case.ts')).toEqual([])
  })

  it('still reports a file with no role at all', () => {
    expect(slicesIn('packages/app/src/buttons/helper.ts')).toEqual(['vertical-slices/file-role'])
  })

  it('tells a file with no role that the front-end roles exist and the roles option opts in', () => {
    const [message] = texts['packages/app/src/buttons/helper.ts'] ?? []
    expect(message).toContain('.use-case')
    expect(message).toContain('front-end')
    expect(message).toContain('.hook')
    expect(message).toContain('.content')
    expect(message).toContain('verticalSlices.roles')
  })
})

describe('monosmith({ verticalSlices })', () => {
  it('passes a slice that follows the rules, and its test', () => {
    expect(slicesFor('packages/app/src/index.ts')).toEqual([])
    expect(slicesFor('packages/app/src/reports/with-auth.middleware.ts')).toEqual([])
    expect(slicesFor('packages/app/src/arithmetic/index.ts')).toEqual([])
    expect(slicesFor('packages/app/src/arithmetic/add.algorithm.ts')).toEqual([])
    expect(slicesFor('packages/app/src/arithmetic/add.algorithm.spec.ts')).toEqual([])
  })

  it('reports two subfeatures importing each other, though no file forms a cycle', () => {
    expect(slicesFor('packages/app/src/billing/fee.policy.ts')).toEqual(['vertical-slices/no-slice-cycle'])
    expect(slicesFor('packages/app/src/invoicing/total.use-case.ts')).toEqual(['vertical-slices/no-slice-cycle'])
  })

  it('reports a sibling reached past its index, and a slice importing its own', () => {
    expect(slicesFor('packages/app/src/reports/report.use-case.ts')).toEqual(['vertical-slices/no-deep-import'])
    expect(slicesFor('packages/app/src/reports/selfish.use-case.ts')).toEqual(['vertical-slices/no-deep-import'])
    expect(slicesFor('packages/app/src/reports/own-index.use-case.ts')).toEqual(['vertical-slices/no-deep-import'])
  })

  it('accepts the index of a sibling named as a file, as nodenext needs', () => {
    expect(slicesFor('packages/app/src/reports/total.use-case.ts')).toEqual([])
    expect(slicesFor('packages/app/src/reports/emitted.use-case.ts')).toEqual([])
  })

  it('holds a TEST to no-deep-import, unlike the role and cycle rules', () => {
    // The three rules treat tests differently, and the distinction is
    // principled rather than an oversight - it was documented as a blanket
    // "tests may reach wherever they need to", which was simply wrong.
    //
    // `file-role` must exempt them: a test is named for the file it tests, so
    // it has no role suffix by design.
    // `no-slice-cycle` must exempt them: it describes the PRODUCTION dependency
    // graph, and a spec is not in the shipped bundle.
    // `no-deep-import` must NOT: it is about respecting a sibling's public
    // API, and a test that reaches past an index couples to that sibling's
    // internals exactly as production code would. Rename a file in one slice
    // and another slice's test breaks - which is the coupling the rule exists
    // to prevent, and it does not care who wrote the import.
    expect(slicesFor('packages/app/src/billing/fee.policy.spec.ts'))
      .toEqual(['vertical-slices/no-deep-import'])
  })

  it('accepts the front-end roles on a .tsx file with no option, and only there', () => {
    expect(slicesFor('packages/app/src/ui/greeting.component.tsx')).toEqual([])
    expect(slicesFor('packages/app/src/ui/farewell.component.ts')).toEqual(['vertical-slices/file-role'])
    expect(slicesFor('packages/app/src/ui/untyped.tsx')).toEqual(['vertical-slices/file-role'])
  })

  it('accepts a recipe on any file, and stories on a .tsx file, in a nested stories/ slice too', () => {
    expect(slicesFor('packages/app/src/ui/button.recipe.ts')).toEqual([])
    expect(slicesFor('packages/app/src/ui/greeting.stories.tsx')).toEqual([])
    expect(slicesFor('packages/app/src/ui/stories/forms.stories.tsx')).toEqual([])
  })

  it('accepts the front-end roles on a .ts file that imports Angular', () => {
    expect(slicesFor('packages/app/src/ui/toolbar.component.ts')).toEqual([])
  })

  it('reports a file without its role, or not in kebab-case, nested, or at the root of src', () => {
    expect(slicesFor('packages/app/src/reports/reportHelper.ts')).toEqual(['vertical-slices/file-role', 'vertical-slices/file-role'])
    // A folder inside a slice used to be reported as "nesting"; it is a slice of its own now.
    expect(slicesFor('packages/app/src/reports/deeper/inner.use-case.ts')).toEqual([])
    expect(slicesFor('packages/app/src/stray.ts')).toEqual(['vertical-slices/file-role'])
    expect(slicesFor('packages/app/src/content.config.ts')).toEqual([])
  })

  it('accepts a group with a feature and an integration, nested to any depth, and a group barrel', () => {
    for (const file of [
      'features/index.ts',
      'features/work-items/index.ts',
      'features/work-items/run.use-case.ts',
      'features/work-items/cleaner/cleaner.use-case.ts',
      'features/work-items/application/cleaner/index.ts',
      'features/work-items/application/cleaner/cleaner.use-case.ts',
      'integrations/git/index.ts',
      'integrations/git/git.client.ts',
    ]) expect(slicesFor(`packages/grouped/src/${file}`)).toEqual([])
  })

  it('reports a production file directly in a group folder', () => {
    expect(slicesFor('packages/grouped/src/features/stray.use-case.ts')).toEqual(['vertical-slices/file-role'])
  })

  it('reports a deep import into a nested slice, a child slice file and a grandchild slice', () => {
    expect(slicesFor('packages/grouped/src/features/work-items/application/cleaner/deep.use-case.ts')).toEqual(['vertical-slices/no-deep-import'])
    expect(slicesFor('packages/grouped/src/features/work-items/reach.use-case.ts')).toEqual(['vertical-slices/no-deep-import'])
    expect(slicesFor('packages/grouped/src/features/work-items/application/cleaner/nested.use-case.ts')).toEqual(['vertical-slices/no-deep-import'])
  })

  it('reports a parent and a child slice importing each other', () => {
    expect(slicesFor('packages/grouped/src/features/ledger/ledger.policy.ts')).toEqual(['vertical-slices/no-slice-cycle'])
    expect(slicesFor('packages/grouped/src/features/ledger/entries/entry.policy.ts')).toEqual(['vertical-slices/no-slice-cycle'])
    expect(texts['packages/grouped/src/features/ledger/ledger.policy.ts']?.[0]).toContain('features/ledger -> features/ledger/entries -> features/ledger')
  })

  it('exempts a testing folder from the role rule but not from kebab-case', () => {
    expect(slicesFor('packages/app/src/reports/testing/quote-data.ts')).toEqual([])
    expect(slicesFor('packages/app/src/reports/testing/camelCase.ts')).toEqual(['vertical-slices/file-role'])
  })

  it('reports a banned word in a file name, in production files and tests, but not in a fixtures folder', () => {
    expect(slicesFor('packages/app/src/reports/date-utils.algorithm.ts')).toEqual(['vertical-slices/file-role'])
    expect(slicesFor('packages/app/src/reports/shared-state.store.spec.ts')).toEqual(['vertical-slices/file-role'])
    expect(slicesFor('packages/app/src/reports/fixtures/helper.contract.ts')).toEqual([])
  })
})

describe('monosmith({ verticalSlices: { groups } })', () => {
  // The option REPLACES the default list: `features` is an ordinary slice here, `domains` is the group.
  let custom: Record<string, string[]>
  let directory: string

  beforeAll(() => {
    directory = mkdtempSync(join(tmpdir(), 'monosmith-eslint-slices-groups-'))
    const entry = pathToFileURL(join(packageRoot, 'index.js')).href
    writeFileSync(
      join(directory, 'eslint.config.mjs'),
      `import monosmith from ${JSON.stringify(entry)}\nexport default monosmith({ verticalSlices: { groups: ['domains'] } })\n`,
    )
    const files: Record<string, string> = {
      'tsconfig.json':                                  FIXTURES['tsconfig.json'],
      'packages/app/tsconfig.json':                     FIXTURES['packages/app/tsconfig.json'],
      'packages/app/src/domains/stray.use-case.ts':     'export const stray = 1\n',
      'packages/app/src/domains/billing/index.ts':      "export { fee } from './fee.policy'\n",
      'packages/app/src/domains/billing/fee.policy.ts': 'export const fee = 1\n',
      // A project without the named group: `features` is an ordinary slice there.
      'packages/plain/tsconfig.json':                   FIXTURES['packages/app/tsconfig.json'],
      'packages/plain/src/features/stray.use-case.ts':  'export const stray = 1\n',
    }
    for (const [filename, contents] of Object.entries(files)) {
      const target = join(directory, filename)
      mkdirSync(dirname(target), { recursive: true })
      writeFileSync(target, contents)
    }
    custom = lintAll(directory)
  })

  afterAll(() => {
    rmSync(directory, { recursive: true, force: true })
  })

  const slicesIn = (filename: string): string[] =>
    (custom[filename] ?? []).filter(rule => rule.startsWith('vertical-slices/') || rule === 'FATAL')

  it('treats the named folder as a group and the default names as plain slices', () => {
    expect(slicesIn('packages/app/src/domains/billing/fee.policy.ts')).toEqual([])
    expect(slicesIn('packages/app/src/domains/stray.use-case.ts')).toEqual(['vertical-slices/file-role'])
    expect(slicesIn('packages/plain/src/features/stray.use-case.ts')).toEqual([])
  })
})

describe('monosmith({ verticalSlices: { vendors } }) in an app with zones', () => {
  // An app keeps its slices in features/, integrations/ and shared/, reached by tsconfig aliases as well
  // as by relative paths. The zone direction, the vendor list and the role limits are one rule: zone-import.
  let zoned: Record<string, string[]>
  let directory: string

  beforeAll(() => {
    directory = mkdtempSync(join(tmpdir(), 'monosmith-eslint-slices-zones-'))
    const entry = pathToFileURL(join(packageRoot, 'index.js')).href
    writeFileSync(
      join(directory, 'eslint.config.mjs'),
      `import monosmith from ${JSON.stringify(entry)}\nexport default monosmith({ verticalSlices: { files: ['apps/*/src/**/*.ts', 'libs/*/src/**/*.ts'], vendors: ['pg', '@scanmate/*'] } })\n`,
    )
    const crmType = "import type { CrmTicket } from '@integrations/crm'\n"
    const files: Record<string, string> = {
      'tsconfig.json': FIXTURES['tsconfig.json'],
      'apps/api/tsconfig.json':
        '{\n  "compilerOptions": {\n    "strict": true,\n    "noEmit": true,\n    "paths": {\n      "@features/*": ["./src/features/*/index.ts"],\n      "@integrations/*": ["./src/integrations/*/index.ts"],\n      "@shared/*": ["./src/shared/*/index.ts"]\n    }\n  },\n  "include": ["src/**/*.ts"]\n}\n',
      'libs/money/tsconfig.json':                      FIXTURES['packages/app/tsconfig.json'],
      'apps/api/src/main.ts':                          "export { placeOrder } from '@features/orders'\n",
      'apps/api/src/integrations/crm/index.ts':        "export type { CrmTicket } from './crm.contract'\nexport { sendTicket } from './crm.client'\n",
      'apps/api/src/integrations/crm/crm.contract.ts': 'export interface CrmTicket {\n  id: string\n}\n',
      // An integration imports its vendor packages, and shared code.
      'apps/api/src/integrations/crm/crm.client.ts':
        "import { Pool } from 'pg'\nimport { scan } from '@scanmate/scan'\nimport { ok } from '@shared/http'\n\nexport const sendTicket = (): unknown => [Pool, scan, ok]\n",
      // An integration never imports a feature, by relative path either.
      'apps/api/src/integrations/crm/order-link.mapper.ts': "import { placeOrder } from '../../features/orders'\n\nexport const link = placeOrder\n",
      'apps/api/src/shared/http/index.ts':                  "export { ok } from './http.mapper'\n",
      'apps/api/src/shared/http/http.mapper.ts':            'export const ok = 200\n',
      // Shared code never imports a feature, and neither does its test.
      'apps/api/src/shared/http/status.mapper.ts':          "import { placeOrder } from '@features/orders'\n\nexport const status = placeOrder\n",
      'apps/api/src/shared/http/http.mapper.test.ts':       "import { placeOrder } from '@features/orders'\n\nexport const used = placeOrder\n",
      'apps/api/src/features/orders/index.ts':              "export { placeOrder } from './place-order.use-case'\n",
      // A use case calls an integration and its own repository.
      'apps/api/src/features/orders/place-order.use-case.ts':
        "import { sendTicket } from '@integrations/crm'\nimport { save } from './order.repository'\n\nexport const placeOrder = (): unknown => [sendTicket, save]\n",
      'apps/api/src/features/orders/order.repository.ts':      'export const save = 1\n',
      // A feature never imports a vendor package itself.
      'apps/api/src/features/orders/order-scan.client.ts':     "import { scan } from '@scanmate/scan'\n\nexport const check = scan\n",
      // Pure roles: no integration at all, not even its types, and no I/O role.
      'apps/api/src/features/orders/order.policy.ts':          `${crmType}\nexport const allowed = (ticket: CrmTicket): boolean => ticket.id !== ''\n`,
      'apps/api/src/features/orders/order.model.ts':           "import { save } from './order.repository'\n\nexport const model = save\n",
      // A mapper and a handler use an integration's types, never its code.
      'apps/api/src/features/orders/order-ticket.mapper.ts':   `${crmType}\nexport const toTicket = (id: string): CrmTicket => ({ id })\n`,
      'apps/api/src/features/orders/order-ticket.contract.ts': `${crmType}\nexport interface OrderTicket {
  ticket: CrmTicket
}\n`,
      'apps/api/src/features/orders/order-send.mapper.ts':   "import { sendTicket } from '@integrations/crm'\n\nexport const send = sendTicket\n",
      'apps/api/src/features/orders/place-order.handler.ts': "import { sendTicket } from '@integrations/crm'\n\nexport const handle = sendTicket\n",
      // A test may reach integrations and vendors.
      'apps/api/src/features/orders/order.policy.test.ts':
        "import { sendTicket } from '@integrations/crm'\nimport { Pool } from 'pg'\n\nexport const used = [sendTicket, Pool]\n",
      // `data` and `manager` are honest inside a compound, banned as the whole name.
      'apps/api/src/features/orders/user-data.contract.ts':     'export interface UserData {\n  id: string\n}\n',
      'apps/api/src/features/orders/data.contract.ts':          'export interface Data {\n  id: string\n}\n',
      'apps/api/src/features/orders/account-manager.policy.ts': 'export interface Data {\n  id: string\n}\n',
      'apps/api/src/features/orders/manager.policy.ts':         'export interface Data {\n  id: string\n}\n',
      // An app's slice outside every zone.
      'apps/api/src/reports/index.ts':                          "export { report } from './report.use-case'\n",
      'apps/api/src/reports/report.use-case.ts':                'export const report = 1\n',
      // A library keeps its slices directly under src.
      'libs/money/src/index.ts':                                "export { fee } from './billing'\n",
      'libs/money/src/billing/index.ts':                        "export { fee } from './fee.policy'\n",
      'libs/money/src/billing/fee.policy.ts':                   'export const fee = 1\n',
      // A library with one zone folder keeps every slice in a zone.
      'libs/office/tsconfig.json':                              FIXTURES['packages/app/tsconfig.json'],
      'libs/office/src/index.ts':                               "export { read } from './integrations/excel'\n",
      'libs/office/src/integrations/excel/index.ts':            "export { read } from './excel.client'\n",
      'libs/office/src/integrations/excel/excel.client.ts':     'export const read = 1\n',
      'libs/office/src/sheets/index.ts':                        "export { sheet } from './sheet.model'\n",
      'libs/office/src/sheets/sheet.model.ts':                  'export const sheet = 1\n',
    }
    for (const [filename, contents] of Object.entries(files)) {
      const target = join(directory, filename)
      mkdirSync(dirname(target), { recursive: true })
      writeFileSync(target, contents)
    }
    zoned = lintAll(directory)
  })

  afterAll(() => {
    rmSync(directory, { recursive: true, force: true })
  })

  const slicesIn = (filename: string): string[] =>
    (zoned[filename.startsWith('libs/') ? filename : `apps/api/src/${filename}`] ?? [])
      .filter(rule => rule.startsWith('vertical-slices/') || rule === 'FATAL')
      .toSorted((a, b) => a.localeCompare(b))

  it('passes a feature that calls an integration and its own repository, and an integration that uses vendors and shared code', () => {
    expect(slicesIn('main.ts')).toEqual([])
    expect(slicesIn('features/orders/place-order.use-case.ts')).toEqual([])
    expect(slicesIn('integrations/crm/crm.client.ts')).toEqual([])
    expect(slicesIn('shared/http/http.mapper.ts')).toEqual([])
  })

  it('reports shared and integration code that imports a feature, by alias or relative path, tests included', () => {
    expect(slicesIn('integrations/crm/order-link.mapper.ts')).toEqual(['vertical-slices/zone-import'])
    expect(slicesIn('shared/http/status.mapper.ts')).toEqual(['vertical-slices/zone-import'])
    expect(slicesIn('shared/http/http.mapper.test.ts')).toEqual(['vertical-slices/zone-import'])
  })

  it('reports a vendor package imported outside integrations, but not in a test', () => {
    expect(slicesIn('features/orders/order-scan.client.ts')).toEqual(['vertical-slices/zone-import'])
    expect(slicesIn('features/orders/order.policy.test.ts')).toEqual([])
  })

  it('keeps pure roles free of integrations, even their types, and of I/O roles', () => {
    expect(slicesIn('features/orders/order.policy.ts')).toEqual(['vertical-slices/zone-import'])
    expect(slicesIn('features/orders/order.model.ts')).toEqual(['vertical-slices/zone-import'])
  })

  it('lets a mapper, a contract and a handler use the types of an integration, never its code', () => {
    expect(slicesIn('features/orders/order-ticket.mapper.ts')).toEqual([])
    expect(slicesIn('features/orders/order-ticket.contract.ts')).toEqual([])
    expect(slicesIn('features/orders/order-send.mapper.ts')).toEqual(['vertical-slices/zone-import'])
    expect(slicesIn('features/orders/place-order.handler.ts')).toEqual(['vertical-slices/zone-import'])
  })

  it('bans data and manager only as the whole name', () => {
    expect(slicesIn('features/orders/user-data.contract.ts')).toEqual([])
    expect(slicesIn('features/orders/data.contract.ts')).toEqual(['vertical-slices/file-role'])
    expect(slicesIn('features/orders/account-manager.policy.ts')).toEqual([])
    expect(slicesIn('features/orders/manager.policy.ts')).toEqual(['vertical-slices/file-role'])
  })

  it('reports an app slice outside the zones, not a slice of a library without zones', () => {
    expect(slicesIn('reports/report.use-case.ts')).toEqual(['vertical-slices/file-role'])
    expect(slicesIn('libs/money/src/billing/fee.policy.ts')).toEqual([])
  })

  it('holds a library with one zone folder to the zones', () => {
    expect(slicesIn('libs/office/src/integrations/excel/excel.client.ts')).toEqual([])
    expect(slicesIn('libs/office/src/sheets/sheet.model.ts')).toEqual(['vertical-slices/file-role'])
  })
})
