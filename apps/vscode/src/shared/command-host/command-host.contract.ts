import type { FolderRegistry, FolderState } from '../folder-session/index.ts'
import type { RefreshScope } from '../refresh/index.ts'
import type { SmithRun } from '../smith-run/index.ts'
import type { FolderWorkspace } from '../workspace/index.ts'

/** Runs smith with these arguments in the folder of one state. */
export type SmithRunner = (args: string[]) => Promise<SmithRun>

/** What a command of the editor may ask of the extension: its folders, its Output channel and its trust rules. */
/** The folder and the file, relative to it, that a derive or why command works on. */
export interface FileTarget {
  state: FolderState
  file:  string
}

export interface CommandHost {
  /** The folders of the extension. */
  registry:        FolderRegistry
  /** The folder that runs smith, picked by the person when there are several; undefined when none runs. */
  pickReadyFolder: () => Promise<FolderState | undefined>
  /** The projects of the folders that run smith. */
  readyWorkspaces: () => FolderWorkspace[]
  /** The folder and file a command works on: a row of the Health view, a file of the explorer or the editor, or the open editor. */
  fileTarget:      (argument: unknown) => FileTarget | undefined
  /** Shows an error in a notice and in the Output channel. */
  fail:            (error: unknown) => void
  /** Brings the Output channel to the front and gives it focus. */
  focusOutput:     () => void
  /** The folder a command works in: the one in the row it came from, else the one the person picks; undefined after saying why there is none. */
  folderFor:       (argument: unknown) => Promise<FolderState | undefined>
  /** Says a command that writes files needs a trusted workspace; true when it may go on. */
  mayWrite:        (what: string) => boolean
  /** The runner of smith in a folder; the caller checked that the folder has a smith. */
  smithIn:         (state: FolderState, log?: boolean, timeoutMs?: number) => SmithRunner
  /** The runner of smith in a folder that has no smith.toml yet: the smith to run, the folder, its name for the log, and a time limit in milliseconds. */
  smithAt:         (executable: string, root: string, name: string, timeoutMs: number) => SmithRunner
  /** Writes one line to the Output channel. */
  note:            (message: string) => void
  /** Brings the Output channel to the front. */
  showOutput:      () => void
  /** Reads the folder from smith again. */
  refreshFolder:   (root: string, scope: RefreshScope, quiet: boolean) => Promise<void>
  /** Writes how an operation ended to the Output channel: its kind, and the words of a refusal. */
  report:          (operation: string, state: FolderState, outcome: { kind: string, message?: string }) => void
}
