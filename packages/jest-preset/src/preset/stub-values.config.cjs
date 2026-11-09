'use strict'

/**
 * What a bundler would load as a style or a file is stubbed in a test, as the
 * Nx Jest setup did: a style import is an object that answers every class name
 * with itself, an image, font or media import is a string. Both are
 * moduleNameMapper entries, so the file does not have to exist or parse.
 */
module.exports = {
  styles: { '\\.(css|less|scss|sass|styl)$': require.resolve('../../style-stub.cjs') },
  files:  { '\\.(png|jpe?g|gif|webp|avif|bmp|ico|woff2?|ttf|eot|otf|mp3|mp4|webm|wav|ogg)$': require.resolve('../../file-stub.cjs') },
}
