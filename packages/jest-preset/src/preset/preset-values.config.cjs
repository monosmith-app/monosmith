'use strict'

/**
 * The values `@nx/jest/preset` provided, without Nx, with the TypeScript-aware
 * resolver wired in. `moduleNameMapper` lets an ESM-style `./x.js` import of a
 * TypeScript file resolve to `./x`.
 */
module.exports = {
  testMatch:            ['**/?(*.)+(spec|test).?([mc])[jt]s?(x)'],
  resolver:             require.resolve('../resolver/index.cjs'),
  moduleFileExtensions: ['ts', 'js', 'mts', 'mjs', 'cts', 'cjs', 'html'],
  coverageReporters:    ['html'],
  moduleNameMapper:     { '^(\\.{1,2}/.*)\\.js$': '$1' },
  transform:            {
    '^.+\\.(ts|js|mts|mjs|cts|cjs|html)$': ['ts-jest', { tsconfig: '<rootDir>/tsconfig.spec.json' }],
  },
  testEnvironment:          'jsdom',
  modulePathIgnorePatterns: ['<rootDir>/dist/', '<rootDir>/out-tsc/'],
  testEnvironmentOptions:   {
    customExportConditions: ['node', 'require', 'default'],
  },
}
