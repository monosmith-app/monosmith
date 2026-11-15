const fs = require('node:fs')

fs.mkdirSync('dist', { recursive: true })
fs.writeFileSync('dist/index.js', "exports.greet = (who) => `hello ${who}`\n")
