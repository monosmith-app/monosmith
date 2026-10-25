'use strict'

const path = require('node:path')
const { svgModule, fileNameModule } = require('./svg-module.mapper.cjs')
const { scriptTransformer } = require('./script-transformer.client.cjs')

const SCRIPT_EXTENSIONS = new Set(['.ts', '.tsx', '.js', '.jsx', '.mjs', '.cjs'])

/**
 * A Jest transformer for what a bundler would load as a file. It stands in
 * for `@nx/react/plugins/jest`, without Nx: an imported SVG becomes a React
 * component (named as SVGR names it), a JS or TS file goes to `@swc/jest` or
 * `babel-jest`, and any other file is its file name.
 */
module.exports = {
  process (src, filename, options) {
    if (filename.endsWith('.svg')) return { code: svgModule(filename) }

    return SCRIPT_EXTENSIONS.has(path.extname(filename))
      ? scriptTransformer(options.config?.transform, options.config?.rootDir ?? process.cwd()).process(src, filename, options)
      : { code: fileNameModule(filename) }
  },
}
