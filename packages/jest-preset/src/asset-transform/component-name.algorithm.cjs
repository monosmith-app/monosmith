'use strict'

/**
 * The name SVGR gives the component of an SVG file: its file name in
 * PascalCase, without anything that is not a letter or a digit.
 *
 * @param {string} name - The file name without its extension.
 * @returns {string} The PascalCase name.
 */
function componentName (name) {
  const camel = name
    .replaceAll(/([^a-z0-9])+(.)?/gi, (_, __, chr) => (chr ? chr.toUpperCase() : ''))
    .replaceAll(/[^a-z\d]/gi, '')

  return camel.charAt(0).toUpperCase() + camel.slice(1)
}

module.exports = { componentName }
