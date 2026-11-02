import type { ViewState } from '../folder-session/index.ts'

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
    { label: '$(output) Open output', command: 'monosmith.showOutput' },
    { label: '$(gear) Open settings', command: 'monosmith.openSettings' },
  ]
  if (state.view === 'adoptable') {
    return [
      { label: '$(arrow-right) Convert it to MonoSmith', description: 'smith lists every change and asks first', command: 'monosmith.adopt.start' },
      { label: '$(eye) See what would change', description: 'a report, no files changed', command: 'monosmith.adopt.preview' },
      ...tail,
    ]
  }
  if (state.view === 'unrecognised') {
    return [
      { label: '$(arrow-right) Bring this folder into MonoSmith', description: 'smith lists every change and asks first', command: 'monosmith.adopt.start' },
      { label: '$(eye) See what would change', description: 'a report, no files changed', command: 'monosmith.adopt.preview' },
      ...tail,
    ]
  }
  if (state.view === 'empty') {
    return [
      { label: '$(add) Create a new MonoSmith workspace', description: 'smith asks a few questions first', command: 'monosmith.new' },
      ...tail,
    ]
  }

  return [
    ...(state.running ? [{ label: '$(debug-stop) Cancel run', command: 'monosmith.cancelRun' }] : []),
    { label: '$(play) Run affected…', description: 'one task in the projects the changes affect', command: 'monosmith.runAffected' },
    ...(state.hasLastRun ? [{ label: '$(debug-rerun) Re-run last', command: 'monosmith.rerunLast' }] : []),
    {
      label:       '$(pulse) Run doctor',
      description: state.failed > 0 ? `${state.failed} tasks failed in the last run` : 'check the workspace',
      command:     'monosmith.doctor',
    },
    { label: '$(refresh) Refresh', description: 'read the projects and what is affected again', command: 'monosmith.refresh' },
    { label: '$(add) Add project…', description: 'smith shows the plan and asks first', command: 'monosmith.addProject' },
    { label: '$(package) Add dependency…', description: 'a library of this workspace, or a package from the registry', command: 'monosmith.addDependency' },
    { label: '$(globe) Open report', description: 'one page: projects, health, derived files', command: 'monosmith.openReport' },
    { label: '$(discard) Undo the last change', description: 'smith shows what it reverts and asks first', command: 'monosmith.undo' },
    { label: '$(hubot) Talk to the MonoSmith agent', description: 'Claude Code with smith’s tools', command: 'monosmith.talkToAgent' },
    { label: '$(output) Open output', command: 'monosmith.showOutput' },
    { label: '$(gear) Open settings', command: 'monosmith.openSettings' },
  ]
}
