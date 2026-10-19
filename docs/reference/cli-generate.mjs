#!/usr/bin/env node
// Keeps docs/reference/cli.md in step with the smith binary.
//
// cli.md is written by hand except for the blocks between
//
//     <!-- smith:begin <command> -->
//     <!-- smith:end <command> -->
//
// This script fills each block from `smith commands --json`: the usage line, the
// aliases, and a table of the command's flags. Run it from the workspace root
// after building smith:
//
//     node docs/reference/cli-generate.mjs            rewrite the blocks
//     node docs/reference/cli-generate.mjs --check    exit 1 when cli.md is out of date
//
// SMITH names the binary to ask (default: `smith` on the PATH). It exits 1, and
// says so, when a command has no block or a block names a command smith no
// longer has, so a new command cannot slip past the reference.

import { execFileSync } from 'node:child_process'
import { readFileSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const target = join(here, 'cli.md')
const check = process.argv.includes('--check')
const smith = process.env.SMITH || 'smith'

// `smith commands --json` carries no default values, so the ones that matter are
// kept here, as `command flag-name`. Defaults already written in a flag's own
// description ("(default: ...)") are left in the description.
const defaults = {
  'add lang': "the stack's, or the workspace's only language",
  'add stack': 'the recipe, else `[default_stacks]` in `smith.toml`',
  'affected base': 'the pull request target, else the release branch',
  'graph base': 'the pull request target, else the release branch',
  'graph format': '`text`',
  'new ci': '`github`',
  'new dir': '`./<name>`',
  'new module-prefix': 'from the `origin` remote, else `example.com/<name>`',
  'new scope': '`@<name>`',
  'pipeline ci': "the workspace's providers"
}
for (const run of ['run', 'build', 'lint', 'test', 'typecheck']) {
  defaults[`${run} base`] = 'the pull request target, else the release branch'
  defaults[`${run} parallel`] = 'the number of CPUs'
}

const commands = JSON.parse(execFileSync(smith, ['commands', '--json'], { encoding: 'utf8', maxBuffer: 16 << 20 }))
const byName = new Map(commands.map((c) => [c.name, c]))

const escapeCell = (s) => s.replace(/\|/g, '\\|')

function usage (c) {
  const args = c.arguments.map((a) => {
    const name = a.variadic ? `${a.name}...` : a.name
    return a.required ? `<${name}>` : `[${name}]`
  })
  return ['smith', c.name, ...args, ...(c.options.length ? ['[flags]'] : [])].join(' ')
}

function block (c) {
  const lines = [c.description, '', `Usage: \`${usage(c)}\``]
  if (c.aliases?.length) lines.push('', `Alias: ${c.aliases.map((a) => `\`smith ${a}\``).join(', ')}`)
  if (c.options.length) {
    lines.push('', '| Flag | Meaning | Default |', '|---|---|---|')
    for (const o of c.options) {
      const key = `${c.name} ${o.name}`
      let meaning = o.description
      let def = defaults[key] ?? ''
      const inline = / \(default[: ]([^)]*)\)/.exec(meaning)
      if (inline && !def) {
        def = inline[1].replace(/^"|"$/g, '`')
      }
      if (o.repeatable) meaning += ' (repeatable)'
      if (!def) def = o.takesValue === 'none' ? 'off' : ''
      lines.push(`| \`${o.flags}\` | ${escapeCell(meaning)} | ${escapeCell(def)} |`)
    }
  } else {
    lines.push('', 'No flags.')
  }
  return lines.join('\n')
}

const original = readFileSync(target, 'utf8')
const seen = new Set()
const problems = []
const re = /<!-- smith:begin ([\w-]+) -->\n[\s\S]*?<!-- smith:end \1 -->/g
const updated = original.replace(re, (whole, name) => {
  seen.add(name)
  const c = byName.get(name)
  if (!c) {
    problems.push(`cli.md has a block for "${name}", which smith does not have`)
    return whole
  }
  return `<!-- smith:begin ${name} -->\n${block(c)}\n\n<!-- smith:end ${name} -->`
})
for (const c of commands) if (!seen.has(c.name)) problems.push(`smith ${c.name} has no block in cli.md`)

for (const p of problems) console.error(p)
if (check) {
  if (updated !== original) problems.push('cli.md is out of date: run node docs/reference/cli-generate.mjs')
  if (problems.length) process.exit(1)
  console.log('cli.md is up to date')
} else {
  writeFileSync(target, updated)
  console.log(`updated ${seen.size} blocks in docs/reference/cli.md`)
  if (problems.length) process.exit(1)
}
