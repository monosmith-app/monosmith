import * as vscode from 'vscode'
import { describeDoctorFailure, type DoctorFailure } from '../smith-run/index.ts'
import { showNotice } from '../notice-buttons/index.ts'
import { say } from '../words/index.ts'

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
  onOutput(say('vscode.doctor-failure.monosmith-was-written-and-smith-doctor', { name: failure.name }))
  for (const line of lines) onOutput(line)
  void askNextStep(failure, operation, next)
}

async function askNextStep (failure: DoctorFailure, operation: string, next?: DoctorFailureFollowUp): Promise<void> {
  const install = say('vscode.add-project.install-tools')
  const health = say('vscode.doctor-failure.open-health')
  const commit = say('vscode.doctor-failure.commit-anyway')
  const leave = say('vscode.doctor-failure.keep-uncommitted')
  const canCommit = next !== undefined && !failure.committed
  const choice = await showNotice(
    'warning',
    say('vscode.doctor-failure.install-tools-fetches-what-is-missing', { describeDoctorFailure: describeDoctorFailure(failure), part: canCommit ? say('vscode.doctor-failure.commit-anyway-makes-one-commit-that') : '' }),
    [
      { label: install, detail: say('vscode.doctor-failure.install-the-tools-the-project-needs') },
      { label: health, detail: say('vscode.doctor-failure.open-the-health-view-it-lists') },
      ...(canCommit ? [{ label: commit, detail: say('vscode.doctor-failure.commit-the-files-now-although-a') }] : []),
      { label: leave, detail: say('vscode.doctor-failure.leave-the-files-uncommitted-review-them') },
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
      const later = say('commands.ws.commit.title')
      void vscode.window.showInformationMessage(say('vscode.doctor-failure.monosmith-is-left-uncommitted-smith-remembers', { name: failure.name }), later).then(async (again) => {
        if (again === later) await next?.commit(operation)
      })
      break
    }
  }
}
