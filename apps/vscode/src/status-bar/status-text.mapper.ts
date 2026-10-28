import type { ViewState } from '../folder-session/index.ts'

/** What the status bar item shows and says. */
export interface StatusText {
  text:    string
  tooltip: string
  /** Colour the item as a warning: tasks failed, or smith cannot be used. */
  warning: boolean
}

/** What the status bar item depends on. */
export interface StatusInput {
  state:     ViewState
  /** smith's version, when it is running. */
  version?:  string
  /** The number of affected projects, if known. */
  affected?: number
  running:   boolean
  /** Tasks the runs so far left failed. */
  failed:    number
  /** Which smith runs and why, as lines. */
  origin:    string[]
}

const CLICK_TO_CONVERT = 'Click for a menu: convert it, see what would change.'
const CLICK = 'Click for a menu: run affected, doctor, refresh, output, settings.'

/**
 * The status bar text: the smith that is running, how many projects the
 * changes affect when that is known, and how many tasks failed. When smith
 * cannot be used, the reason.
 */
export function describeStatus (input: StatusInput): StatusText {
  switch (input.state) {
    case 'starting': {
      return { text: '$(sync~spin) smith', tooltip: 'MonoSmith: starting smith…', warning: false }
    }
    case 'empty': {
      return { text: '$(tools) smith', tooltip: 'MonoSmith: no open folder has a smith.toml.', warning: false }
    }
    case 'adoptable': {
      return {
        text:    '$(tools) smith',
        tooltip: `MonoSmith: this folder is not a MonoSmith workspace yet. It can be converted.\n${CLICK_TO_CONVERT}`,
        warning: false,
      }
    }
    case 'missing': {
      return { text: '$(circle-slash) smith', tooltip: `MonoSmith: smith was not found.\n${CLICK}`, warning: true }
    }
    case 'untrusted': {
      return { text: '$(shield) smith', tooltip: `MonoSmith: this workspace is not trusted, so its own smith is not used.\n${CLICK}`, warning: true }
    }
    case 'failed': {
      return { text: '$(error) smith', tooltip: `MonoSmith: smith could not start or has stopped. See the output.\n${CLICK}`, warning: true }
    }
    case 'ready': {
      const parts = [`smith ${input.version ?? ''}`.trim()]
      if (input.affected !== undefined) parts.push(`${input.affected} affected`)
      if (input.failed > 0) parts.push(`${input.failed} failed`)
      const lines = [...input.origin]
      if (input.running) lines.push('A run is in progress.')
      if (input.failed > 0) lines.push(`${input.failed} ${input.failed === 1 ? 'task' : 'tasks'} failed in the last runs.`)
      lines.push(CLICK)

      return {
        text:    `${input.running ? '$(sync~spin)' : (input.failed > 0 ? '$(warning)' : '$(tools)')} ${parts.join(' · ')}`,
        tooltip: lines.join('\n'),
        warning: input.failed > 0,
      }
    }
  }
}
