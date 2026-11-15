import * as vscode from 'vscode'
import { parsePackages } from './install-arguments.mapper.ts'
import type { InstallOutcome, InstallRequest, LinkNode } from './install-request.contract.ts'
import { linkableLibraries } from './linkable-libraries.policy.ts'
import { say } from '../words/index.ts'

const TITLE = say('vscode.install-dependency.monosmith-add-a-dependency')

/**
 * Asks what to add. The first question is the one that matters: a library of this workspace or a package from
 * the registry. The first lists only the libraries smith would accept, with kind and language, so nothing is typed.
 * @param nodes - every project of the workspace
 * @param preselected - the project to add to, when the command came from a project
 * @param only - skips the first question: only a library of this workspace, or only a registry package
 * @returns the request, or undefined when the person pressed Escape
 */
export async function askInstallRequest (nodes: readonly LinkNode[], preselected?: string, only?: 'workspace' | 'registry'): Promise<InstallRequest | undefined> {
  const pickedProject = preselected === undefined ? await vscode.window.showQuickPick(nodes.map((node) => ({ label: node.name, description: `${node.kind} · ${node.language}` })), { title: say('vscode.install-dependency.1-3-which-project', { title: TITLE }) }) : undefined
  const project = preselected ?? pickedProject?.label
  if (project === undefined) return undefined
  const source = only === undefined
    ? await vscode.window.showQuickPick(
        [
          { label: say('packageprompt.source.workspace'), description: say('vscode.install-dependency.linked-nothing-is-downloaded'), detail: say('vscode.install-dependency.uses-the-code-of-another-project', { project }), value: 'workspace' as const },
          { label: say('vscode.install-dependency.a-package-from-the-registry'), description: '(default)', detail: say('vscode.install-dependency.downloaded-from-the-language-registry-npm'), value: 'registry' as const },
        ],
        { title: say('vscode.install-dependency.2-3-what-do-you-add', { title: TITLE, project }) },
      )
    : { value: only }
  if (source === undefined) return undefined
  if (source.value === 'registry') {
    const typed = await vscode.window.showInputBox({ title: `${TITLE} (3/3): packages`, prompt: say('vscode.install-dependency.package-names-separated-by-spaces'), placeHolder: 'zod  react@19', validateInput: (value) => parsePackages(value).length === 0 ? say('vscode.install-dependency.type-at-least-one-package') : undefined })

    return typed === undefined ? undefined : { project, source: 'registry', packages: parsePackages(typed), dev: false }
  }
  const libraries = linkableLibraries(nodes, project)
  if (libraries.length === 0) {
    void vscode.window.showWarningMessage(say('vscode.install-dependency.monosmith-no-library-of-this-workspace', { project, project2: project }))

    return undefined
  }
  const library = await vscode.window.showQuickPick(libraries.map((node) => ({ label: node.name, description: `${node.kind} · ${node.language}` })), { title: say('vscode.install-dependency.3-3-which-library', { title: TITLE }) })
  if (library === undefined) return undefined
  const language = nodes.find((node) => node.name === project)?.language
  if (language !== 'node') return { project, source: 'workspace', library: library.label, packages: [], dev: false }
  const range = await vscode.window.showQuickPick(
    [
      { label: say('vscode.install-dependency.follow-the-workspace'), description: '(default)', detail: say('packageprompt.range.follow.meaning'), value: undefined },
      { label: say('packageprompt.range.any.label'), detail: say('vscode.install-dependency.npm-uses-the-library-of-this'), value: '*' },
    ],
    { title: say('vscode.install-dependency.which-version-of-does-ask-for', { title: TITLE, label: library.label, project }) },
  )

  return range === undefined ? undefined : { project, source: 'workspace', library: library.label, ...(range.value !== undefined && { range: range.value }), packages: [], dev: false }
}

/**
 * Tells the person how adding a dependency ended.
 * @param outcome - what `installDependency` returned
 * @param onOutput - gets every line smith printed, for the output channel
 */
export function tellDependencyOutcome (outcome: InstallOutcome, onOutput: (line: string) => void): void {
  switch (outcome.kind) {
    case 'added': {
      for (const line of outcome.lines) onOutput(line)
      void vscode.window.showInformationMessage(say('vscode.install-dependency.monosmith-the-dependency-is-added-not'))
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.add-project.monosmith-nothing-was-added', { message: outcome.message }))
      break
    }
    case 'cancelled': {
      break
    }
  }
}
