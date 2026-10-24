import type { ListNode, ListResult } from '../list-view/index.ts'
import type { ReleasePreview, ReleaseResult } from './release-report.contract.ts'

const LEVEL_COLORS: Record<string, string> = {
  major: 'charts.red',
  minor: 'charts.yellow',
  patch: 'charts.green',
}

/**
 * Lays the release plan out as rows: a row per package that would release,
 * with the version it moves to and under it why, its tag and the commits that
 * cause it, then the packages with nothing to release, closed.
 * Publishing is not offered here: it is `smith release`, which asks first.
 */
export function toReleaseRows (preview: ReleasePreview): ListResult {
  const results = preview.results ?? []
  const unchanged = preview.unchanged ?? []
  const nodes: ListNode[] = [
    ...(preview.error === undefined ? [] : [{ id: 'release:error', label: preview.error, icon: 'error', color: 'charts.red' }]),
    ...results.map((result) => toResultRow(result)),
    ...(unchanged.length === 0
      ? []
      : [{
          id:          'release:unchanged',
          label:       'Nothing to release',
          description: String(unchanged.length),
          icon:        'circle-slash',
          children:    unchanged.map((name) => ({ id: `release:unchanged:${name}`, label: name, icon: 'package' })),
        }]),
  ]
  if (results.length === 0) {
    return { nodes, message: preview.error === undefined ? 'Nothing to release.' : undefined, summary: 'nothing to release' }
  }

  return { nodes, summary: `${results.length} to release` }
}

function toResultRow (result: ReleaseResult): ListNode {
  const commits = result.commits ?? []
  const from = result.current === undefined || result.current === '' ? 'new' : result.current
  const level = result.level ?? ''
  const children: ListNode[] = [
    ...(result.reason === undefined || result.reason === '' ? [] : [{ id: `release:${result.project}:reason`, label: result.reason, icon: 'info', tooltip: result.reason }]),
    ...(result.tag === undefined || result.tag === '' ? [] : [{ id: `release:${result.project}:tag`, label: result.tag, icon: 'tag', tooltip: 'The tag the release creates.' }]),
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
    tooltip:     [`${result.project} ${from} → ${result.next}`, result.reason, result.forced === true ? 'Forced.' : undefined].filter(Boolean).join('\n'),
    children,
  }
}
