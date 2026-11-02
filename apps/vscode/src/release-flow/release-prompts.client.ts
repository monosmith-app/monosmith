import * as vscode from 'vscode'
import type { PlannedPackage, ReleaseChoices, ReleaseOutcome } from './release-flow.contract.ts'

const TITLE = 'MonoSmith: Release'

/**
 * Shows the plan in a modal and asks which packages to hold back. Nothing is held by default.
 * @param plan - the text of `smith release --dry-run`
 * @param packages - the packages the release would publish
 * @returns the names to hold back (maybe none), or undefined when the person cancels
 */
export async function chooseHold (plan: string, packages: PlannedPackage[]): Promise<string[] | undefined> {
  const next = 'Continue'
  const seen = await vscode.window.showInformationMessage(
    'MonoSmith: this is what a release would do now.',
    { modal: true, detail: `${plan}\n\nNothing is published yet. Next you can hold packages back and change a bump. After that comes the last question.` },
    next,
  )
  if (seen !== next) return undefined
  const picked = await vscode.window.showQuickPick(
    packages.map((item) => ({ label: item.name, description: `${item.current ?? 'new'} → ${item.next} · ${item.level}`, detail: 'Hold it back: it does not release this time. A package that releases only because of it is held back too.' })),
    { title: `${TITLE}: hold packages back?`, placeHolder: 'Tick the packages that must not release this time. Enter with nothing ticked releases them all.', canPickMany: true, ignoreFocusOut: true },
  )

  return picked?.map((item) => item.label)
}

/**
 * Asks which bumps to override, one at a time: a package, then a level, until the person is done.
 * A package that was never released has no bump to override (smith refuses it), so it is not offered.
 * @param packages - the packages that are not held back
 * @returns `<project>=patch|minor|major` for each override, or undefined when the person cancels
 */
export async function chooseBump (packages: PlannedPackage[]): Promise<string[] | undefined> {
  const bumps = new Map<string, string>()
  const releasable = packages.filter((item) => item.current !== undefined)
  if (releasable.length === 0) return []
  for (;;) {
    const done = '$(check) Keep the bumps as they are'
    const picked = await vscode.window.showQuickPick(
      [
        { label: done, description: bumps.size === 0 ? 'the commits decide the levels' : `${bumps.size} changed`, name: '' },
        ...releasable.map((item) => ({ label: item.name, description: `${item.current ?? ''} → ${item.next} · ${bumps.get(item.name) ?? item.level}${bumps.has(item.name) ? ' (changed)' : ''}`, detail: 'Change its bump: a request, not an inference. On a 0.x package, major gives 1.0.0.', name: item.name })),
      ],
      { title: `${TITLE}: change a bump?`, placeHolder: 'Pick a package to change its level, or keep them all.', ignoreFocusOut: true },
    )
    if (picked === undefined) return undefined
    if (picked.name === '') return [...bumps].map(([name, level]) => `${name}=${level}`)
    const level = await vscode.window.showQuickPick(
      [
        { label: 'patch', detail: 'A fix: 1.2.3 → 1.2.4.' },
        { label: 'minor', detail: 'A new feature: 1.2.3 → 1.3.0.' },
        { label: 'major', detail: 'A breaking change: 1.2.3 → 2.0.0.' },
      ],
      { title: `${TITLE}: which level for ${picked.name}?`, ignoreFocusOut: true },
    )
    if (level === undefined) return undefined
    bumps.set(picked.name, level.label)
  }
}

/**
 * Shows the final plan and says that a release cannot be undone.
 * @param plan - the text of the last `smith release --dry-run`, with the choices
 * @param choices - the packages held back and the bumps changed
 * @returns true on a yes
 */
export async function confirmRelease (plan: string, choices: ReleaseChoices): Promise<boolean> {
  const go = 'Release now'
  const held = choices.hold.length === 0 ? '' : `\nHeld back: ${choices.hold.join(', ')}.`
  const bumped = choices.bump.length === 0 ? '' : `\nChanged bumps: ${choices.bump.join(', ')}.`
  const choice = await vscode.window.showWarningMessage(
    'MonoSmith: release now?',
    { modal: true, detail: `${plan}${held}${bumped}\n\nThis cannot be undone. smith creates the tags, pushes them and publishes the packages to the registries. Undo does not revert it.` },
    go,
  )

  return choice === go
}

/**
 * Says how the release ended. It does not wait for the person to read it.
 * @param outcome - what `releaseWorkspace` returned
 * @param onOutput - gets every line smith printed, for the output channel
 */
export function tellReleaseOutcome (outcome: ReleaseOutcome, onOutput: (line: string) => void): void {
  switch (outcome.kind) {
    case 'released': {
      for (const line of outcome.lines) onOutput(line)
      void vscode.window.showInformationMessage(`MonoSmith: ${outcome.lines.at(-1) ?? 'the release is done.'}`)
      break
    }
    case 'nothing': {
      for (const line of outcome.lines) onOutput(line)
      void vscode.window.showInformationMessage(`MonoSmith: ${outcome.lines.find((line) => line.startsWith('Nothing to release')) ?? 'there is nothing to release.'} The Release Preview view explains what counts.`)
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was released. ${outcome.message}`)
      break
    }
    case 'cancelled': {
      break
    }
  }
}
