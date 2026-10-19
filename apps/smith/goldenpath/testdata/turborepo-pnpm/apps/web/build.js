const fs = require('node:fs')
const { banner } = require('@acme/ui')

fs.mkdirSync('dist', { recursive: true })
fs.writeFileSync('dist/main.js', `console.log(${JSON.stringify(banner)})\n`)
