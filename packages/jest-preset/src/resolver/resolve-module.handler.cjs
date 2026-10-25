'use strict'

const { findWithTypeScript } = require('./typescript-lookup.client.cjs')

/**
 * Jest's resolver contract: Jest's own resolution first, then TypeScript's.
 * It stands in for what `@nx/jest/plugins/resolver` did when Jest cannot find
 * a module, without Nx.
 *
 * @param {string} request - The module specifier.
 * @param {object} options - Jest's resolver options (`defaultResolver`, `basedir`, `rootDir`).
 * @returns {string | undefined} The resolved file.
 */
module.exports = function resolver (request, options) {
  try {
    return options.defaultResolver(request, options)
  } catch (error) {
    // Jest looks for a custom sequencer the same way and treats "not found" as "use the default".
    if (request === '@jest/test-sequencer' || request.startsWith('jest-sequencer-')) return
    const resolved = findWithTypeScript(request, options.basedir, options.rootDir)
    if (resolved === undefined) throw error

    return resolved
  }
}
