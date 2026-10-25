#!/usr/bin/env node
// Generates the master table in everything-compared.md from features.json.
//
//   node docs/product/comparison/build-everything-compared.mjs           rewrite the generated blocks
//   node docs/product/comparison/build-everything-compared.mjs --check   exit 1 when the page is out of date
//   node docs/product/comparison/build-everything-compared.mjs --report  print the derived row lists
//
// The blocks between the <!-- everything-compared:... --> markers are generated; the rest of the page
// is written by hand. --check also compares the row lists the page names in the
// <!-- rows:only-smith ... --> and <!-- rows:not-yet ... --> comments with the lists derived from the data.

import { readFileSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const data = JSON.parse(readFileSync(join(here, 'features.json'), 'utf8'))
const pagePath = join(here, 'everything-compared.md')

const MARK = { Y: '✅', P: '◐', U: '❔', N: '❌', '-': '–' }
const HEAD = {
  nx: 'Nx', turborepo: 'Turbo', bazel: 'Bazel', pants: 'Pants', lerna: 'Lerna', rush: 'Rush', moon: 'moon',
  buck2: 'Buck2', gradle: 'Gradle', earthly: 'Earthly', lage: 'Lage', workspaces: 'JS pkg mgrs', cargo: 'Cargo', please: 'Please',
}
const tools = data.tools
const problems = []

for (const r of data.rows) {
  if (!data.groups.some((g) => g.id === r.g)) problems.push(`${r.id}: unknown group ${r.g}`)
  if (Array.from(r.t).length !== tools.length) problems.push(`${r.id}: t has ${Array.from(r.t).length} marks, want ${tools.length}`)
  for (const c of Array.from(r.t)) if (!(c in MARK)) problems.push(`${r.id}: bad mark ${c}`)
  if (!(r.s.m in MARK) || r.s.m === 'U') problems.push(`${r.id}: bad smith mark ${r.s.m}`)
  if (!(r.s.e in data.evidenceCodes)) problems.push(`${r.id}: bad evidence ${r.s.e}`)
  if (r.s.m !== 'Y' && !r.s.p) problems.push(`${r.id}: a ${r.s.m} needs a planned reference`)
  if (Array.from(r.t).includes('U') && !r.searched && !r.n) {
    // a row with unverified cells names what was searched, or the note says why
  }
}
if (new Set(data.rows.map((r) => r.id)).size !== data.rows.length) problems.push('duplicate row id')
if (problems.length) {
  console.error(problems.join('\n'))
  process.exit(2)
}

const marks = (r) => Array.from(r.t)

function onlySmith() {
  return data.rows.filter((r) => !r.weak && r.s.m === 'Y' && marks(r).every((c) => c === 'U' || c === 'N' || c === '-')).map((r) => r.id)
}
function nearlyOnly() {
  return data.rows.filter((r) => !r.weak && r.s.m === 'Y' && !marks(r).includes('Y') && marks(r).includes('P')).map((r) => r.id)
}
function weakOnly() {
  return data.rows.filter((r) => r.weak && r.s.m === 'Y' && marks(r).every((c) => c === 'U' || c === 'N' || c === '-')).map((r) => r.id)
}
function notYet() {
  return data.rows.filter((r) => r.s.m !== 'Y' && marks(r).some((c) => c === 'Y')).map((r) => r.id)
}
function gapsWithoutRivals() {
  return data.rows.filter((r) => r.s.m !== 'Y' && !marks(r).some((c) => c === 'Y')).map((r) => r.id)
}

function esc(s) {
  return s.replace(/\|/g, '\\|')
}

function tableBlock() {
  const out = []
  const head = ['Capability', '**MonoSmith**', ...tools.map((t) => HEAD[t]), 'Smith: evidence, issue']
  out.push(`| ${head.join(' | ')} |`)
  out.push(`|${head.map((h, i) => (i === 0 ? '---' : ':---:')).join('|')}|`)
  const cols = new Array(tools.length).fill(0)
  const unk = new Array(tools.length).fill(0)
  let smithY = 0
  let smithP = 0
  let smithN = 0
  for (const g of data.groups) {
    const rows = data.rows.filter((r) => r.g === g.id)
    if (!rows.length) continue
    out.push(`| **${g.title}** |${' |'.repeat(head.length - 1)}`)
    for (const r of rows) {
      const m = marks(r)
      m.forEach((c, i) => {
        if (c === 'Y') cols[i]++
        if (c === 'U') unk[i]++
      })
      if (r.s.m === 'Y') smithY++
      if (r.s.m === 'P') smithP++
      if (r.s.m === 'N') smithN++
      const ev = r.s.e + (r.s.p ? `, ${esc(r.s.p)}` : '')
      out.push(`| ${esc(r.label)} | **${MARK[r.s.m]}** | ${m.map((c) => MARK[c]).join(' | ')} | ${ev} |`)
    }
  }
  out.push(`| **Score: rows marked ✅** (of ${data.rows.length}) | **${smithY}** | ${cols.join(' | ')} | ◐ ${smithP}, ❌ ${smithN} |`)
  out.push(`| **Rows marked ❔** (not verified, not counted as absent) | – | ${unk.join(' | ')} | |`)
  return out.join('\n')
}

function notesBlock() {
  const out = []
  for (const g of data.groups) {
    const rows = data.rows.filter((r) => r.g === g.id && (r.n || r.s.n || r.searched))
    if (!rows.length) continue
    out.push(`**${g.title}**`)
    out.push('')
    for (const r of rows) {
      const parts = []
      if (r.s.n) parts.push(`MonoSmith: ${r.s.n}.`)
      if (r.n) parts.push(r.n.endsWith('.') ? r.n : `${r.n}.`)
      if (r.searched) parts.push(`Searched for ❔: ${r.searched}.`)
      out.push(`- *${r.label}.* ${parts.join(' ')}`)
    }
    out.push('')
  }
  return out.join('\n').trimEnd()
}

const BLOCKS = { table: tableBlock(), notes: notesBlock() }

function replaceBlock(text, name, body) {
  const begin = `<!-- everything-compared:${name}:begin -->`
  const end = `<!-- everything-compared:${name}:end -->`
  const a = text.indexOf(begin)
  const b = text.indexOf(end)
  if (a < 0 || b < 0 || b < a) throw new Error(`markers for ${name} not found in everything-compared.md`)
  return `${text.slice(0, a + begin.length)}\n\n${body}\n\n${text.slice(b)}`
}

const mode = process.argv[2]
if (mode === '--report') {
  console.log('rows:', data.rows.length)
  console.log('only-smith (smith ✅, every other cell ❌ ❔ –):', onlySmith().join(', '))
  console.log('only-smith but searched too little to claim:', weakOnly().join(', '))
  console.log('nearly-only (smith ✅, others at most ◐):', nearlyOnly().join(', '))
  console.log('not-yet (smith ❌ or ◐, at least one tool ✅):', notYet().join(', '))
  console.log('gaps without a rival ✅:', gapsWithoutRivals().join(', '))
  process.exit(0)
}

let page = readFileSync(pagePath, 'utf8')
let next = page
for (const [name, body] of Object.entries(BLOCKS)) next = replaceBlock(next, name, body)

const listed = (name) => {
  const m = next.match(new RegExp(`<!-- rows:${name} ([^>]*?) -->`))
  return m ? m[1].split(',').map((s) => s.trim()).filter(Boolean) : null
}
const same = (a, b) => a.length === b.length && [...a].sort().join() === [...b].sort().join()
const listProblems = []
for (const [name, derived] of [['only-smith', onlySmith()], ['not-yet', notYet()], ['no-rival', gapsWithoutRivals()], ['weak', weakOnly()], ['nearly-only', nearlyOnly()]]) {
  const got = listed(name)
  if (!got) listProblems.push(`the page has no <!-- rows:${name} ... --> comment`)
  else if (!same(got, derived)) listProblems.push(`rows:${name} on the page is [${got.join(', ')}] but the data derives [${derived.join(', ')}]`)
}

if (mode === '--check') {
  let bad = false
  if (next !== page) {
    console.error('everything-compared.md is out of date: run node docs/product/comparison/build-everything-compared.mjs')
    bad = true
  }
  if (listProblems.length) {
    console.error(listProblems.join('\n'))
    bad = true
  }
  if (bad) process.exit(1)
  console.log('everything-compared.md is up to date')
} else {
  if (next !== page) writeFileSync(pagePath, next)
  if (listProblems.length) console.warn(listProblems.join('\n'))
  console.log(next !== page ? 'everything-compared.md rewritten' : 'everything-compared.md already up to date')
}
