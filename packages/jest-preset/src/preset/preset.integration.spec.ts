import { spawnSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'

/**
 * The preset is checked the way a project uses it: the real Jest binary runs
 * the tests of a throwaway project whose whole configuration is
 * `preset: '@monosmith/jest-preset'`. The project lives inside this package, so
 * the workspace's `node_modules` (jest, ts-jest, typescript, babel) serve it.
 */

const packageRoot = join(__dirname, '..', '..')
const jestBin = join(dirname(require.resolve('jest/package.json')), 'bin', 'jest.js')

const tsconfigSpec = {
  compilerOptions: {
    target:           'es2022',
    module:           'commonjs',
    moduleResolution: 'bundler',
    strict:           true,
    esModuleInterop:  true,
    jsx:              'react-jsx',
    types:            ['jest'],
    paths:            { '@lib/*': ['./src/lib/*'] },
  },
}

let project: string

/** Writes files (a path relative to the project, and its text) into the throwaway project. */
function write (files: Record<string, string>): void {
  for (const [name, text] of Object.entries(files)) {
    mkdirSync(dirname(join(project, name)), { recursive: true })
    writeFileSync(join(project, name), text)
  }
}

/** Runs Jest over the project with the given config (the preset is always the one under test). */
function runJest (config: Record<string, unknown>): { status: number | null, output: string } {
  const result = spawnSync(process.execPath, [jestBin, '--ci', '--no-cache', '--config', JSON.stringify({ rootDir: project, ...config })], { cwd: project, encoding: 'utf8' })

  return { status: result.status, output: `${result.stdout}\n${result.stderr}` }
}

beforeEach(() => {
  project = mkdtempSync(join(packageRoot, '.fixture-'))
  write({ 'tsconfig.spec.json': JSON.stringify(tsconfigSpec) })
})

afterEach(() => {
  rmSync(project, { recursive: true, force: true })
})

describe('preset: @monosmith/jest-preset', () => {
  it('finds the specs, maps ./x.js to ./x, resolves tsconfig paths and mocks a module that only has types', () => {
    write({
      'src/lib/greet.ts':                          'export const greet = (name: string): string => `hello ${name}`\n',
      'src/counter.ts':                            'export const counter = (): number => 41\n',
      'node_modules/@types/only-types/index.d.ts': 'export declare const flavour: string\n',
      'src/uses-all.spec.ts':                      `
import { flavour } from 'only-types'
import { greet } from '@lib/greet'
import { counter } from './counter.js'

jest.mock('only-types', () => ({ flavour: 'mocked' }))

it('resolves through the preset', () => {
  expect(greet('smith')).toBe('hello smith')
  expect(counter() + 1).toBe(42)
  expect(flavour).toBe('mocked')
})
`,
    })

    const { status, output } = runJest({ preset: '@monosmith/jest-preset', testEnvironment: 'node' })

    expect(output).toContain('1 passed')
    expect(status).toBe(0)
  })

  it('reports a module nobody can resolve, with Jest\'s own error', () => {
    write({ 'src/missing.spec.ts': "it('imports nothing real', () => { jest.requireActual('definitely-not-installed') })\n" })

    const { status, output } = runJest({ preset: '@monosmith/jest-preset', testEnvironment: 'node' })

    expect(status).not.toBe(0)
    expect(output).toContain("Cannot find module 'definitely-not-installed'")
  })

  it('keeps what a project adds beside the preset', () => {
    write({
      'src/value.ts':     'export const value = 7\n',
      'src/setup.ts':     '(globalThis as any).fromSetup = 6\n',
      'src/adds.spec.ts': "import { value } from './value'\nit('sees the setup file', () => { expect(value + (globalThis as any).fromSetup).toBe(13) })\n",
    })

    const { status, output } = runJest({ preset: '@monosmith/jest-preset', testEnvironment: 'node', setupFiles: ['<rootDir>/src/setup.ts'] })

    expect(output).toContain('1 passed')
    expect(status).toBe(0)
  })
})

describe('preset: @monosmith/jest-preset/react', () => {
  it('runs a component test with an SVG import, an image import and JSX', () => {
    write({
      'src/logo.svg':  '<svg xmlns="http://www.w3.org/2000/svg"></svg>\n',
      'src/photo.png': 'not really a png',
      'src/app.tsx':   `
import logo, { ReactComponent as Logo } from './logo.svg'
import photo from './photo.png'

export function App () {
  return <main><Logo data-logo={logo} data-photo={photo} /></main>
}
`,
      'src/app.spec.tsx': `
import { renderToStaticMarkup } from 'react-dom/server'
import { App } from './app'

it('renders the SVG as a component and the PNG as its name', () => {
  const html = renderToStaticMarkup(<App />)
  expect(html).toContain('data-logo="logo.svg"')
  expect(html).toContain('data-photo="photo.png"')
})
`,
    })

    const { status, output } = runJest({ preset: '@monosmith/jest-preset/react' })

    expect(output).toContain('1 passed')
    expect(status).toBe(0)
  })
})

describe('the values', () => {
  it('are what @nx/jest/preset provided, with the resolver wired in', () => {
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    const preset = require('../../jest-preset.cjs') as Record<string, unknown>

    expect(preset.testMatch).toEqual(['**/?(*.)+(spec|test).?([mc])[jt]s?(x)'])
    expect(preset.moduleNameMapper).toEqual({ '^(\\.{1,2}/.*)\\.js$': '$1' })
    expect(preset.coverageReporters).toEqual(['html'])
    expect(preset.moduleFileExtensions).toEqual(['ts', 'js', 'mts', 'mjs', 'cts', 'cjs', 'html'])
    expect(preset.resolver).toBe(join(packageRoot, 'src', 'resolver', 'index.cjs'))
  })
})
