import * as vscode from 'vscode'
import { parsePackages } from './install-arguments.mapper.ts'
import type { InstallOutcome, InstallRequest, LinkNode } from './install-request.contract.ts'
import { linkableLibraries } from './linkable-libraries.policy.ts'

const TITLE = 'MonoSmith: add a dependency'

/**
 * Asks what to add. The first question is the one that matters: a library of this workspace or a package from
 * the registry. The first lists only the libraries smith would accept, with kind and language, so nothing is typed.
 * @param nodes - every project of the workspace
 * @param preselected - the project to add to, when the command came from a project
 * @returns the request, or undefined when the person pressed Escape
 */
export async function askInstallRequest (nodes: readonly LinkNode[], preselected?: string): Promise<InstallRequest | undefined> {
  const pickedProject = preselected === undefined ? await vscode.window.showQuickPick(nodes.map((node) => ({ label: node.name, description: `${node.kind} · ${node.language}` })), { title: `${TITLE} (1/3): which project?` }) : undefined
  const project = preselected ?? pickedProject?.label
  if (project === undefined) return undefined
  const source = await vscode.window.showQuickPick(
    [
      { label: 'A library of this workspace', description: 'linked, nothing is downloaded', detail: `${project} uses the code of another project here, always its current copy.`, value: 'workspace' as const },
      { label: 'A package from the registry', description: '(default)', detail: 'Downloaded from the language registry (npm, PyPI, crates.io…), a copy outside this workspace.', value: 'registry' as const },
    ],
    { title: `${TITLE} (2/3): what do you add to ${project}?` },
  )
  if (source === undefined) return undefined
  if (source.value === 'registry') {
    const typed = await vscode.window.showInputBox({ title: `${TITLE} (3/3): packages`, prompt: 'Package names, separated by spaces', placeHolder: 'zod  react@19', validateInput: (value) => parsePackages(value).length === 0 ? 'Type at least one package.' : undefined })

    return typed === undefined ? undefined : { project, source: 'registry', packages: parsePackages(typed), dev: false }
  }
  const libraries = linkableLibraries(nodes, project)
  if (libraries.length === 0) {
    void vscode.window.showWarningMessage(`MonoSmith: no library of this workspace can be linked to ${project}. A library must be a lib or a package, in the same language, and must not already depend on ${project}.`)

    return undefined
  }
  const library = await vscode.window.showQuickPick(libraries.map((node) => ({ label: node.name, description: `${node.kind} · ${node.language}` })), { title: `${TITLE} (3/3): which library?` })
  if (library === undefined) return undefined
  const language = nodes.find((node) => node.name === project)?.language
  if (language !== 'node') return { project, source: 'workspace', library: library.label, packages: [], dev: false }
  const range = await vscode.window.showQuickPick(
    [
      { label: 'Follow the workspace', description: '(default)', detail: 'Use the range the workspace already uses for its own libraries. When it has none yet, smith writes * and says so.', value: undefined },
      { label: 'Any version (*)', detail: 'npm uses the library of this workspace whatever its version. You never edit the line when the library changes.', value: '*' },
    ],
    { title: `${TITLE}: which version of ${library.label} does ${project} ask for?` },
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
      void vscode.window.showInformationMessage('MonoSmith: the dependency is added. Not committed: review with git status.')
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was added. ${outcome.message}`)
      break
    }
    case 'cancelled': {
      break
    }
  }
}
