import type { ComposeCheck, ComposeItem, ComposeLayer, ComposePlan, ComposeQuestion, ComposeReport, ComposeTool } from './compose-stack.contract.ts'

type Json = Record<string, unknown>

function record (value: unknown): Json {
  return typeof value === 'object' && value !== null && !Array.isArray(value) ? value as Json : {}
}

function records (value: unknown): Json[] {
  return Array.isArray(value) ? value.filter((item): item is Json => typeof item === 'object' && item !== null) : []
}

function text (value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function words (value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string') : []
}

function readQuestion (value: unknown): ComposeQuestion | undefined {
  const q = record(value)
  if (text(q.axis) === '') return undefined

  return {
    choice:  text(q.choice),
    about:   text(q.about),
    axis:    text(q.axis),
    options: records(q.options).map((o) => ({ name: text(o.name), title: text(o.title) || text(o.name), description: text(o.description), suggested: o.suggested === true })),
  }
}

function readCheck (value: unknown): ComposeCheck | undefined {
  const c = record(value)
  if (Object.keys(c).length === 0) return undefined
  const question = readQuestion(c.question)

  return {
    language:  text(c.language),
    complete:  c.complete === true,
    conflict:  text(c.conflict),
    missing:   text(c.missing),
    ...(question !== undefined && { question }),
    selection: words(c.selection),
    added:     words(c.added),
  }
}

function readPlan (value: unknown): ComposePlan | undefined {
  const p = record(value)
  if (text(p.command) === '' && words(p.notes).length === 0) return undefined

  return {
    command:    text(p.command),
    argv:       words(p.argv),
    files:      records(p.files).map((f) => ({ action: text(f.action), path: text(f.path), note: text(f.note) })),
    tools:      records(p.tools).map((t) => ({ choice: text(t.choice), name: text(t.name), title: text(t.title) || text(t.name), binary: text(t.binary), state: text(t.state), command: text(t.command) })),
    notes:      words(p.notes),
    recipeArgv: words(p.recipeArgv),
    recipeNote: text(p.recipeNote),
  }
}

/**
 * Reads `smith compose --json`.
 * @param json - what smith printed
 * @returns the layers, the items that have a name, and the check and plan when asked for
 * @throws SyntaxError when the text is not JSON, TypeError when it is not an object
 */
export function parseReport (json: string): ComposeReport {
  const value: unknown = JSON.parse(json)
  if (typeof value !== 'object' || value === null || Array.isArray(value)) throw new TypeError('smith printed no catalogue')
  const root = value as Json
  const layers: ComposeLayer[] = records(root.layers).map((l) => ({ id: text(l.id), title: text(l.title), about: text(l.about) }))
  const items: ComposeItem[] = records(root.items)
    .map((i) => ({
      id:          text(i.id),
      kind:        text(i.kind),
      name:        text(i.name),
      title:       text(i.title) || text(i.name),
      description: text(i.description),
      aliases:     words(i.aliases),
      layer:       text(i.layer),
      axis:        text(i.axis),
      many:        i.many === true,
      language:    text(i.language),
      suggested:   i.suggested === true,
      isDefault:   i.default === true,
      problem:     text(i.problem),
      route:       text(i.route),
    }))
    .filter((i) => i.id !== '' && i.name !== '')
  const check = readCheck(root.check)
  const plan = readPlan(root.plan)

  return { layers, items, ...(check !== undefined && { check }), ...(plan !== undefined && { plan }) }
}

/**
 * The arguments of `smith compose --json` for the picks so far.
 * @param request - the language and choice names ticked, and the project to plan for
 * @returns the arguments after `smith`; no `--with` when nothing is ticked
 */
export function composeArguments (request: { language?: string, picks: readonly string[], kind?: string, name?: string }): string[] {
  const args = ['compose', '--json']
  if (request.language !== undefined && request.language !== '') args.push('--lang', request.language)
  if (request.picks.length > 0) args.push('--with', request.picks.join(','))
  if (request.kind !== undefined) args.push('--kind', request.kind)
  if (request.name !== undefined) args.push('--name', request.name)

  return args
}

/**
 * One line for a tool of the plan, as the TUI words it: the name, whether it is
 * there and which pick needs it, and the command that installs it when it is not.
 * @param tool - a tool of the plan
 * @returns "Node.js (installed, for node)" or "the Tauri system libraries (missing, for tauri): smith tools install tauri-linux-libs"
 */
export function describeTool (tool: ComposeTool): string {
  const state = tool.state === '' ? 'not checked' : tool.state
  const line = `${tool.title} (${state}${tool.choice === '' ? '' : `, for ${tool.choice}`})`

  return tool.state === 'missing' || tool.state === 'outdated' ? `${line}: ${tool.command}` : line
}

/**
 * The tools of a plan that are not there or too old.
 * @param plan - the plan smith printed
 * @returns the missing and outdated tools, in the order of the plan
 */
export function toolsToInstall (plan: ComposePlan): ComposeTool[] {
  return plan.tools.filter((tool) => tool.state === 'missing' || tool.state === 'outdated')
}

/**
 * Words the plan for the confirmation: the command, the files, the tools and the notes.
 * @param plan - the plan smith printed
 * @returns the text shown in the dialog
 */
export function describePlan (plan: ComposePlan): string {
  const lines = [`Run: ${plan.command}`, '']
  if (plan.files.length > 0) {
    lines.push('It writes:')
    for (const f of plan.files) lines.push(`  ${f.action} ${f.path}${f.note === '' ? '' : ` (${f.note})`}`)
    lines.push('')
  }
  if (plan.tools.length > 0) {
    lines.push('It needs these tools on this computer:')
    for (const t of plan.tools) lines.push(`  ${describeTool(t)}`)
    lines.push('')
  }
  lines.push(...plan.notes)

  return lines.join('\n').trim()
}

/**
 * The lines a quick pick shows for an item: title, where it is for, what it is.
 * @param item - one catalogue item
 * @returns the label, description and detail of the pick; detail holds the aliases, so typing "desktop" finds Electron
 */
export function pickLabels (item: ComposeItem): { label: string, description: string, detail: string } {
  const where = item.language === '' || item.kind === 'language' ? '' : `${item.language} · `
  // Electron sits on the shell layer but answers the framework question; say so, so the conflict message reads the same.
  const slot = item.kind === 'choice' && item.axis !== '' && item.axis !== item.layer && !item.many ? ` (the ${item.axis} choice)` : ''
  const mark = item.isDefault ? ' (default)' : (item.suggested ? ' (suggested)' : '')
  const problem = item.problem === '' ? '' : ` Cannot be used: ${item.problem}.`
  const route = item.route === '' ? '' : ` Not part of one project: turn it on with ${item.route}.`

  return {
    label:       `${item.title}${mark}`,
    description: `${where}${item.layer}${slot}`,
    detail:      `${item.description}${problem}${route} ${item.aliases.join(' ')}`.trim(),
  }
}

/**
 * The titles of the ticked items, in catalogue order, to show what is chosen
 * when a search hides the rows.
 * @param items - the catalogue
 * @param ids - the ids ticked
 * @returns "React, Electron", or '' when nothing is ticked
 */
export function tickedTitles (items: readonly ComposeItem[], ids: readonly string[]): string {
  return items.filter((item) => ids.includes(item.id)).map((item) => item.title).join(', ')
}

/**
 * The stack in words, to name it where a question about it is asked.
 * @param items - the catalogue
 * @param language - the language the picks fix
 * @param picks - the names of the choices ticked
 * @returns "React + Electron (node)"
 */
export function stackSummary (items: readonly ComposeItem[], language: string, picks: readonly string[]): string {
  const titles = picks.map((name) => items.find((item) => item.kind === 'choice' && item.name === name && item.language === language)?.title ?? name)
  const lang = items.find((item) => item.kind === 'language' && item.name === language)?.title ?? language

  return `${titles.join(' + ')}${lang === '' ? '' : ` (${lang})`}`
}

/**
 * Keeps one pick per single-choice axis, like radio buttons: a choice ticked in
 * this round takes the place of a choice ticked before on the same axis and
 * language, unless both allow several. The catalogue decides the axes; no name
 * is known here.
 * @param items - the catalogue
 * @param previous - the ids ticked before this round
 * @param chosen - the ids ticked now
 * @returns the ids to keep, and a one-line note naming what was replaced (empty when nothing was)
 */
export function settleRivals (items: readonly ComposeItem[], previous: readonly string[], chosen: readonly string[]): { ids: string[], note: string } {
  const byId = new Map(items.map((item) => [item.id, item]))
  const fresh = chosen.filter((id) => !previous.includes(id))
  const dropped = new Set<string>()
  const notes: string[] = []
  for (const id of fresh) {
    const pick = byId.get(id)
    const isLanguage = pick?.kind === 'language'
    if (pick === undefined || (!isLanguage && (pick.kind !== 'choice' || pick.axis === ''))) continue
    const rivals = chosen.filter((other) => {
      const item = byId.get(other)
      const later = fresh.slice(fresh.indexOf(id) + 1).includes(other)
      if (other === id || later || item === undefined || dropped.has(other)) return false

      return isLanguage ? item.kind === 'language' : item.kind === 'choice' && item.axis === pick.axis && item.language === pick.language && !(item.many && pick.many)
    })
    for (const rival of rivals) dropped.add(rival)
    if (rivals.length > 0) notes.push(`${pick.title} replaces ${rivals.map((other) => byId.get(other)?.title ?? other).join(' and ')}: a project has one ${isLanguage ? 'language' : pick.axis}.`)
  }

  return { ids: chosen.filter((id) => !dropped.has(id)), note: notes.join(' ') }
}
