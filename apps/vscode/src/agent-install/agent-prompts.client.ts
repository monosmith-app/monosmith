import * as vscode from 'vscode'
import { mcpConfigText } from './mcp-config.mapper.ts'
import type { InstallOutcome, InstallPreview, InstallScope } from './install-run.contract.ts'
import type { AgentFileOutcome } from './write-agent-file.use-case.ts'
import { say } from '../words/index.ts'

const DOCS = 'https://github.com/monosmith-app/monosmith/blob/main/docs/guides/agent.md'
const CLAUDE_CODE = 'https://claude.com/claude-code'

/** What the person chose on the one-time offer. */
export type OfferAnswer = 'install' | 'later' | 'never'

function what (scope: InstallScope): string {
  return say(scope === 'agent' ? 'vscode.agent-install.the-monosmith-agent-and-the-smith' : 'vscode.agent-install.the-smith-mcp-server')
}

/**
 * Shows the files an install would write and asks. The diffs open in an
 * editor on request, and the question is then asked again.
 */
export async function confirmInstall (preview: InstallPreview, scope: InstallScope, folder: string): Promise<boolean> {
  const detail = say('vscode.agent-install.smith-merges-into-an-existing-mcp', { join: preview.actions.map((action) => `● ${action}`).join('\n') })
  const choice = await vscode.window.showInformationMessage(say('vscode.agent-install.monosmith-add-to', { what: what(scope), folder }), { modal: true, detail }, say('packageprompt.button.install'), say('vscode.agent-install.show-the-diff'))
  if (choice === say('packageprompt.button.install')) return true
  if (choice !== say('vscode.agent-install.show-the-diff')) return false
  const document = await vscode.workspace.openTextDocument({ language: 'diff', content: `${preview.diff}\n` })
  await vscode.window.showTextDocument(document, { preview: true })
  const again = await vscode.window.showInformationMessage(say('vscode.agent-install.monosmith-write-the-shown-in-the', { length: preview.actions.length, part: say(preview.actions.length === 1 ? 'vscode.agent-install.file' : 'vscode.agent-install.files') }), say('packageprompt.button.install'))

  return again === say('packageprompt.button.install')
}

/** Says how an install ended. It does not wait for the person to read it. */
export function tellInstallOutcome (outcome: InstallOutcome, scope: InstallScope, root: string): void {
  switch (outcome.kind) {
    case 'written': {
      const open = say('vscode.agent-install.open-mcp-json')
      const agentsFile = say('vscode.agent-install.write-agents-md')
      void vscode.window.showInformationMessage(
        say('vscode.agent-install.monosmith-added-in-vs-code-start', { what: what(scope) }),
        open,
        agentsFile,
      ).then(async (choice) => {
        if (choice === open) await vscode.window.showTextDocument(vscode.Uri.joinPath(vscode.Uri.file(root), '.vscode', 'mcp.json'))
        else if (choice === agentsFile) await vscode.commands.executeCommand('monosmith.agentFile', { root })
      })
      break
    }
    case 'current': {
      void vscode.window.showInformationMessage(say('vscode.agent-install.monosmith-is-already-installed-and-current', { what: what(scope) }))
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.agent-install.monosmith-nothing-was-written', { message: outcome.message }))
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
    say('vscode.agent-install.monosmith-has-no-monosmith-agent-it', { folder }),
    say('packageprompt.button.install'),
    say('adopt.agent.decline'),
    say('vscode.agent-install.never-for-this-workspace'),
  )
  if (choice === say('packageprompt.button.install')) return 'install'
  if (choice === say('adopt.agent.decline')) return 'later'

  return choice === undefined ? undefined : 'never'
}

/** Says Claude Code is missing and offers what helps: the MCP config for another client, or the docs. */
export async function explainNoClaudeCode (): Promise<void> {
  const copy = say('vscode.agent-install.copy-the-mcp-config')
  const choice = await vscode.window.showInformationMessage(
    say('vscode.agent-install.monosmith-the-claude-command-was-not', { claudeCode: CLAUDE_CODE }),
    copy,
    say('vscode.agent-install.open-the-docs'),
  )
  if (choice === copy) {
    await vscode.env.clipboard.writeText(mcpConfigText())
    void vscode.window.showInformationMessage(say('vscode.agent-install.monosmith-copied-the-mcp-config-add'))
  } else if (choice === say('vscode.agent-install.open-the-docs')) {
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
  const detail = say('vscode.agent-install.smith-writes-only-the-lines-between', { join: outOfDate.map((line) => `● ${line}`).join('\n') })
  const choice = await vscode.window.showInformationMessage(say('vscode.agent-install.monosmith-write-agents-md-for-the', { folder }), { modal: true, detail }, say('vscode.agent-install.write'))

  return choice === say('vscode.agent-install.write')
}

/**
 * Says how writing `AGENTS.md` ended. It does not wait for the person to read it.
 * @param outcome - what `writeAgentFile` returned
 * @param root - the workspace folder, to open the file from
 */
export function tellAgentFileOutcome (outcome: AgentFileOutcome, root: string): void {
  switch (outcome.kind) {
    case 'written': {
      const open = say('vscode.agent-install.open-agents-md')
      void vscode.window.showInformationMessage(say('vscode.agent-install.monosmith-agents-md-is-written-commit'), open).then(async (choice) => {
        if (choice === open) await vscode.window.showTextDocument(vscode.Uri.joinPath(vscode.Uri.file(root), 'AGENTS.md'))
      })
      break
    }
    case 'current': {
      void vscode.window.showInformationMessage(say('vscode.agent-install.monosmith-agents-md-is-already-current'))
      break
    }
    case 'refused': {
      void vscode.window.showWarningMessage(say('vscode.agent-install.monosmith-agents-md-was-not-written', { message: outcome.message }))
      break
    }
    case 'declined': {
      break
    }
  }
}

/** Says adding the agent writes files and needs a trusted workspace. */
export function tellAgentNeedsTrust (): void {
  void vscode.window.showWarningMessage(say('vscode.agent-install.monosmith-adding-the-agent-writes-files'))
}

/**
 * Says smith.toml was not changed when the person asked never to be offered the agent.
 * @param failure - smith's words for why
 */
export function tellAgentOptOutFailed (failure: string): void {
  void vscode.window.showWarningMessage(say('vscode.agent-install.monosmith-smith-toml-was-not-changed', { failure }))
}

/**
 * Says the folder will not be offered the agent again.
 * @param folder - the name of the folder
 */
export function tellAgentOptedOut (folder: string): void {
  void vscode.window.showInformationMessage(say('vscode.agent-install.monosmith-will-not-be-offered-the', { folder }))
}
