import type { ViewState } from '../folder-session/index.ts'
import { say } from '../words/index.ts'

/** One entry of the menu the status bar item opens. */
export interface MenuEntry {
  label:        string
  description?: string
  /** The command a pick runs. */
  command:      string
}

/** What the menu depends on. */
export interface MenuState {
  running:    boolean
  hasLastRun: boolean
  failed:     number
  /** What the view shows; with no folder that runs smith, the menu offers what that state offers instead of runs. */
  view?:      ViewState
}

/**
 * The entries of the status bar menu, in the order of what a person wants
 * most: run what the changes affect, run again, check health, look at the
 * output, change settings.
 */
export function statusMenuEntries (state: MenuState): MenuEntry[] {
  const tail = [
    { label: say('vscode.quick-pick.open-output'), command: 'monosmith.showOutput' },
    { label: say('vscode.quick-pick.open-settings'), command: 'monosmith.openSettings' },
  ]
  if (state.view === 'adoptable') {
    return [
      { label: say('vscode.quick-pick.convert-it-to-monosmith'), description: say('vscode.quick-pick.smith-lists-every-change-and-asks'), command: 'monosmith.adopt.start' },
      { label: say('vscode.quick-pick.see-what-would-change'), description: say('vscode.quick-pick.a-report-no-files-changed'), command: 'monosmith.adopt.preview' },
      ...tail,
    ]
  }
  if (state.view === 'unrecognised') {
    return [
      { label: say('vscode.quick-pick.bring-this-folder-into-monosmith'), description: say('vscode.quick-pick.smith-lists-every-change-and-asks'), command: 'monosmith.adopt.start' },
      { label: say('vscode.quick-pick.see-what-would-change'), description: say('vscode.quick-pick.a-report-no-files-changed'), command: 'monosmith.adopt.preview' },
      ...tail,
    ]
  }
  if (state.view === 'empty') {
    return [
      { label: say('vscode.quick-pick.create-a-new-monosmith-workspace'), description: say('vscode.quick-pick.smith-asks-a-few-questions-first'), command: 'monosmith.new' },
      ...tail,
    ]
  }

  return [
    ...(state.running ? [{ label: say('vscode.quick-pick.cancel-run'), command: 'monosmith.cancelRun' }] : []),
    { label: say('vscode.quick-pick.run-affected'), description: say('vscode.quick-pick.one-task-in-the-projects-the'), command: 'monosmith.runAffected' },
    ...(state.hasLastRun ? [{ label: say('vscode.quick-pick.re-run-last'), command: 'monosmith.rerunLast' }] : []),
    {
      label:       say('vscode.quick-pick.run-doctor'),
      description: state.failed > 0 ? say('vscode.quick-pick.tasks-failed-in-the-last-run', { failed: state.failed }) : say('vscode.quick-pick.check-the-workspace'),
      command:     'monosmith.doctor',
    },
    { label: say('vscode.quick-pick.refresh'), description: say('vscode.quick-pick.read-the-projects-and-what-is'), command: 'monosmith.refresh' },
    { label: say('vscode.quick-pick.add-project'), description: say('vscode.quick-pick.smith-shows-the-plan-and-asks'), command: 'monosmith.addProject' },
    { label: say('vscode.add-project.compose-a-stack'), description: say('vscode.quick-pick.search-frameworks-desktop-shells-and-cloud'), command: 'monosmith.composeStack' },
    { label: say('vscode.quick-pick.add-dependency'), description: say('vscode.quick-pick.a-library-of-this-workspace-or'), command: 'monosmith.addDependency' },
    { label: say('vscode.quick-pick.open-report'), description: say('vscode.quick-pick.one-page-projects-health-derived-files'), command: 'monosmith.openReport' },
    { label: say('vscode.quick-pick.undo-the-last-change'), description: say('vscode.quick-pick.smith-shows-what-it-reverts-and'), command: 'monosmith.undo' },
    { label: say('vscode.quick-pick.talk-to-the-monosmith-agent'), description: say('vscode.quick-pick.claude-code-with-smith-s-tools'), command: 'monosmith.talkToAgent' },
    { label: say('vscode.quick-pick.open-output'), command: 'monosmith.showOutput' },
    { label: say('vscode.quick-pick.open-settings'), command: 'monosmith.openSettings' },
  ]
}
