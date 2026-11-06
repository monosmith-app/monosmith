import * as vscode from 'vscode'
import type { WebEnvironments } from './web-environments.contract.ts'
import { checkEnvironmentNames, defaultEnvironment, defaultFirst, describeEnvironments, parseEnvironmentNames, startingEnvironments } from './web-environments.mapper.ts'
import { say } from '../words/index.ts'

const TITLE = say('vscode.add-project.monosmith-add-project')

/**
 * Asks how a web app is built once per environment as ONE question: the choices in a sentence (the
 * names, the environment `smith run build` builds, where the builds go, the .env files) with
 * "Use these" as the default, "Change the names…" and "Change the build default, the output or the
 * commit…". The last two open the detailed questions, each a presented choice with the default first.
 * The default environment is prd, else the last name: never lcl by accident.
 * @param app - the name of the new app
 * @param workspaceDefaults - `[workspace] environments`, when the workspace has them
 * @returns the answers, or undefined when the person pressed Escape at any question
 */
export async function askWebEnvironments (app: string, workspaceDefaults: readonly string[] = []): Promise<WebEnvironments | undefined> {
  let current = startingEnvironments(workspaceDefaults)
  for (;;) {
    const picked = await vscode.window.showQuickPick(
      [
        { label: say('vscode.web-environments.use-these'), description: 'default', detail: describeEnvironments(current), step: 'use' as const },
        { label: say('guide.web_environments.change_names'), detail: say('vscode.web-environments.vite-writes-its-vite-values-into'), step: 'names' as const },
        { label: say('vscode.web-environments.change-the-build-default-the-output'), detail: say('vscode.web-environments.which-environment-smith-run-build-builds'), step: 'details' as const },
      ],
      { title: say('vscode.web-environments.environments-of', { title: TITLE, app }), placeHolder: say('vscode.web-environments.nothing-is-written-until-you-have') },
    )
    if (picked === undefined) return undefined
    if (picked.step === 'use') return current
    if (picked.step === 'names') {
      const names = await askNames(app, current.names)
      if (names === undefined) return undefined
      current = { ...current, names, default: defaultEnvironment(names) }
    } else {
      const details = await askDetails(current)
      if (details === undefined) return undefined
      current = details
    }
  }
}

async function askNames (app: string, names: readonly string[]): Promise<string[] | undefined> {
  const typed = await vscode.window.showInputBox({
    title:         say('vscode.web-environments.environments-of', { title: TITLE, app }),
    prompt:        say('vscode.web-environments.edit-the-names-lowercase-words-separated'),
    value:         names.join(', '),
    validateInput: checkEnvironmentNames,
  })

  return typed === undefined ? undefined : parseEnvironmentNames(typed)
}

async function askDetails (current: WebEnvironments): Promise<WebEnvironments | undefined> {
  const fallback = defaultEnvironment(current.names)
  const chosen = await vscode.window.showQuickPick(
    defaultFirst(current.names, current.names.includes(current.default) ? current.default : fallback).map((name) => ({ label: name, description: name === fallback ? 'default' : '', detail: name === fallback ? say('vscode.web-environments.built-by-smith-run-build-what') : say('vscode.web-environments.built-by-smith-run-build', { name }) })),
    { title: say('vscode.web-environments.which-environment-is-build', { title: TITLE }), placeHolder: say('vscode.web-environments.ci-and-releases-build-this-one') },
  )
  if (chosen === undefined) return undefined

  const output = await vscode.window.showQuickPick(
    [
      { label: say('guide.web_environments_steps.in_the_workspace'), description: 'default', detail: say('vscode.web-environments.dist-apps-app-name-at-the'), value: 'workspace' as const },
      { label: say('guide.web_environments_steps.in_the_project'), detail: say('vscode.web-environments.dist-name-inside-the-app-s'), value: 'project' as const },
    ],
    { title: say('vscode.web-environments.where-do-the-builds-go', { title: TITLE }) },
  )
  if (output === undefined) return undefined

  const files = await vscode.window.showQuickPick(
    [
      { label: say('vscode.web-environments.commit-the-env-files'), description: 'default', detail: say('vscode.web-environments.they-hold-vite-values-which-end'), value: 'commit' as const },
      { label: say('vscode.web-environments.commit-only-env-example'), detail: say('vscode.web-environments.each-machine-and-ci-job-makes'), value: 'example' as const },
    ],
    { title: say('vscode.web-environments.the-env-files', { title: TITLE }) },
  )

  return files === undefined ? undefined : { names: current.names, default: chosen.label, output: output.value, envFiles: files.value }
}
