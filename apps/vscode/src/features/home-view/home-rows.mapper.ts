import type { ListNode, ListResult } from '../../shared/list-view/index.ts'
import { commandLineOf } from '../command-palette/index.ts'
import type { HomeButton, HomeFacts, HomeItem, HomeSeverity } from './home-facts.contract.ts'
import { homeQueue } from './home-queue.policy.ts'
import { say, sayCount, type WordKey } from '../../shared/words/index.ts'

/** The mark and colour of each severity. */
const MARKS: Record<HomeSeverity, { icon: string, color?: string }> = {
  failure: { icon: 'error', color: 'charts.red' },
  warning: { icon: 'warning', color: 'charts.yellow' },
  info:    { icon: 'info', color: 'charts.blue' },
  muted:   { icon: 'circle-outline' },
  success: { icon: 'pass', color: 'charts.green' },
}

/** The tag of each kind of item, as the terminal's Home shows it. */
const TAGS: Record<HomeItem['kind'], WordKey> = {
  empty:    'dashboard.tag.start',
  checking: 'dashboard.tag.checking',
  doctor:   'dashboard.tag.failed',
  run:      'dashboard.tag.failed',
  setup:    'dashboard.tag.setup',
  warnings: 'dashboard.tag.warning',
  changes:  'dashboard.tag.changed',
  release:  'dashboard.tag.release',
  updates:  'dashboard.tag.updates',
  good:     'dashboard.tag.good',
}

/**
 * Lays Home out as the terminal's Home does, in a tree: the four boxes (Changed, Last run, Health,
 * Release), then **Needs you now** with each item and its buttons, then **Start something**. Choosing
 * an item runs its first button; each button shows the smith command line it stands for.
 * @param facts - what the extension holds now
 * @returns the rows, with the workspace line beside the title
 */
export function toHomeRows (facts: HomeFacts): ListResult {
  const queue = homeQueue(facts)
  const empty = facts.projectsLoaded && facts.projects === 0

  return {
    nodes: [
      ...boxes(facts),
      {
        id:          'home:queue',
        label:       say('dashboard.list.title'),
        description: sayCount('vscode.home-view.items', queue.length),
        icon:        'bell',
        expanded:    true,
        children:    queue.map((item, index) => itemRow(item, index)),
      },
      ...(empty ? [] : [startRow()]),
    ],
    summary: topLine(facts),
  }
}

/** `12 projects (Go, TypeScript) · 3 changed · last run 3 min ago`: the parts that have facts. */
function topLine (facts: HomeFacts): string {
  const languages = [...new Set(facts.languages.filter(Boolean))].sort((a, b) => a.localeCompare(b))
  const projects = sayCount('dashboard.project', facts.projects)

  return [
    languages.length === 0 ? projects : say('dashboard.projects.summary', { projects, languages: languages.join(', ') }),
    facts.affected > 0 ? say('dashboard.top.changed', { count: facts.affected }) : '',
    facts.last === undefined || facts.last.ago === '' ? '' : say('dashboard.top.last_run', { ago: facts.last.ago }),
  ].filter(Boolean).join(' · ')
}

/** One box of the four: its name, and its value beside it with the mark of its state. */
function box (id: string, label: WordKey, value: string, severity: HomeSeverity, command: string, sub = ''): ListNode {
  const mark = MARKS[severity]

  return { id: `home:box:${id}`, label: say(label), description: [value, sub].filter(Boolean).join(' · '), icon: mark.icon, ...(mark.color !== undefined && { color: mark.color }), command }
}

function boxes (facts: HomeFacts): ListNode[] {
  return [changedBox(facts), runBox(facts), healthBox(facts), releaseBox(facts)]
}

function changedBox (facts: HomeFacts): ListNode {
  const open = 'monosmith.projects.focus'
  if (!facts.projectsLoaded) return box('changed', 'dashboard.box.changed', say('dashboard.box.loading'), 'muted', open)
  if (facts.projects === 0) return box('changed', 'dashboard.box.changed', say('dashboard.box.changed.empty'), 'muted', open)
  if (facts.affected === 0) return box('changed', 'dashboard.box.changed', say('dashboard.box.changed.none'), 'success', open)

  return box('changed', 'dashboard.box.changed', sayCount('dashboard.project', facts.affected), 'info', open)
}

