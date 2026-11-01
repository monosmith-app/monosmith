import * as vscode from 'vscode'
import { mcpConfigText } from './mcp-config.mapper.ts'
import type { InstallOutcome, InstallPreview, InstallScope } from './install-run.contract.ts'
import type { AgentFileOutcome } from './write-agent-file.use-case.ts'

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
      const agentsFile = 'Write AGENTS.md'
      void vscode.window.showInformationMessage(
        `MonoSmith: added ${what(scope)}. In VS Code, start the smith server from .vscode/mcp.json; it asks you to trust it once. AGENTS.md tells other agents the rules of the workspace.`,
        open,
        agentsFile,
      ).then(async (choice) => {
        if (choice === open) await vscode.window.showTextDocument(vscode.Uri.joinPath(vscode.Uri.file(root), '.vscode', 'mcp.json'))
        else if (choice === agentsFile) await vscode.commands.executeCommand('monosmith.agentFile', { root })
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

/**
 * Names what `AGENTS.md` lacks and asks before smith writes it. smith keeps
 * the text outside its markers.
 * @param outOfDate - the lines `smith agent file --check` printed
 * @param folder - the name of the workspace folder
 * @returns whether the person agreed
 */
export async function confirmAgentFile (outOfDate: readonly string[], folder: string): Promise<boolean> {
  const detail = `${outOfDate.map((line) => `● ${line}`).join('\n')}\n\nsmith writes only the lines between its markers. Your own text in AGENTS.md and CLAUDE.md stays. Review the files with git afterwards.`
  const choice = await vscode.window.showInformationMessage(`MonoSmith: write AGENTS.md for the AI agents of ${folder}?`, { modal: true, detail }, 'Write')

  return choice === 'Write'
}

/**
 * Says how writing `AGENTS.md` ended. It does not wait for the person to read it.
 * @param outcome - what `writeAgentFile` returned
 * @param root - the workspace folder, to open the file from
 */
export function tellAgentFileOutcome (outcome: AgentFileOutcome, root: string): void {
  switch (outcome.kind) {
    case 'written': {
      const open = 'Open AGENTS.md'
      void vscode.window.showInformationMessage('MonoSmith: AGENTS.md is written. Commit it so every agent and teammate reads the same rules.', open).then(async (choice) => {
        if (choice === open) await vscode.window.showTextDocument(vscode.Uri.joinPath(vscode.Uri.file(root), 'AGENTS.md'))
      })
      break
    }
    case 'current': {
      void vscode.window.showInformationMessage('MonoSmith: AGENTS.md is already current.')
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(`MonoSmith: AGENTS.md was not written. ${outcome.message}`)
      break
    }
    case 'declined': {
      break
    }
  }
}
