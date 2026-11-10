'use strict'

// What a style import (`import './app.css'`, `import styles from './app.module.scss'`)
// is in a test: an object that answers every class name with the name itself,
// so `styles.title` is 'title' and a plain import does nothing.
module.exports = new Proxy({}, {
  get: (_target, key) => (key === '__esModule' ? false : key),
})
