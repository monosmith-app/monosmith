import * as vscode from 'vscode'

const TITLE = 'MonoSmith: Release'

/**
 * Offers to release a package that no commit releases: the same code under a new version.
 * It is a choice, never a default: the first answer is to release only what the plan lists.
 * @param unchanged - the packages with no releasing commit
 * @param nothingPlanned - true when the plan releases nothing, so this is the only way to release
 * @returns `<project>=patch|minor|major` for each package chosen (maybe none), or undefined when the person cancels
 */
export async function chooseForced (unchanged: string[], nothingPlanned: boolean): Promise<string[] | undefined> {
  const forced = new Map<string, string>()
  for (;;) {
    const done = nothingPlanned && forced.size === 0 ? '$(close) Release nothing' : '$(check) Continue with the plan'
    const picked = await vscode.window.showQuickPick(
      [
        { label: done, description: forced.size === 0 ? 'no package is released without a releasing commit' : `${forced.size} chosen`, name: '' },
        ...unchanged.map((name) => ({ label: name, description: forced.has(name) ? `chosen: ${forced.get(name)}` : 'no commit releases it', detail: 'Release it anyway: the same code under a new version. It creates a tag at HEAD, publishes to the registry, makes a GitHub Release and cannot be undone.', name })),
      ],
      { title: `${TITLE}: release a package without a releasing commit?`, placeHolder: 'Pick a package to release it anyway, or continue without one.', ignoreFocusOut: true },
    )
    if (picked === undefined) return undefined
    if (picked.name === '') return [...forced].map(([name, level]) => `${name}=${level}`)
    const level = await vscode.window.showQuickPick(
      [
        { label: 'patch', detail: 'Same code, 1.2.3 → 1.2.4.' },
        { label: 'minor', detail: '1.2.3 → 1.3.0.' },
        { label: 'major', detail: '1.2.3 → 2.0.0.' },
      ],
      { title: `${TITLE}: which level for ${picked.name}?`, ignoreFocusOut: true },
    )
    if (level === undefined) return undefined
    forced.set(picked.name, level.label)
  }
}
