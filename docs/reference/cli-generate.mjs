#!/usr/bin/env node
// Keeps the flag tables of the command documentation in step with the smith binary.
//
// Two places carry generated blocks:
//
//   docs/reference/cli.md   one compact block per top-level command
//   docs/manual/*.md        one full block per command and subcommand
//
// A block sits between
//
//     <!-- smith:begin <command path> -->
//     <!-- smith:end <command path> -->
//
// where the command path is the words a user types after `smith`, for example
// `new` or `ci generate`. The script fills each block from `smith commands --json`:
// the usage line, the aliases, the arguments and a table of the flags. Everything
// outside a block is written by hand.
//
// Run it from the workspace root after building smith:
//
//     node docs/reference/cli-generate.mjs            rewrite the blocks
//     node docs/reference/cli-generate.mjs --check    exit 1 when a page is out of date
//
// SMITH names the binary to ask (default: `smith` on the PATH). The script exits 1,
// and says so, when a command has no block, or a block names a command smith no
// longer has, so a new command or flag cannot slip past the documentation.

import { execFileSync } from 'node:child_process'
import { readFileSync, readdirSync, writeFileSync, existsSync } from 'node:fs'
import { dirname, join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const docs = join(here, '..')
const check = process.argv.includes('--check')
const smith = process.env.SMITH || 'smith'

// `smith commands --json` carries no default values or value types, so the ones
// that matter are kept here, as `command path|flag-name`. Defaults already written in a
// flag's own description ("(default: ...)") are left in the description. Each entry
// was checked against the flag's declaration in apps/smith.
const defaults = {
  'add|lang': "the stack's, or the workspace's only language",
  'add|stack': 'the recipe, else `[default_stacks]` in `smith.toml`',
  'adopt|level': '`conservative`',
  'affected|base': 'the pull request target, else the release branch',
  'graph|base': 'the pull request target, else the release branch',
  'graph|format': '`text`',
  'convert|build': '`smith`',
  'convert|ci': '`smith`',
  'convert|level': '`conservative`',
  'convert|lint': '`smith` when the ESLint config imports a shared config smith replaces, else `keep`',
  'convert|test': 'the runner the workspace uses, else `keep`',
  'new|ci': '`github`',
  'new|dir': '`./<name>`',
  'new|module-prefix': 'from the `origin` remote, else `example.com/<name>`',
  'new|scope': '`@<name>`',
  'pipeline|ci': "the workspace's providers",
  'release|format': '`text`',
  'slices check|lang': '`go`',
  'cache prune|max-size-mb': '`[cache] max_size_mb`'
}
for (const run of ['run', 'build', 'lint', 'test', 'typecheck']) {
  defaults[`${run}|base`] = 'the pull request target, else the release branch'
  defaults[`${run}|parallel`] = 'the number of CPUs'
}

// Value types that are not simply "string".
const types = {
  'build|parallel': 'integer',
  'lint|parallel': 'integer',
  'test|parallel': 'integer',
  'typecheck|parallel': 'integer',
  'run|parallel': 'integer',
  'doctor|choose': 'integers, comma-separated',
  'adopt|level': '`conservative` or `full`',
  'convert|level': '`conservative` or `full`',
  'convert|lint': '`smith` or `keep`',
  'convert|test': '`jest`, `vitest` or `keep`',
  'convert|build': '`smith` or `keep`',
  'convert|ci': '`smith` or `keep`',
  'new|ci': 'string, repeatable: `github`, `azure`, `gitlab`',
  'pipeline|ci': '`github`, `azure` or `both`',
  'graph|format': '`text`, `json`, `dot`, `html` or `tui`',
  'release|format': '`text` or `markdown`',
  'graph|blast-radius': 'project or file path',
  'graph|diff': 'revision',
  'add|choice': 'name=answer, repeatable',
  'recipe check|choice': 'name=answer, repeatable',
  'slices check|lang': 'language name',
  'cache prune|max-size-mb': 'integer (megabytes)'
}

// Say which binary answered: a stale smith earlier on the PATH (an old install in
// ~/.local/bin, say) silently gives old flags and no subcommands.
console.log(`using ${execFileSync(smith, ['--version'], { encoding: 'utf8' }).trim()} (${smith})`)

const commands = JSON.parse(execFileSync(smith, ['commands', '--json'], { encoding: 'utf8', maxBuffer: 16 << 20 }))

// Every command path, top-level and nested, with its description.
const byPath = new Map()
const paths = []
;(function walk (list, prefix) {
  for (const c of list) {
    const path = prefix ? `${prefix} ${c.name}` : c.name
    byPath.set(path, c)
    paths.push(path)
    walk(c.subcommands ?? [], path)
  }
})(commands, '')

const escapeCell = (s) => s.replace(/\|/g, '\\|')

// `smith commands --json` does not carry everything a usage line says: it lists
// why-miss and explain with separate arguments (the commands take project:task),
// and drops the "..." of some variadic arguments. These usage lines are the ones
// each command prints in its own `--help`.
const usageLines = {
  'why-miss': 'smith why-miss <project>:<task> [flags]',
  explain: 'smith explain <project>[:<task>] [flags]',
  build: 'smith build [project]... [flags]',
  test: 'smith test [project]... [flags]',
  lint: 'smith lint [project]... [flags]',
  typecheck: 'smith typecheck [project]... [flags]',
  'plugin enable': 'smith plugin enable <name>...',
  'plugin disable': 'smith plugin disable <name>...',
  'plugin set': 'smith plugin set <name>...',
  'plugin check': 'smith plugin check <name | path to plugin> [args...] [flags]',
  'slices check': 'smith slices check <project directory>... [flags]'
}

// What each positional argument means, as `command path|argument name`.
const argumentMeanings = {
  'new|name': 'the workspace name, in kebab-case',
  'add|kind': '`app`, `lib` or `package`, or a combined kind such as `go-lib` (`smith kinds`)',
  'add|name': "the project's name, in kebab-case",
  'remove|project': 'the project to remove',
  'run|task': 'the tasks to run, in order, such as `lint test`',
  'build|project': 'projects to run (the same as `-p`)',
  'test|project': 'projects to run (the same as `-p`)',
  'lint|project': 'projects to run (the same as `-p`)',
  'typecheck|project': 'projects to run (the same as `-p`)',
  'dev|projects': 'the projects to start; each needs a `start` task',
  'install|packages': 'packages to add; none installs the whole workspace',
  'uninstall|packages': 'packages to remove',
  'up|packages': 'packages to update; none lists the updates',
  'deps up|packages': 'packages to update; none lists the updates',
  'config get|setting': 'the setting, such as `ci.providers`',
  'config set|setting': 'the setting, such as `ci.providers`',
  'config set|value': 'the new value; a list setting takes several',
  'config unset|setting': 'the setting to remove',
  'plugin enable|name': 'built-in plugin names',
  'plugin disable|name': 'built-in plugin names',
  'plugin set|name': 'built-in plugin names',
  'plugin check|name | path to plugin': 'a plugin by name, or the path of its executable (further arguments are passed to it)',
  'recipe check|file-or-url': 'the recipe file or https URL',
  'recipe new|path': 'where to write the recipe; default `<name>.recipe.toml`',
  'slices check|project directory': 'directories of projects of one language',
  'pipeline|template': 'the template to add; none lists them'
}

const argumentOverrides = {
  'why-miss': [{ name: 'project:task', required: true, variadic: false, description: 'the task to explain, such as web:build' }],
  explain: [{ name: 'project[:task]', required: true, variadic: false, description: 'the project, optionally with one of its tasks, such as web:build' }]
}

function usage (path, c) {
  if (usageLines[path]) return usageLines[path]
  const args = c.arguments.map((a) => {
    const name = a.variadic ? `${a.name}...` : a.name
    return a.required ? `<${name}>` : `[${name}]`
  })
  const sub = (c.subcommands ?? []).length ? ['<command>'] : []
  return ['smith', path, ...sub, ...args, ...(c.options.length ? ['[flags]'] : [])].join(' ')
}

function defaultOf (path, o) {
  const key = `${path}|${o.name}`
  const inline = / \(default[: ]([^)]*)\)/.exec(o.description)
  if (defaults[key]) return defaults[key]
  if (inline) return inline[1].replace(/^"|"$/g, '`')
  return o.takesValue === 'none' ? 'off' : ''
}

