import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import type { NewWorkspaceChoices, NewWorkspaceOutcome } from './new-workspace.contract.ts'
import { DEFAULT_CHOICES, describeChoices, toCreateArguments } from './new-workspace-arguments.mapper.ts'

/** How the person wants to set the workspace up. */
export type SetupWay = 'defaults' | 'customise' | 'terminal'

/** What creating a workspace needs from the editor. */
export interface NewWorkspaceDeps {
  /** The name of the open folder, the name smith would use if the person types none. */
  folderName: string
  /** Runs smith in the folder for a short answer (`new … --dry-run`). */
  run:        (args: string[]) => Promise<SmithRun>
  /** Runs smith in the folder to its end (`new …`); it installs nothing but writes files and may commit. */
  apply:      (args: string[]) => Promise<SmithRun>
  /** Asks "How do you want to set it up?": the defaults, customise, or smith's own questions in a terminal. Resolves undefined when the person cancels. */
  askWay:     (defaults: readonly string[]) => Promise<SetupWay | undefined>
  /** Asks the name, saying that the folder's name is used unless another is typed. Resolves undefined when the person cancels. */
  askName:    (folderName: string) => Promise<string | undefined>
  /** Asks each question of `smith new` with the defaults as the pre-selected answers. Resolves undefined when the person cancels. */
  askCustom:  (start: NewWorkspaceChoices) => Promise<NewWorkspaceChoices | undefined>
  /** Shows what will be chosen and the files smith lists, and asks. */
  confirm:    (summary: readonly string[], plan: string) => Promise<boolean>
}

/**
 * Creates a workspace in the open folder with the same two-way start as the TUI:
 * "Start with the defaults" lists what it chooses and asks only the name;
 * "Customise" asks every question. Before anything is written it shows the
 * choices and the files `smith new --dry-run` lists, and asks. The name of the
 * folder is used only after the prompt says so.
 * @param deps - how to run smith and ask the person
 * @returns how it ended; it never throws
 */
export async function createWorkspace (deps: NewWorkspaceDeps): Promise<NewWorkspaceOutcome> {
  const defaultLines = describeChoices({ name: deps.folderName, ...DEFAULT_CHOICES }, deps.folderName).slice(1)
  const way = await deps.askWay(defaultLines)
  if (way === undefined) return { kind: 'cancelled' }
  if (way === 'terminal') return { kind: 'terminal' }

  const name = await deps.askName(deps.folderName)
  if (name === undefined) return { kind: 'cancelled' }
  const start: NewWorkspaceChoices = { name, ...DEFAULT_CHOICES }
  const choices = way === 'customise' ? await deps.askCustom(start) : start
  if (choices === undefined) return { kind: 'cancelled' }

  const preview = await deps.run(toCreateArguments(choices, true))
  if (preview.code !== 0) return { kind: 'refused', message: describeFailure(preview) }
  if (!await deps.confirm(describeChoices(choices, deps.folderName), preview.stdout.trim())) return { kind: 'cancelled' }

  const done = await deps.apply(toCreateArguments(choices))

  return done.code === 0 ? { kind: 'created', name: choices.name, lines: done.stdout.split(/\r?\n/).map((line) => line.trim()).filter(Boolean), committed: choices.commit } : { kind: 'refused', message: describeFailure(done) }
}
