'use strict'

const path = require('node:path')

const CONFIG_NAMES = ['tsconfig.spec.json', 'tsconfig.test.json', 'tsconfig.jest.json', 'tsconfig.json']

/** One compiler setup per project root: a multi-project run resolves for several roots in one process. */
const setups = new Map()

/**
 * The project's own TypeScript (so its version reads its tsconfig), else the
 * one installed beside this package.
 *
 * @param {string} rootDir - The Jest project's root.
 * @returns {import('typescript')} The compiler API.
 */
function loadTypeScript (rootDir) {
  try {
    return require(require.resolve('typescript', { paths: [rootDir] }))
  } catch {
    return require('typescript')
  }
}

/**
 * The compiler options of the tsconfig a test run would use, with a host to read files through.
 *
 * @param {string} rootDir - The Jest project's root.
 * @returns {{ ts: import('typescript'), options: object, host: object }} The setup.
 */
function setupFor (rootDir) {
  const known = setups.get(rootDir)
  if (known !== undefined) return known

  const ts = loadTypeScript(rootDir)
  const configFile = CONFIG_NAMES
    .map((name) => ts.findConfigFile(rootDir, ts.sys.fileExists, name))
    .find(Boolean)
  const options = configFile
    ? ts.parseJsonConfigFileContent(ts.readConfigFile(configFile, ts.sys.readFile).config, ts.sys, path.dirname(configFile)).options
    : {}
  const setup = { ts, options, host: ts.createCompilerHost(options, true) }
  setups.set(rootDir, setup)

  return setup
}

/**
 * Asks TypeScript where a module is: it finds one that only has types
 * (`jest.mock('vscode')` with just `@types/vscode` installed) and what a
 * tsconfig `paths` entry names.
 *
 * @param {string} request - The module specifier.
 * @param {string} basedir - The directory of the importing file.
 * @param {string} rootDir - The Jest project's root.
 * @returns {string | undefined} The resolved file, or `undefined` when TypeScript cannot find it or is not installed.
 */
function findWithTypeScript (request, basedir, rootDir) {
  let setup
  try {
    setup = setupFor(rootDir)
  } catch {
    return
  }
  const { ts, options, host } = setup
  const resolved = ts.resolveModuleName(request, path.join(basedir, 'fake-placeholder.ts'), options, host).resolvedModule?.resolvedFileName

  return resolved ? path.resolve(rootDir, resolved) : undefined
}

module.exports = { findWithTypeScript }
