/**
 * A preset package is plain CommonJS that Jest loads directly: nothing is
 * compiled. The specs sit beside what they test, and the integration specs run
 * the real Jest binary against throwaway projects that use the preset.
 */
export default {
  displayName:     'jest-preset',
  testEnvironment: 'node',
  rootDir:         '.',
  roots:           ['<rootDir>/src'],
  transform:       {
    '^.+\\.ts$': ['ts-jest', { tsconfig: '<rootDir>/tsconfig.jest.json' }],
  },
  moduleFileExtensions: ['ts', 'js', 'mjs', 'cjs', 'json'],
  collectCoverage:      false,
  testTimeout:          60_000,
}
