'use strict'

const base = require('./preset-values.config.cjs')

/**
 * The base preset for a React project: what `@nx/react` added to a project's
 * Jest config. It adds `tsx` and `jsx` to the module file extensions; files
 * that are not scripts, styles or JSON load through the asset transform; scripts go through `babel-jest` with the presets
 * `@nx/react/babel` stood for (`@babel/preset-env`, `@babel/preset-typescript`
 * and React's automatic JSX runtime).
 */
module.exports = {
  ...base,
  moduleFileExtensions: [...base.moduleFileExtensions, 'tsx', 'jsx'],
  transform:            {
    '^(?!.*\\.(js|jsx|ts|tsx|css|json)$)': require.resolve('../asset-transform/index.cjs'),
    '^.+\\.[tj]sx?$':                      ['babel-jest', {
      presets: [
        ['@babel/preset-env', { targets: { node: 'current' }, loose: true }],
        ['@babel/preset-typescript', { allowDeclareFields: true }],
        ['@babel/preset-react', { runtime: 'automatic', development: true }],
      ],
    }],
  },
}
