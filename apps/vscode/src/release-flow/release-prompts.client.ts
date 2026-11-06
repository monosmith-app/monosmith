import * as vscode from 'vscode'
import { showNotice } from '../notice-buttons/index.ts'
import { forcedConsequences } from './forced-release.mapper.ts'
import type { PlannedPackage, ReleaseChoices, ReleaseOutcome } from './release-flow.contract.ts'
import { say } from '../words/index.ts'

const TITLE = say('vscode.release-flow.monosmith-release')

/**
 * Shows the plan in a modal and asks which packages to hold back. Nothing is held by default.
 * @param plan - the text of `smith release --dry-run`
 * @param packages - the packages the release would publish
 * @returns the names to hold back (maybe none), or undefined when the person cancels
 */
export async function chooseHold (plan: string, packages: PlannedPackage[]): Promise<string[] | undefined> {
  const next = say('vscode.release-flow.continue')
  const seen = await vscode.window.showInformationMessage(
    say('vscode.release-flow.monosmith-this-is-what-a-release'),
    { modal: true, detail: say('vscode.release-flow.nothing-is-published-yet-next-you', { plan }) },
    next,
  )
  if (seen !== next) return undefined
  const picked = await vscode.window.showQuickPick(
    packages.map((item) => ({ label: item.name, description: `${item.current ?? 'new'} → ${item.next} · ${item.level}`, detail: say('vscode.release-flow.hold-it-back-it-does-not') })),
    { title: say('vscode.release-flow.hold-packages-back', { title: TITLE }), placeHolder: say('vscode.release-flow.tick-the-packages-that-must-not'), canPickMany: true, ignoreFocusOut: true },
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
    const done = say('vscode.release-flow.keep-the-bumps-as-they-are')
    const picked = await vscode.window.showQuickPick(
      [
        { label: done, description: bumps.size === 0 ? say('vscode.release-flow.the-commits-decide-the-levels') : `${bumps.size} changed`, name: '' },
        ...releasable.map((item) => ({ label: item.name, description: `${item.current ?? ''} → ${item.next} · ${bumps.get(item.name) ?? item.level}${bumps.has(item.name) ? say('vscode.release-flow.changed') : ''}`, detail: say('vscode.release-flow.change-its-bump-a-request-not'), name: item.name })),
      ],
      { title: say('vscode.release-flow.change-a-bump', { title: TITLE }), placeHolder: say('vscode.release-flow.pick-a-package-to-change-its'), ignoreFocusOut: true },
    )
    if (picked === undefined) return undefined
    if (picked.name === '') return [...bumps].map(([name, level]) => `${name}=${level}`)
    const level = await vscode.window.showQuickPick(
      [
        { label: 'patch', detail: say('vscode.release-flow.a-fix-1-2-3-1') },
        { label: 'minor', detail: say('vscode.release-flow.a-new-feature-1-2-3') },
        { label: 'major', detail: say('vscode.release-flow.a-breaking-change-1-2-3') },
      ],
      { title: say('vscode.release-flow.which-level-for', { title: TITLE, name: picked.name }), ignoreFocusOut: true },
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
  const go = say('vscode.release-flow.release-now')
  const held = choices.hold.length === 0 ? '' : say('vscode.release-flow.held-back', { join: choices.hold.join(', ') })
  const bumped = choices.bump.length === 0 ? '' : say('vscode.release-flow.changed-bumps', { join: choices.bump.join(', ') })
  const choice = await vscode.window.showWarningMessage(
    say('vscode.release-flow.monosmith-release-now'),
    { modal: true, detail: say('vscode.release-flow.this-cannot-be-undone-smith-creates', { plan, held, bumped, forcedConsequences: forcedConsequences(choices.forced ?? []) }) },
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
      void vscode.window.showInformationMessage(say('vscode.commit-changes.monosmith', { part: outcome.lines.at(-1) ?? say('vscode.release-flow.the-release-is-done') }))
      break
    }
    case 'nothing': {
      for (const line of outcome.lines) onOutput(line)
      void tellNothingToRelease(outcome.lines, outcome.unchanged)
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.release-flow.monosmith-nothing-was-released', { message: outcome.message }))
      break
    }
    case 'cancelled': {
      break
    }
  }
}

/**
 * Says why `MonoSmith: Release…` has nothing to release, and offers the two ways on:
 * the Release Preview, which explains what counts, and, when a package exists that no commit
 * releases, the forced release of it (the same code under a new version).
 * @param lines - what smith printed
 * @param unchanged - the packages no commit releases
 */
async function tellNothingToRelease (lines: readonly string[], unchanged: readonly string[]): Promise<void> {
  const preview = say('vscode.release-flow.open-preview')
  const forced = say('vscode.release-flow.without-a-commit')
  const first = lines.find((line) => line.startsWith(say('releaseview.nothing'))) ?? say('vscode.release-flow.nothing-to-release')
  const why = unchanged.length > 0
    ? say('vscode.release-flow.no-commit-releases-without-a-commit', { join: unchanged.join(', ') })
    : say('vscode.release-flow.a-package-releases-when-a-commit')
  const choice = await showNotice(
    'info',
    say('vscode.release-flow.monosmith-open-preview-opens-the-release', { first, why }),
    [
      { label: preview, detail: say('vscode.release-flow.open-the-release-preview-view-it') },
      ...(unchanged.length > 0 ? [{ label: forced, detail: say('vscode.release-flow.choose-a-package-to-release-although') }] : []),
    ],
  )
  if (choice === preview) await vscode.commands.executeCommand('monosmith.release.focus')
  if (choice === forced) await vscode.commands.executeCommand('monosmith.release')
}
