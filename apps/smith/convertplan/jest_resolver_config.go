package convertplan

// JestResolverFile is the workspace-root module the written-out Jest preset
// names as its resolver, in place of @nx/jest/plugins/resolver.
const JestResolverFile = "jest.resolver.cjs"

// JestResolver is the part of Nx's Jest resolver a workspace without Nx still
// needs: when Jest cannot find a module, TypeScript is asked. It finds what
// only has types (jest.mock('vscode') with just @types/vscode installed, which
// the mock replaces anyway) and what a tsconfig `paths` entry names.
const JestResolver = `// Written by smith convert: what @nx/jest/plugins/resolver did when Jest cannot
// find a module, without Nx. TypeScript is asked: it finds a module that only
// has types (jest.mock('vscode') with just @types/vscode installed) and what a
// tsconfig "paths" entry names.
const path = require('node:path')

let compiler

function compilerSetup (rootDir) {
  const ts = require('typescript')
  const configFile = ['tsconfig.spec.json', 'tsconfig.test.json', 'tsconfig.jest.json', 'tsconfig.json']
    .map((name) => ts.findConfigFile(rootDir, ts.sys.fileExists, name))
    .find(Boolean)
  const options = configFile
    ? ts.parseJsonConfigFileContent(ts.readConfigFile(configFile, ts.sys.readFile).config, ts.sys, path.dirname(configFile)).options
    : {}

  return { ts, options, host: ts.createCompilerHost(options, true) }
}

module.exports = function resolver (request, options) {
  try {
    return options.defaultResolver(request, options)
  } catch (error) {
    if (request === '@jest/test-sequencer' || request.startsWith('jest-sequencer-')) return undefined
    compiler ??= compilerSetup(options.rootDir)
    const { ts, options: compilerOptions, host } = compiler
    const resolved = ts.resolveModuleName(request, path.join(options.basedir, 'fake-placeholder.ts'), compilerOptions, host).resolvedModule?.resolvedFileName
    if (!resolved) throw error

    return path.resolve(options.rootDir, resolved)
  }
}
`
