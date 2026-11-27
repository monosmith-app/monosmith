import type { ListNode, ListResult } from '../../shared/list-view/index.ts'
import type { ReleaseCheck } from './release-checks.policy.ts'
import type { ReleasePreview, ReleaseResult } from './release-report.contract.ts'
import { say, sayCount } from '../../shared/words/index.ts'

const LEVEL_COLORS: Record<string, string> = {
  major: 'charts.red',
  minor: 'charts.yellow',
  patch: 'charts.green',
}

/**
 * Lays the release plan out as the terminal's Release view does: a row per package that would
 * release, with the version it moves to and, under it, its detail (why it releases, its changes,
 * its tag and the command it stands for); the packages with nothing to release, closed; then the
 * checks before release and the actions, each with its smith command. Nothing is published from
 * a row: `Release now` is `MonoSmith: Release…`, which shows the plan and asks first.
 * @param preview - the plan of `smith release --dry-run`
 * @param explanation - what the CLI says when nothing releases (see `explanationLines`), shown as rows
 * @param checks - the checks before release; none are shown when absent
 * @returns the rows, with the headline beside the title
 */
export function toReleaseRows (preview: ReleasePreview, explanation: readonly string[] = [], checks?: readonly ReleaseCheck[]): ListResult {
  const results = preview.results ?? []
  const unchanged = preview.unchanged ?? []
  const channel = preview.channel === undefined || preview.channel === '' ? undefined : preview.channel
  const nodes: ListNode[] = [
    ...(channel === undefined ? [] : [toChannelRow(channel)]),
    ...(preview.error === undefined ? [] : [{ id: 'release:error', label: preview.error, icon: 'error', color: 'charts.red' }]),
    ...results.map((result) => toResultRow(result)),
    ...(unchanged.length === 0
      ? []
      : [{
          id:          'release:unchanged',
          label:       say('vscode.release-preview.unchanged'),
          description: String(unchanged.length),
          icon:        'circle-slash',
          children:    unchanged.map((name) => ({ id: `release:unchanged:${name}`, label: name, icon: 'package' })),
        }]),
    ...(results.length === 0 ? [] : actionRows(checks)),
  ]
  const suffix = channel === undefined ? '' : ` · channel ${channel}`
  if (results.length === 0) {
    const why: ListNode[] = explanation.length === 0
      ? []
      : [{ id: 'release:why', label: say('vscode.release-preview.why-nothing-releases'), icon: 'info', expanded: true, children: explanation.map((line, index) => ({ id: `release:why:${index}`, label: line, icon: 'dash', tooltip: line })) }]

    return { nodes: [...why, ...nodes], message: preview.error === undefined ? say('vscode.release-flow.nothing-to-release') : undefined, summary: say('vscode.release-preview.nothing-to-release', { suffix }) }
  }

  return { nodes, summary: `${headline(results)}${suffix}` }
}

/** `✓ 2 packages to release · 1 minor · 1 patch`. */
function headline (results: readonly ReleaseResult[]): string {
  const levels = (['major', 'minor', 'patch'] as const)
    .map((level) => ({ level, count: results.filter((result) => result.level === level).length }))
    .filter(({ count }) => count > 0)
    .map(({ level, count }) => say(`dashboard.box.release.${level}`, { count }))

  return [`✓ ${sayCount('vscode.release-preview.to-release', results.length)}`, ...levels].join(' · ')
}

/** The checks before release, the note that nothing is published unasked, and the three actions with their commands. */
function actionRows (checks: readonly ReleaseCheck[] | undefined): ListNode[] {
  const action = (id: string, label: string, command: string, line: string, icon: string): ListNode => ({ id: `release:action:${id}`, label, description: `$ ${line}`, icon, command, tooltip: line })

  return [
    ...(checks === undefined
      ? []
      : [{
          id:       'release:checks',
          label:    say('releasepage.checks.title'),
          icon:     checks.some((check) => check.warn) ? 'warning' : 'pass',
          color:    checks.some((check) => check.warn) ? 'charts.yellow' : 'charts.green',
          expanded: true,
          children: [
            ...checks.map((check, index) => ({ id: `release:check:${index}`, label: check.text, icon: check.warn ? 'warning' : 'pass', color: check.warn ? 'charts.yellow' : 'charts.green' })),
            { id: 'release:check:registry', label: say('releasepage.checks.registry'), icon: 'circle-outline' },
          ],
        }]),
    { id: 'release:note', label: say('releasepage.confirm_note'), icon: 'info' },
    action('release', say('vscode.release-preview.release-now'), 'monosmith.release', 'smith release', 'rocket'),
    action('dry-run', say('releasepage.button.preview'), 'monosmith.release.dryRun', 'smith release --dry-run', 'eye'),
    action('notes', say('releasepage.button.notes'), 'monosmith.release.notes', 'smith release --dry-run --notes', 'note'),
  ]
}

