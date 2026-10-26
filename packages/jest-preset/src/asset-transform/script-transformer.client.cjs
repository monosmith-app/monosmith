'use strict'

const transformers = new Map()

/**
 * Loads a package from the Jest project, else from beside this one.
 *
 * @param {string} name - The package.
 * @param {string} rootDir - The Jest project's root.
 * @returns {object} Its exports.
 */
function loadFrom (name, rootDir) {
  try {
    return require(require.resolve(name, { paths: [rootDir] }))
  } catch {
    return require(name)
  }
}

/**
 * The transformer for the JavaScript and TypeScript files the asset transform
 * is asked about: `@swc/jest` when the project's own transforms use it, else
 * `babel-jest`.
 *
 * @param {Array<unknown> | undefined} transforms - The project's Jest transform entries.
 * @param {string} rootDir - The Jest project's root.
 * @returns {{ process: Function }} A transformer.
 */
function scriptTransformer (transforms, rootDir) {
  const swc = (transforms ?? []).some((t) => Array.isArray(t) && t.some((v) => typeof v === 'string' && v.includes('@swc/jest')))
  const key = `${swc ? 'swc' : 'babel'}:${rootDir}`
  let transformer = transformers.get(key)
  if (transformer === undefined) {
    const loaded = swc ? loadFrom('@swc/jest', rootDir) : loadFrom('babel-jest', rootDir).default
    transformer = loaded.createTransformer()
    transformers.set(key, transformer)
  }

  return transformer
}

module.exports = { scriptTransformer }
