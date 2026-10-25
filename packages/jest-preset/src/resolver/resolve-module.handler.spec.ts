import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import resolver from './resolve-module.handler.cjs'

type Options = Parameters<typeof resolver>[1]

/** Jest's options for a request made from `basedir`, with a default resolver that finds only what `known` names. */
function optionsFor (rootDir: string, known: Record<string, string> = {}): Options {
  return {
    basedir:         rootDir,
    rootDir,
    defaultResolver: (request: string) => {
      if (Object.hasOwn(known, request)) return known[request]
      throw new Error(`Cannot find module '${request}'`)
    },
  }
}

describe('the resolver', () => {
  let root: string

  beforeEach(() => {
    root = mkdtempSync(join(tmpdir(), 'jest-preset-resolver-'))
  })

  afterEach(() => {
    rmSync(root, { recursive: true, force: true })
  })

  it('uses what Jest finds first', () => {
    expect(resolver('left-pad', optionsFor(root, { 'left-pad': '/found/left-pad.js' }))).toBe('/found/left-pad.js')
  })

  it('asks TypeScript for a tsconfig paths entry', () => {
    mkdirSync(join(root, 'lib'))
    writeFileSync(join(root, 'lib', 'greet.ts'), 'export const greet = 1\n')
    writeFileSync(join(root, 'tsconfig.json'), JSON.stringify({ compilerOptions: { module: 'commonjs', moduleResolution: 'bundler', paths: { '@lib/*': ['./lib/*'] } } }))

    expect(resolver('@lib/greet', optionsFor(root))).toBe(join(root, 'lib', 'greet.ts'))
  })

  it('asks TypeScript for a module that only has types', () => {
    const types = join(root, 'node_modules', '@types', 'only-types')
    mkdirSync(types, { recursive: true })
    writeFileSync(join(types, 'index.d.ts'), 'export declare const x: number\n')
    writeFileSync(join(root, 'tsconfig.json'), JSON.stringify({ compilerOptions: { module: 'commonjs', moduleResolution: 'bundler' } }))

    expect(resolver('only-types', optionsFor(root))).toBe(join(types, 'index.d.ts'))
  })

  it('reports Jest\'s own error when neither finds the module', () => {
    expect(() => resolver('nowhere-to-be-found', optionsFor(root))).toThrow("Cannot find module 'nowhere-to-be-found'")
  })

  it('leaves the test sequencer to Jest', () => {
    expect(resolver('@jest/test-sequencer', optionsFor(root))).toBeUndefined()
    expect(resolver('jest-sequencer-custom', optionsFor(root))).toBeUndefined()
  })
})
