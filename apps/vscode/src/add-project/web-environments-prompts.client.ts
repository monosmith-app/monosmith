import * as vscode from 'vscode'
import type { WebEnvironments } from './add-options.contract.ts'
import { checkEnvironmentNames, DEFAULT_ENVIRONMENTS, defaultEnvironment, parseEnvironmentNames } from './web-environments.mapper.ts'

const TITLE = 'MonoSmith: Add project'

/**
 * Asks how a web app is built once per environment, each as a presented choice with the
 * default marked: the names (pre-filled, edit them), the environment `smith run build` builds,
 * where the builds go and whether the .env files are committed. The same questions as
 * `smith add` asks in a terminal.
 * @param app - the name of the new app
 * @param workspaceDefaults - `[workspace] environments`, when the workspace has them
 * @returns the answers, or undefined when the person pressed Escape at any question
 */
export async function askWebEnvironments (app: string, workspaceDefaults: readonly string[] = []): Promise<WebEnvironments | undefined> {
  const typed = await vscode.window.showInputBox({
    title:         `${TITLE}: environments of ${app}`,
    prompt:        'Vite writes its VITE_ values into the bundle when it builds, so the app is built once per environment, each with its own .env.<name> file. Edit the names (lowercase words, separated by commas).',
    value:         (workspaceDefaults.length > 0 ? workspaceDefaults : DEFAULT_ENVIRONMENTS).join(', '),
    validateInput: checkEnvironmentNames,
  })
  if (typed === undefined) return undefined
  const names = parseEnvironmentNames(typed)

  const fallback = defaultEnvironment(names)
  const chosen = await vscode.window.showQuickPick(
    names.map((name) => ({ label: name, description: name === fallback ? 'default' : '', detail: name === fallback ? 'built by `smith run build`: what CI and the release build' : `built by \`smith run build:${name}\`` })),
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

  return files === undefined ? undefined : { names, default: chosen.label, output: output.value, envFiles: files.value }
}
