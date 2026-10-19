const assert = require('node:assert')
const { execFileSync } = require('node:child_process')

assert.strictEqual(execFileSync('node', ['dist/main.js']).toString().trim(), 'hello ui')