function typeOf (path, o) {
  if (o.takesValue === 'none') return 'boolean'
  const key = `${path}|${o.name}`
  if (types[key]) return types[key]
  return o.repeatable ? 'string, repeatable' : 'string'
}

// compact: the three-column table of docs/reference/cli.md.
function block (path, c, compact) {
  const lines = [c.description, '', `Usage: \`${usage(path, c)}\``]
  if (c.aliases?.length) lines.push('', `Alias: ${c.aliases.map((a) => `\`smith ${a}\``).join(', ')}`)
  const args = argumentOverrides[path] ?? c.arguments
  if (!compact && args.length) {
    lines.push('', '| Argument | Required | Meaning |', '|---|---|---|')
    for (const a of args) {
      lines.push(`| \`${escapeCell(a.variadic ? `${a.name}...` : a.name)}\` | ${a.required ? 'yes' : 'no'} | ${escapeCell(argumentMeanings[`${path}|${a.name}`] ?? (a.description && a.description !== a.name ? a.description : ''))} |`)
    }
  }
  if ((c.subcommands ?? []).length) {
    lines.push('', '| Subcommand | Meaning |', '|---|---|')
    for (const s of c.subcommands) lines.push(`| \`smith ${path} ${s.name}\` | ${escapeCell(s.description)} |`)
  }
  if (c.options.length) {
    if (compact) {
      lines.push('', '| Flag | Meaning | Default |', '|---|---|---|')
      for (const o of c.options) {
        const meaning = o.description + (o.repeatable ? ' (repeatable)' : '')
        lines.push(`| \`${o.flags}\` | ${escapeCell(meaning)} | ${escapeCell(defaultOf(path, o))} |`)
      }
    } else {
      lines.push('', '| Flag | Type | Default | Effect |', '|---|---|---|---|')
      for (const o of c.options) {
        lines.push(`| \`${o.flags}\` | ${escapeCell(typeOf(path, o))} | ${escapeCell(defaultOf(path, o))} | ${escapeCell(o.description)} |`)
      }
    }
  } else {
    lines.push('', 'No flags.')
  }
  return lines.join('\n')
}

