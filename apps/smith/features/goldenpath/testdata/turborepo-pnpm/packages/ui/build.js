const fs = require('node:fs')
// Resolving @acme/util needs its build, which is what ^build orders first.
const util = require('@acme/util')

fs.mkdirSync('dist', { recursive: true })
fs.writeFileSync('dist/index.js', `exports.banner = ${JSON.stringify(util.greet('ui'))}\n`)
