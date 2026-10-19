const assert = require('node:assert')

assert.strictEqual(require('./dist/index.js').banner, 'hello ui')
