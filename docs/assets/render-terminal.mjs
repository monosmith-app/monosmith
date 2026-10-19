#!/usr/bin/env node
// Renders captured terminal output (plain text or ANSI colours) as a PNG, or as
// an HTML page when no browser is available. The screenshots in docs/assets were
// made with it; see docs/assets/README.md.
//
//   node docs/assets/render-terminal.mjs <input-file> <output.png|output.html>
//        [--title "acme — smith test"] [--cmd "smith test"] [--width 900] [--scale 2]
//        [--trim-top N] [--max-lines N]
//
// <input-file> holds the text exactly as the terminal printed it. --cmd adds a
// prompt line ("$ smith test") above it. The PNG needs the `playwright` package,
// found from the current directory (npm i -D playwright, then npx playwright
// install chromium; or set RENDER_CHANNEL=chrome to use an installed Chrome).

import { readFileSync, writeFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { pathToFileURL } from 'node:url'
import { resolve } from 'node:path'

const args = process.argv.slice(2)
const positional = []
const options = { width: 900, scale: 2, title: '', cmd: '', 'trim-top': 0, 'max-lines': 0 }
for (let i = 0; i < args.length; i++) {
  if (args[i].startsWith('--')) options[args[i].slice(2)] = args[++i]
  else positional.push(args[i])
}
const [inputFile, outputFile] = positional
if (!inputFile || !outputFile) {
  console.error('usage: render-terminal.mjs <input-file> <output.png|output.html> [--title t] [--cmd c] [--width px] [--scale n]')
  process.exit(2)
}

// One Dark-like palette for the 16 ANSI colours.
const palette = [
  '#282c34', '#e06c75', '#98c379', '#e5c07b', '#61afef', '#c678dd', '#56b6c2', '#abb2bf',
  '#5c6370', '#ff7b86', '#a9e08a', '#f0d28a', '#7ac0ff', '#d68aee', '#6fd0dc', '#ffffff'
]
const foreground = '#d7dae0'
const background = '#1e2127'

function color256 (n) {
  if (n < 16) return palette[n]
  if (n >= 232) { const v = 8 + (n - 232) * 10; return `rgb(${v},${v},${v})` }
  n -= 16
  const level = (x) => (x === 0 ? 0 : 55 + x * 40)
  return `rgb(${level(Math.floor(n / 36))},${level(Math.floor(n / 6) % 6)},${level(n % 6)})`
}

const escapeHtml = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

function toHtml (text) {
  let state = {}
  let out = ''
  const open = () => {
    const css = []
    let fg = state.fg
    let bg = state.bg
    if (state.reverse) { [fg, bg] = [bg ?? background, fg ?? foreground] }
    if (fg) css.push(`color:${fg}`)
    if (bg) css.push(`background:${bg}`)
    if (state.bold) css.push('font-weight:700')
    if (state.dim) css.push('opacity:.6')
    if (state.italic) css.push('font-style:italic')
    if (state.underline) css.push('text-decoration:underline')
    return css.length ? `<span style="${css.join(';')}">` : ''
  }
  const re = /\x1b\[([0-9;]*)([A-Za-z])|\x1b\][^\x07]*\x07/g
  let last = 0
  let m
  let span = ''
  const emit = (chunk) => {
    if (!chunk) return
    const html = escapeHtml(chunk)
    out += span ? `${span}${html}</span>` : html
  }
  while ((m = re.exec(text)) !== null) {
    emit(text.slice(last, m.index))
    last = re.lastIndex
    if (m[2] !== 'm') continue
    const codes = (m[1] === '' ? '0' : m[1]).split(';').map(Number)
    for (let i = 0; i < codes.length; i++) {
      const c = codes[i]
      if (c === 0) state = {}
      else if (c === 1) state.bold = true
      else if (c === 2) state.dim = true
      else if (c === 3) state.italic = true
      else if (c === 4) state.underline = true
      else if (c === 7) state.reverse = true
      else if (c === 22) { state.bold = false; state.dim = false }
      else if (c === 23) state.italic = false
      else if (c === 24) state.underline = false
      else if (c === 27) state.reverse = false
      else if (c >= 30 && c <= 37) state.fg = palette[c - 30]
      else if (c >= 90 && c <= 97) state.fg = palette[c - 90 + 8]
      else if (c >= 40 && c <= 47) state.bg = palette[c - 40]
      else if (c >= 100 && c <= 107) state.bg = palette[c - 100 + 8]
      else if (c === 39) state.fg = undefined
      else if (c === 49) state.bg = undefined
      else if (c === 38 || c === 48) {
        const key = c === 38 ? 'fg' : 'bg'
        if (codes[i + 1] === 5) { state[key] = color256(codes[i + 2]); i += 2 } else if (codes[i + 1] === 2) { state[key] = `rgb(${codes[i + 2]},${codes[i + 3]},${codes[i + 4]})`; i += 4 }
      }
    }
    span = open()
  }
  emit(text.slice(last))
  return out
}

let text = readFileSync(inputFile, 'utf8').replace(/\r\n/g, '\n').replace(/\n+$/, '')
let lines = text.split('\n')
if (Number(options['trim-top'])) lines = lines.slice(Number(options['trim-top']))
if (Number(options['max-lines'])) lines = lines.slice(0, Number(options['max-lines']))
text = lines.join('\n')
const prompt = options.cmd ? `<span style="color:${palette[2]}">$</span> <span style="color:#ffffff;font-weight:700">${escapeHtml(options.cmd)}</span>\n` : ''

const page = `<!doctype html>
<html><head><meta charset="utf-8"><title>${escapeHtml(options.title || 'terminal')}</title>
<style>
  html,body{margin:0;background:transparent}
  body{padding:12px;display:inline-block}
  .term{width:${Number(options.width)}px;border-radius:10px;overflow:hidden;background:${background};
        box-shadow:0 0 0 1px #3a3f4b;font-family:"Cascadia Mono","Cascadia Code",Consolas,"DejaVu Sans Mono","Liberation Mono",monospace}
  .bar{display:flex;align-items:center;gap:7px;padding:9px 12px;background:#2b2f38;color:#9aa1ad;font-size:12px}
  .dot{width:11px;height:11px;border-radius:50%;display:inline-block}
  pre{margin:0;padding:14px 16px 16px;color:${foreground};font-size:13px;line-height:1.38;white-space:pre-wrap;word-break:break-word;font-family:inherit}
</style></head>
<body><div class="term" id="term">
<div class="bar"><span class="dot" style="background:#ff5f57"></span><span class="dot" style="background:#febc2e"></span><span class="dot" style="background:#28c840"></span><span style="margin-left:8px">${escapeHtml(options.title || '')}</span></div>
<pre>${prompt}${toHtml(text)}</pre>
</div></body></html>`

if (outputFile.endsWith('.html')) {
  writeFileSync(outputFile, page)
  console.log(`wrote ${outputFile}`)
} else {
  const require = createRequire(pathToFileURL(resolve('.') + '/'))
  let chromium
  try { ({ chromium } = require('playwright')) } catch { console.error('the playwright package is needed for PNG output (or write an .html file)'); process.exit(1) }
  // RENDER_CHANNEL=chrome or msedge uses an installed browser instead of a downloaded Chromium.
  const browser = await chromium.launch(process.env.RENDER_CHANNEL ? { channel: process.env.RENDER_CHANNEL } : {})
  const context = await browser.newContext({ deviceScaleFactor: Number(options.scale), viewport: { width: Number(options.width) + 40, height: 600 } })
  const tab = await context.newPage()
  await tab.setContent(page)
  await tab.locator('#term').screenshot({ path: outputFile, omitBackground: true })
  await browser.close()
  console.log(`wrote ${outputFile}`)
}