/** The notice `smith release` prints for a channel, worded as smith words it. */
function channelNotice (channel: string): string {
  return say('vscode.release-preview.release-channel-prerelease-versions-published-under', { channel, channel2: channel })
}

function toChannelRow (channel: string): ListNode {
  return {
    id:          'release:channel',
    label:       say('vscode.release-preview.release-channel', { channel }),
    description: say('vscode.release-preview.prerelease-versions'),
    icon:        'beaker',
    color:       'charts.yellow',
    tooltip:     channelNotice(channel),
    children:    [{ id: 'release:channel:notice', label: channelNotice(channel), icon: 'info', tooltip: channelNotice(channel) }],
    expanded:    true,
  }
}

function toResultRow (result: ReleaseResult): ListNode {
  const commits = result.commits ?? []
  const from = result.current === undefined || result.current === '' ? 'new' : result.current
  const level = result.level ?? ''
  const id = `release:${result.project}`
  const children: ListNode[] = [
    ...(result.plugin === undefined || result.plugin === '' ? [] : [{ id: `${id}:plugin`, label: say('releasepage.detail.plugin', { plugin: result.plugin }), icon: 'extensions' }]),
    ...(result.reason === undefined || result.reason === '' ? [] : [{ id: `${id}:reason`, label: result.reason, description: say('releasepage.detail.why'), icon: 'info', tooltip: result.reason }]),
    ...(commits.length === 0
      ? []
      : [{
          id:          `${id}:changes`,
          label:       say('releasepage.detail.changes'),
          description: String(commits.length),
          icon:        'git-commit',
          expanded:    commits.length <= 5,
          children:    commits.map((commit, index) => ({ id: `${id}:commit:${index}`, label: commit, icon: 'git-commit', tooltip: commit })),
        }]),
    ...(result.tag === undefined || result.tag === '' ? [] : [{ id: `${id}:tag`, label: result.tag, description: say('releasepage.detail.tag'), icon: 'tag', tooltip: say('vscode.release-preview.the-tag-the-release-creates') }]),
    ...(result.error === undefined ? [] : [{ id: `${id}:error`, label: result.error, icon: 'error', color: 'charts.red' }]),
    { id: `${id}:same`, label: say('releasepage.detail.same_as'), description: say('releasepage.detail.command'), icon: 'terminal', command: 'monosmith.release.notes' },
  ]

  return {
    id,
    label:       result.project,
    description: `${from} → ${result.next}${level === '' ? '' : ` · ${level}`}`,
    icon:        'rocket',
    ...(LEVEL_COLORS[level] !== undefined && { color: LEVEL_COLORS[level] }),
    expanded:    children.length <= 5,
    tooltip:     [`${result.project} ${from} → ${result.next}`, result.reason, result.forced === true ? say('vscode.release-preview.forced') : undefined].filter(Boolean).join('\n'),
    children,
  }
}

/**
 * Picks the explanation out of the text `smith release --dry-run` prints when nothing releases:
 * which commits count, that apps and libraries never release, and the way to a first release.
 * @param text - the text smith printed
 * @returns the lines, without the dry-run notice, the arrows and the indent; empty when smith explained nothing
 */
export function explanationLines (text: string): string[] {
  return text.split(/\r?\n/)
    .map((line) => line.replace(/^\s*→\s*/, '').trim())
    .filter((line) => line !== '' && !line.startsWith(say('vscode.release-preview.dry-run')))
}
