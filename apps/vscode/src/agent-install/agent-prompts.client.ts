import * as vscode from 'vscode'
import { mcpConfigText } from './mcp-config.mapper.ts'
import type { InstallOutcome, InstallPreview, InstallScope } from './install-run.contract.ts'

const DOCS = 'https://github.com/monosmith-app/monosmith/blob/main/docs/guides/agent.md'
const CLAUDE_CODE = 'https://claude.com/claude-code'

/** What the person chose on the one-time offer. */
export type OfferAnswer = 'install' | 'later' | 'never'

function what (scope: InstallScope): string {
  return scope === 'agent' ? 'the MonoSmith agent and the smith MCP server' : 'the smith MCP server'
}

/**
 * Shows the files an install would write and asks. The diffs open in an
 * editor on request, and the question is then asked again.
 */
export async function confirmInstall (preview: InstallPreview, scope: InstallScope, folder: string): Promise<boolean> {
  const detail = `${preview.actions.map((action) => `● ${action}`).join('\n')}\n\nsmith merges into an existing MCP file and keeps your other servers. It never overwrites a file you edited.`
  const choice = await vscode.window.showInformationMessage(`MonoSmith: add ${what(scope)} to ${folder}?`, { modal: true, detail }, 'Install', 'Show the diff')
  if (choice === 'Install') return true
  if (choice !== 'Show the diff') return false
  const document = await vscode.workspace.openTextDocument({ language: 'diff', content: `${preview.diff}\n` })
  await vscode.window.showTextDocument(document, { preview: true })
  const again = await vscode.window.showInformationMessage(`MonoSmith: write the ${preview.actions.length} ${preview.actions.length === 1 ? 'file' : 'files'} shown in the diff?`, 'Install')

  return again === 'Install'
}

/** Says how an install ended. It does not wait for the person to read it. */
export function tellInstallOutcome (outcome: InstallOutcome, scope: InstallScope, root: string): void {
  switch (outcome.kind) {
    case 'written': {
      const open = 'Open .vscode/mcp.json'
      void vscode.window.showInformationMessage(
        `MonoSmith: added ${what(scope)}. In VS Code, start the smith server from .vscode/mcp.json; it asks you to trust it once.`,
        open,
      ).then(async (choice) => {
        if (choice === open) await vscode.window.showTextDocument(vscode.Uri.joinPath(vscode.Uri.file(root), '.vscode', 'mcp.json'))
      })
      break
    }
    case 'current': {
      void vscode.window.showInformationMessage(`MonoSmith: ${what(scope)} is already installed and current.`)
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: nothing was written. ${outcome.message}`)
      break
    }
    case 'declined': {
      break
    }
  }
}

/** The one-time offer after the editor finds a smith workspace without the agent. */
export async function offerAgent (folder: string): Promise<OfferAnswer | undefined> {
  const choice = await vscode.window.showInformationMessage(
    `MonoSmith: ${folder} has no MonoSmith agent. It answers questions about the workspace and runs smith through its MCP server, and asks before it changes anything.`,
    'Install',
    'Not now',
    'Never for this workspace',
  )
  if (choice === 'Install') return 'install'
  if (choice === 'Not now') return 'later'

  return choice === undefined ? undefined : 'never'
}

/** Says Claude Code is missing and offers what helps: the MCP config for another client, or the docs. */
export async function explainNoClaudeCode (): Promise<void> {
  const copy = 'Copy the MCP config'
  const choice = await vscode.window.showInformationMessage(
    `MonoSmith: the claude command was not found. Install Claude Code (${CLAUDE_CODE}) to talk to the agent in a terminal. In VS Code, the smith MCP server works in agent mode without it.`,
    copy,
    'Open the docs',
  )
  if (choice === copy) {
    await vscode.env.clipboard.writeText(mcpConfigText())
    void vscode.window.showInformationMessage('MonoSmith: copied the MCP config. Add it to the .mcp.json of your client.')
  } else if (choice === 'Open the docs') {
    await vscode.env.openExternal(vscode.Uri.parse(DOCS))
  }
}
