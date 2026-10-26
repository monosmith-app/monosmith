'use strict'

const path = require('node:path')
const { componentName } = require('./component-name.algorithm.cjs')

/**
 * The CommonJS module an imported SVG becomes: its file name as the default
 * export, and a React component as `ReactComponent`.
 *
 * @param {string} filename - The SVG's path.
 * @returns {string} The module's code.
 */
function svgModule (filename) {
  const assetFilename = JSON.stringify(path.basename(filename))
  const name = 'Svg' + componentName(path.parse(filename).name)

  return `const React = require('react');
module.exports = {
  __esModule: true,
  default: ${assetFilename},
  ReactComponent: function ${name}(props) {
    return React.createElement('svg', Object.assign({}, props, { children: ${assetFilename} }));
  },
};`
}

/**
 * The CommonJS module any other asset (an image, a font) becomes: its file name.
 *
 * @param {string} filename - The asset's path.
 * @returns {string} The module's code.
 */
function fileNameModule (filename) {
  return `module.exports = ${JSON.stringify(path.basename(filename))};`
}

module.exports = { svgModule, fileNameModule }
