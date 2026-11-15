import type { ListNode, ListResult } from '../list-view/index.ts'
import type { ReleasePreview, ReleaseResult } from './release-report.contract.ts'
import { say } from '../words/index.ts'

const LEVEL_COLORS: Record<string, string> = {
  major: 'charts.red',
  minor: 'charts.yellow',
  patch: 'charts.green',
}

/**
 * Lays the release plan out as rows: a row per package that would release,
 * with the version it moves to and under it why, its tag and the commits that
 * cause it, then the packages with nothing to release, closed.
 * Publishing is not offered here: it is `MonoSmith: Release…`, which asks first.
 * @param preview - the plan of `smith release --dry-run`
 * @param explanation - what the CLI says when nothing releases (see `explanationLines`), shown as rows
 */
export function toReleaseRows (preview: ReleasePreview, explanation: readonly string[] = []): ListResult {
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
  ]
  const suffix = channel === undefined ? '' : ` · channel ${channel}`
  if (results.length === 0) {
    const why: ListNode[] = explanation.length === 0
      ? []
      : [{ id: 'release:why', label: say('vscode.release-preview.why-nothing-releases'), icon: 'info', expanded: true, children: explanation.map((line, index) => ({ id: `release:why:${index}`, label: line, icon: 'dash', tooltip: line })) }]

    return { nodes: [...why, ...nodes], message: preview.error === undefined ? say('vscode.release-flow.nothing-to-release') : undefined, summary: say('vscode.release-preview.nothing-to-release', { suffix }) }
  }

  return { nodes, summary: `${results.length} to release${suffix}` }
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
  const children: ListNode[] = [
    ...(result.reason === undefined || result.reason === '' ? [] : [{ id: `release:${result.project}:reason`, label: result.reason, icon: 'info', tooltip: result.reason }]),
    ...(result.tag === undefined || result.tag === '' ? [] : [{ id: `release:${result.project}:tag`, label: result.tag, icon: 'tag', tooltip: say('vscode.release-preview.the-tag-the-release-creates') }]),
    ...commits.map((commit, index) => ({ id: `release:${result.project}:commit:${index}`, label: commit, icon: 'git-commit', tooltip: commit })),
    ...(result.error === undefined ? [] : [{ id: `release:${result.project}:error`, label: result.error, icon: 'error', color: 'charts.red' }]),
  ]

  return {
    id:          `release:${result.project}`,
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
