import * as vscode from 'vscode'
import { describeDoctorFailure, type DoctorFailure } from '../smith-run/index.ts'

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
  const install = 'Install the tools'
  const health = 'Open Health'
  const commit = 'Commit anyway'
  const leave = 'Leave it uncommitted'
  const choice = await vscode.window.showWarningMessage(
    describeDoctorFailure(failure),
    install,
    health,
    ...(next === undefined || failure.committed ? [] : [commit]),
    leave,
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
