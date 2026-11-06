import * as vscode from 'vscode'
import type { WebEnvironments } from './web-environments.contract.ts'
import { checkEnvironmentNames, defaultEnvironment, defaultFirst, describeEnvironments, parseEnvironmentNames, startingEnvironments } from './web-environments.mapper.ts'

const TITLE = 'MonoSmith: Add project'

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
        { label: 'Use these', description: 'default', detail: describeEnvironments(current), step: 'use' as const },
        { label: 'Change the names…', detail: 'Vite writes its VITE_ values into the bundle when it builds, so the app is built once per environment, each with its own .env.<name> file.', step: 'names' as const },
        { label: 'Change the build default, the output or the commit…', detail: 'Which environment `smith run build` builds, where the builds go and whether git keeps the .env files.', step: 'details' as const },
      ],
      { title: `${TITLE}: environments of ${app}`, placeHolder: 'Nothing is written until you have seen the plan' },
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
    title:         `${TITLE}: environments of ${app}`,
    prompt:        'Edit the names (lowercase words, separated by commas).',
    value:         names.join(', '),
    validateInput: checkEnvironmentNames,
  })

  return typed === undefined ? undefined : parseEnvironmentNames(typed)
}

async function askDetails (current: WebEnvironments): Promise<WebEnvironments | undefined> {
  const fallback = defaultEnvironment(current.names)
  const chosen = await vscode.window.showQuickPick(
    defaultFirst(current.names, current.names.includes(current.default) ? current.default : fallback).map((name) => ({ label: name, description: name === fallback ? 'default' : '', detail: name === fallback ? 'built by `smith run build`: what CI and the release build' : `built by \`smith run build:${name}\`` })),
    { title: `${TITLE}: which environment is "build"?`, placeHolder: 'CI and releases build this one; the others are build:<name>' },
  )
  if (chosen === undefined) return undefined

  const output = await vscode.window.showQuickPick(
    [
      { label: 'In the workspace', description: 'default', detail: 'dist/apps/<app>/<name> at the workspace root, with every other build', value: 'workspace' as const },
      { label: 'In the project', detail: 'dist/<name> inside the app\'s own folder; smith doctor then accepts it', value: 'project' as const },
    ],
    { title: `${TITLE}: where do the builds go?` },
  )
  if (output === undefined) return undefined

  const files = await vscode.window.showQuickPick(
    [
      { label: 'Commit the .env files', description: 'default', detail: 'they hold VITE_ values, which end up in the browser, so they hold no secret; CI builds with them', value: 'commit' as const },
      { label: 'Commit only .env.example', detail: 'each machine and CI job makes its own .env.<name>; git ignores them', value: 'example' as const },
    ],
    { title: `${TITLE}: the .env files` },
  )

  return files === undefined ? undefined : { names: current.names, default: chosen.label, output: output.value, envFiles: files.value }
}
