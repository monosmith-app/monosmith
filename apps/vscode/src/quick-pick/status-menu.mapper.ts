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
  /** No folder runs smith, but one can be converted: offer that instead of runs. */
  adoptable?: boolean
}

/**
 * The entries of the status bar menu, in the order of what a person wants
 * most: run what the changes affect, run again, check health, look at the
 * output, change settings.
 */
export function statusMenuEntries (state: MenuState): MenuEntry[] {
  if (state.adoptable === true) {
    return [
      { label: '$(arrow-right) Convert it to MonoSmith', description: 'smith lists every change and asks first', command: 'monosmith.adopt.start' },
      { label: '$(eye) See what would change', description: 'a report, no files changed', command: 'monosmith.adopt.preview' },
      { label: '$(output) Open output', command: 'monosmith.showOutput' },
      { label: '$(gear) Open settings', command: 'monosmith.openSettings' },
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
    { label: '$(hubot) Talk to the MonoSmith agent', description: 'Claude Code with smith’s tools', command: 'monosmith.talkToAgent' },
    { label: '$(output) Open output', command: 'monosmith.showOutput' },
    { label: '$(gear) Open settings', command: 'monosmith.openSettings' },
  ]
}
