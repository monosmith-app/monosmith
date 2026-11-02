import * as vscode from 'vscode'
import { describeDoctorFailure, type DoctorFailure } from '../smith-run/index.ts'
import { showNotice } from '../notice-buttons/index.ts'

/** What the notice of a project that fails its checks can do beyond showing text: the editor carries these out. */
export interface DoctorFailureFollowUp {
  /** Commits the changes of the operation as one commit, after its own confirmation. */
  commit: (operation: string) => Promise<void>
}

/**
 * Says a project was written but fails its checks, and offers the next step as
 * buttons: install the tools, open Health, commit anyway, or leave it. The full
 * text of smith's steps goes to the Output channel. It does not wait for the person to read it.
 * @param failure - what smith printed
 * @param operation - the words that name the operation, such as "add app web"
 * @param lines - everything smith printed, for the Output channel
 * @param onOutput - gets every line, for the Output channel
 * @param next - carries out the buttons that need the editor
 */
export function tellDoctorFailure (failure: DoctorFailure, operation: string, lines: readonly string[], onOutput: (line: string) => void, next?: DoctorFailureFollowUp): void {
  onOutput(`MonoSmith: ${failure.name} was written, and smith doctor found a problem with it. smith's steps:`)
  for (const line of lines) onOutput(line)
  void askNextStep(failure, operation, next)
}

async function askNextStep (failure: DoctorFailure, operation: string, next?: DoctorFailureFollowUp): Promise<void> {
  const install = 'Install tools'
  const health = 'Open Health'
  const commit = 'Commit anyway'
  const leave = 'Keep uncommitted'
  const canCommit = next !== undefined && !failure.committed
  const choice = await showNotice(
    'warning',
    `${describeDoctorFailure(failure)} Install tools fetches what is missing. Open Health lists every check.${canCommit ? ' Commit anyway makes one commit that Undo can revert.' : ''}`,
    [
      { label: install, detail: 'Install the tools the project needs, with smith tools install.' },
      { label: health, detail: 'Open the Health view: it lists every check and its remedy.' },
      ...(canCommit ? [{ label: commit, detail: 'Commit the files now, although a check fails. It is one commit that Undo can revert.' }] : []),
      { label: leave, detail: 'Leave the files uncommitted: review them with git, and commit when you decide.' },
    ],
  )
  switch (choice) {
    case install: {
      await vscode.commands.executeCommand('monosmith.installTools')
      break
    }
    case health: {
      await vscode.commands.executeCommand('monosmith.health.focus')
      break
    }
    case commit: {
      await next?.commit(operation)
      break
    }
    case leave: {
      const later = 'Commit the changes…'
      void vscode.window.showInformationMessage(`MonoSmith: ${failure.name} is left uncommitted. smith remembers its files. Commit them when you decide: run "MonoSmith: Commit the changes…".`, later).then(async (again) => {
        if (again === later) await next?.commit(operation)
      })
      break
    }
  }
}
