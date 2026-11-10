import { basename } from 'node:path'
import * as vscode from 'vscode'
import type { CommandHost } from '../command-host/index.ts'
import { showInsight, type InsightRow } from '../project-insight/index.ts'
import type { InsightDocuments } from '../project-insight/index.ts'
import { pickTask } from '../quick-pick/index.ts'
import { readFileWhy, readSettingWhy, readTaskWhy } from './read-why.use-case.ts'
import { say } from '../words/index.ts'

/** The parts of a row of the Projects view that a why command reads. */
interface WhyRow extends InsightRow {
  task?: string
}

/**
 * Shows why a file is as it is, why a task runs again, or how a project resolved.
 * @param host - the folders and the Output channel of the extension
 * @param insight - where the pages are shown
 * @param argument - what the editor passed to the command
 */
export async function showWhy (host: CommandHost, insight: InsightDocuments, argument: unknown): Promise<void> {
  const node = typeof argument === 'object' && argument !== null && 'kind' in argument ? argument as WhyRow : undefined
  const whyTask = async (root: string, project: string, task: string): Promise<void> => {
    const state = host.registry.folders.get(root)
    if (state !== undefined) await insight.show(`${project}-${task}-why.md`, await readTaskWhy(host.smithIn(state), project, task))
  }
  if (node?.kind === 'task' && node.project !== undefined && node.task !== undefined) return await whyTask(node.root, node.project, node.task)
  if (node?.kind === 'project') return await showInsight(host, insight, node, 'explain')

  const file = host.fileTarget(argument)
  // A file given by the explorer or the editor menu is answered at once; from the palette the person says what to ask about.
  const kinds = [
    ...(file === undefined ? [] : [{ label: say('vscode.why.this-file-2'), description: file.file, pick: 'file' as const }]),
    { label: say('vscode.why.a-project'), description: say('vscode.why.how-its-tasks-resolved-and-where'), pick: 'project' as const },
    { label: say('vscode.why.a-task'), description: say('vscode.why.why-it-runs-again-instead-of'), pick: 'task' as const },
    { label: say('vscode.why.a-setting'), description: say('vscode.why.why-a-setting-of-smith-toml'), pick: 'setting' as const },
  ]
  const chosen = argument instanceof vscode.Uri ? undefined : await vscode.window.showQuickPick(kinds, { title: say('vscode.why.monosmith-why'), placeHolder: say('vscode.why.what-do-you-want-to-know') })
  switch (argument instanceof vscode.Uri ? 'file' : chosen?.pick) {
    case 'file': {
      if (file !== undefined) await insight.show(`${basename(file.file)}-why.md`, await readFileWhy(host.smithIn(file.state), file.file))
      break
    }
    case 'project': {
      await showInsight(host, insight, undefined, 'explain')
      break
    }
    case 'setting': {
      const state = await host.folderFor(argument)
      const setting = state === undefined ? undefined : await vscode.window.showInputBox({ title: say('vscode.why.monosmith-why'), prompt: say('vscode.why.which-setting-of-smith-toml'), placeHolder: 'workspace.commit_operations' })
      if (state !== undefined && setting !== undefined && setting.trim() !== '') await insight.show(`${setting.trim()}-why.md`, await readSettingWhy(host.smithIn(state), setting.trim()))
      break
    }
    case 'task': {
      const picked = await pickTask(host.readyWorkspaces())
      if (picked?.target.kind === 'task') await whyTask(picked.root, picked.target.project, picked.target.task)
      else if (picked !== undefined) void vscode.window.showInformationMessage(say('vscode.why.monosmith-choose-one-task-of-a'))
      break
    }
    default: {
      break
    }
  }
}