function runBox (facts: HomeFacts): ListNode {
  const last = facts.last
  if (last === undefined) return box('run', 'dashboard.box.run', say('dashboard.box.run.none'), 'muted', 'monosmith.runs.focus')
  const tally = [
    last.passed > 0 ? sayCount('dashboard.box.run.passed', last.passed) : '',
    last.cached > 0 ? sayCount('dashboard.box.run.cached', last.cached) : '',
  ].filter(Boolean).join(' · ')
  if (last.running) return box('run', 'dashboard.box.run', say('runpage.state.running', { what: last.command }), 'info', 'monosmith.runs.open')
  if (last.stopped) return box('run', 'dashboard.box.run', say('dashboard.box.run.stopped'), 'muted', 'monosmith.runs.open', tally)
  if (!last.ok) return box('run', 'dashboard.box.run', last.failed > 0 ? sayCount('dashboard.box.run.failed', last.failed) : say('dashboard.box.run.failed_plain'), 'failure', 'monosmith.runs.open', tally)

  return box('run', 'dashboard.box.run', tally === '' ? say('dashboard.box.run.passed_plain') : tally, 'success', 'monosmith.runs.open')
}

function healthBox (facts: HomeFacts): ListNode {
  if (facts.healthError !== undefined) return box('health', 'dashboard.box.health', say('dashboard.box.health.error'), 'failure', 'monosmith.health.focus')
  const health = facts.health
  if (health === undefined) return box('health', 'dashboard.box.health', say('dashboard.box.loading'), 'muted', 'monosmith.health.focus')
  if (health.failed > 0) return box('health', 'dashboard.box.health', sayCount('dashboard.box.health.failed', health.failed), 'failure', 'monosmith.health.focus', health.warnings > 0 ? sayCount('dashboard.box.health.warnings', health.warnings) : '')
  if (health.warnings > 0) return box('health', 'dashboard.box.health', sayCount('dashboard.box.health.warnings', health.warnings), 'warning', 'monosmith.health.focus', sayCount('dashboard.box.health.failed_checks', 0))

  return box('health', 'dashboard.box.health', say('dashboard.box.health.pass'), 'success', 'monosmith.health.focus', sayCount('dashboard.box.health.passed', health.passed))
}

function releaseBox (facts: HomeFacts): ListNode {
  if (!facts.releaseKnown) return box('release', 'dashboard.box.release', say('dashboard.box.loading'), 'muted', 'monosmith.release.focus')
  const release = facts.release
  if (release === undefined) return box('release', 'dashboard.box.release', say('dashboard.box.release.none'), 'muted', 'monosmith.release.focus')
  const levels = ([['major', release.major], ['minor', release.minor], ['patch', release.patch]] as const)
    .filter(([, count]) => count > 0)
    .map(([level, count]) => say(`dashboard.box.release.${level}`, { count }))

  return box('release', 'dashboard.box.release', sayCount('dashboard.box.release.ready', release.packages), 'success', 'monosmith.release.focus', levels.join(' · '))
}

/** A button as a row: its label, and the command line it stands for beside it. */
function buttonRow (button: HomeButton, id: string): ListNode {
  return { id, label: button.label, ...(button.line !== undefined && { description: `$ ${button.line}`, tooltip: button.line }), icon: 'play', command: button.command }
}

function itemRow (item: HomeItem, index: number): ListNode {
  const mark = MARKS[item.severity]
  const id = `home:item:${index}:${item.kind}`
  const [primary] = item.buttons

  return {
    id,
    label:       item.title,
    description: [say(TAGS[item.kind]), item.detail].filter(Boolean).join(' · '),
    icon:        mark.icon,
    ...(mark.color !== undefined && { color: mark.color }),
    tooltip:     [item.title, item.detail, ...item.buttons.map((button) => `${button.label}${button.line === undefined ? '' : `: $ ${button.line}`}`)].filter(Boolean).join('\n'),
    ...(primary !== undefined && { command: primary.command }),
    // The first item shows its buttons; the others keep them a click away.
    ...(item.buttons.length > 0 && { expanded: index === 0, children: item.buttons.map((button, at) => buttonRow(button, `${id}:button:${at}`)) }),
  }
}

/** Start something: the actions Home offers to start work, each with its smith command line. */
function startRow (): ListNode {
  const starts: Array<[WordKey, string]> = [
    ['dashboard.good.build', 'monosmith.buildAll'],
    ['dashboard.start.lint', 'monosmith.lintAll'],
    ['dashboard.start.task', 'monosmith.runTask'],
    ['dashboard.add.label', 'monosmith.addProject'],
    ['dashboard.compose.label', 'monosmith.composeStack'],
    ['dashboard.start.settings', 'monosmith.settings.focus'],
  ]

  return {
    id:       'home:start',
    label:    say('dashboard.start.title'),
    icon:     'rocket',
    expanded: true,
    children: starts.map(([label, command]) => {
      const line = commandLineOf(command)

      return buttonRow({ label: say(label), command, ...(line !== undefined && { line }) }, `home:start:${command}`)
    }),
  }
}
