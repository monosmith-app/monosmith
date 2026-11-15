const assert = require('node:assert')

assert.strictEqual(require('./dist/index.js').greet('util'), 'hello util')