function markdownFiles (dir) {
  if (!existsSync(dir)) return []
  return readdirSync(dir).filter((n) => n.endsWith('.md')).map((n) => join(dir, n))
}

const problems = []
const seenIn = { cli: new Set(), manual: new Set() }
let stale = 0
let blocks = 0
const targets = [
  { file: join(docs, 'reference', 'cli.md'), compact: true, seen: seenIn.cli },
  ...markdownFiles(join(docs, 'manual')).map((file) => ({ file, compact: false, seen: seenIn.manual }))
]
const re = /<!-- smith:begin ([\w -]+) -->\n[\s\S]*?<!-- smith:end \1 -->/g
for (const t of targets) {
  const original = readFileSync(t.file, 'utf8')
  const where = relative(docs, t.file).replace(/\\/g, '/')
  const updated = original.replace(re, (whole, path) => {
    const c = byPath.get(path)
    if (!c) {
      problems.push(`${where} has a block for "${path}", which smith does not have`)
      return whole
    }
    if (t.compact && path.includes(' ')) {
      problems.push(`${where} has a block for "${path}": cli.md carries top-level commands only`)
      return whole
    }
    t.seen.add(path)
    blocks++
    return `<!-- smith:begin ${path} -->\n${block(path, c, t.compact)}\n\n<!-- smith:end ${path} -->`
  })
  if (updated !== original) {
    stale++
    if (check) problems.push(`${where} is out of date: run node docs/reference/cli-generate.mjs`)
    else writeFileSync(t.file, updated)
  }
}
for (const c of commands) if (!seenIn.cli.has(c.name)) problems.push(`smith ${c.name} has no block in reference/cli.md`)
if (existsSync(join(docs, 'manual'))) {
  for (const p of paths) if (!seenIn.manual.has(p)) problems.push(`smith ${p} has no block in docs/manual`)
}

for (const p of problems) console.error(p)
if (check) {
  if (!problems.length) console.log(`${blocks} blocks are up to date`)
  process.exit(problems.length ? 1 : 0)
} else {
  console.log(`updated ${stale} pages, ${blocks} blocks`)
  if (problems.length) process.exit(1)
}
