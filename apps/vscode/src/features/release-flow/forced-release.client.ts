import * as vscode from 'vscode'
import { firstVersionCommand } from './forced-release.mapper.ts'
import { say } from '../words/index.ts'

const TITLE = say('vscode.release-flow.monosmith-release')

/**
 * Offers to release a package that no commit releases: the same code under a new version.
 * It is a choice, never a default: the first answer is to release only what the plan lists.
 * @param unchanged - the packages with no releasing commit
 * @param nothingPlanned - true when the plan releases nothing, so this is the only way to release
 * @param neverReleased - the packages smith refuses to force (no version to bump): each shows one line with the commit that sets its first version
 * @returns `<project>=patch|minor|major` for each package chosen (maybe none), or undefined when the person cancels
 */
export async function chooseForced (unchanged: string[], nothingPlanned: boolean, neverReleased: string[] = []): Promise<string[] | undefined> {
  const forced = new Map<string, string>()
  for (;;) {
    const done = say(nothingPlanned && forced.size === 0 ? 'vscode.release-flow.release-nothing' : 'vscode.release-flow.continue-with-the-plan')
    const never = neverReleased.map((name) => ({ label: `$(info) ${name}`, description: say('vscode.release-flow.never-released-no-version-to-bump'), detail: say('vscode.release-flow.smith-refuses-to-force-it-to', { firstVersionCommand: firstVersionCommand(name) }), name: `never:${name}` }))
    const picked = await vscode.window.showQuickPick(
      [
        { label: done, description: forced.size === 0 ? say('vscode.release-flow.no-package-is-released-without-a') : `${forced.size} chosen`, name: '' },
        ...unchanged.map((name) => ({ label: name, description: forced.has(name) ? `chosen: ${forced.get(name)}` : say('vscode.release-flow.no-commit-releases-it'), detail: say('vscode.release-flow.release-it-anyway-the-same-code'), name })),
        ...never,
      ],
      { title: say('vscode.release-flow.release-a-package-without-a-releasing', { title: TITLE }), placeHolder: say('vscode.release-flow.pick-a-package-to-release-it'), ignoreFocusOut: true },
    )
    if (picked === undefined) return undefined
    if (picked.name === '') return [...forced].map(([name, level]) => `${name}=${level}`)
    if (picked.name.startsWith('never:')) {
      const name = picked.name.slice('never:'.length)
      const copy = say('vscode.release-flow.copy-the-commit-command')
      const answer = await vscode.window.showInformationMessage(say('vscode.release-flow.monosmith-was-never-released-so-there', { name }), { modal: true, detail: say('vscode.release-flow.to-choose-its-first-version-make', { name, firstVersionCommand: firstVersionCommand(name) }) }, copy)
      if (answer === copy) await vscode.env.clipboard.writeText(firstVersionCommand(name))
      continue
    }
    const level = await vscode.window.showQuickPick(
      [
        { label: 'patch', detail: say('vscode.release-flow.same-code-1-2-3-1') },
        { label: 'minor', detail: '1.2.3 → 1.3.0.' },
        { label: 'major', detail: '1.2.3 → 2.0.0.' },
      ],
      { title: say('vscode.release-flow.which-level-for', { title: TITLE, name: picked.name }), ignoreFocusOut: true },
    )
    if (level === undefined) return undefined
    forced.set(picked.name, level.label)
  }
}
