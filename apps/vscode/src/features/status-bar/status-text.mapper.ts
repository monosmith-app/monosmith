import type { ViewState } from '../folder-session/index.ts'
import { say } from '../words/index.ts'

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

const CLICK_TO_CONVERT = say('vscode.status-bar.click-for-a-menu-convert-it')
const CLICK_TO_CREATE = say('vscode.status-bar.click-for-a-menu-create-a')
const CLICK_TO_BRING_IN = say('vscode.status-bar.click-for-a-menu-bring-this')
const CLICK = say('vscode.status-bar.click-for-a-menu-run-affected')

/**
 * The status bar text: the smith that is running, how many projects the
 * changes affect when that is known, and how many tasks failed. When smith
 * cannot be used, the reason.
 */
export function describeStatus (input: StatusInput): StatusText {
  switch (input.state) {
    case 'starting': {
      return { text: '$(sync~spin) smith', tooltip: say('vscode.status-bar.monosmith-starting-smith'), warning: false }
    }
    case 'nofolder': {
      return { text: '$(tools) smith', tooltip: say('vscode.status-bar.monosmith-no-folder-is-open'), warning: false }
    }
    case 'empty': {
      return {
        text:    '$(tools) smith',
        tooltip: say('vscode.status-bar.monosmith-this-folder-is-empty', { clickToCreate: CLICK_TO_CREATE }),
        warning: false,
      }
    }
    case 'unrecognised': {
      return {
        text:    '$(tools) smith',
        tooltip: say('vscode.status-bar.monosmith-this-folder-is-not-a', { clickToBringIn: CLICK_TO_BRING_IN }),
        warning: false,
      }
    }
    case 'adoptable': {
      return {
        text:    '$(tools) smith',
        tooltip: say('vscode.status-bar.monosmith-this-folder-is-not-a-2', { clickToConvert: CLICK_TO_CONVERT }),
        warning: false,
      }
    }
    case 'missing': {
      return { text: '$(circle-slash) smith', tooltip: say('vscode.status-bar.monosmith-smith-was-not-found', { click: CLICK }), warning: true }
    }
    case 'untrusted': {
      return { text: '$(shield) smith', tooltip: say('vscode.status-bar.monosmith-this-workspace-is-not-trusted', { click: CLICK }), warning: true }
    }
    case 'failed': {
      return { text: '$(error) smith', tooltip: say('vscode.status-bar.monosmith-smith-could-not-start-or', { click: CLICK }), warning: true }
    }
    case 'ready': {
      const parts = [`smith ${input.version ?? ''}`.trim()]
      if (input.affected !== undefined) parts.push(`${input.affected} affected`)
      if (input.failed > 0) parts.push(`${input.failed} failed`)
      const lines = [...input.origin]
      if (input.running) lines.push(say('vscode.status-bar.a-run-is-in-progress'))
      if (input.failed > 0) lines.push(say('vscode.status-bar.failed-in-the-last-runs', { failed: input.failed, part: say(input.failed === 1 ? 'vscode.status-bar.task' : 'vscode.status-bar.tasks') }))
      lines.push(CLICK)

      return {
        text:    `${input.running ? '$(sync~spin)' : (input.failed > 0 ? '$(warning)' : '$(tools)')} ${parts.join(' · ')}`,
        tooltip: lines.join('\n'),
        warning: input.failed > 0,
      }
    }
  }
}
