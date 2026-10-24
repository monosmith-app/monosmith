import * as vscode from 'vscode'
import { createDiagnosticLineParser, type DiagnosticLineParser } from './parse-diagnostic.algorithm.ts'
import type { ParsedDiagnostic } from './parsed-diagnostic.contract.ts'

const SEVERITIES: Record<ParsedDiagnostic['severity'], vscode.DiagnosticSeverity> = {
  error:   vscode.DiagnosticSeverity.Error,
  warning: vscode.DiagnosticSeverity.Warning,
  info:    vscode.DiagnosticSeverity.Information,
}

/** The Problems panel entries that come from the output of tasks. */
export interface TaskDiagnostics extends vscode.Disposable {
  /** A task starts: what its last run reported is gone, and its output is read anew. */
  begin:    (key: string) => void
  /**
   * Reads one line of a task's output; a failure in it becomes a problem.
   * @param key - the task, as given to `begin`
   * @param resolve - finds the file a printed path means, or undefined when there is none
   */
  line:     (key: string, text: string, resolve: (file: string) => string | undefined) => void
  /** Removes every problem the tasks reported. */
  clearAll: () => void
}

/**
 * Collects the failures tasks print into a DiagnosticCollection, so they show
 * in the Problems panel and as squiggles. A task's problems are replaced when
 * it runs again, so a run that passes leaves none.
 */
export function createTaskDiagnostics (): TaskDiagnostics {
  const collection = vscode.languages.createDiagnosticCollection('MonoSmith')
  const parsers = new Map<string, DiagnosticLineParser>()
  const byTask = new Map<string, Map<string, vscode.Diagnostic[]>>()

  const publish = (): void => {
    const merged = new Map<string, vscode.Diagnostic[]>()
    for (const files of byTask.values()) {
      for (const [file, found] of files) merged.set(file, [...(merged.get(file) ?? []), ...found])
    }
    collection.clear()
    collection.set([...merged].map(([file, found]) => [vscode.Uri.file(file), found]))
  }

  let pending: NodeJS.Timeout | undefined
  const schedule = (): void => {
    pending ??= setTimeout(() => {
      pending = undefined
      publish()
    }, 100)
  }

  return {
    begin: (key) => {
      parsers.set(key, createDiagnosticLineParser())
      if (byTask.delete(key)) publish()
    },
    line: (key, text, resolve) => {
      const found = parsers.get(key)?.feed(text)
      if (found === undefined) return
      const file = resolve(found.file)
      if (file === undefined) return
      const at = new vscode.Position(Math.max(0, found.line - 1), Math.max(0, found.column - 1))
      const diagnostic = new vscode.Diagnostic(new vscode.Range(at, at.translate(0, 1)), found.message, SEVERITIES[found.severity])
      diagnostic.source = `smith ${key.split('\0').at(-1) ?? ''}`.trim()
      if (found.code !== undefined) diagnostic.code = found.code
      const files = byTask.get(key) ?? new Map<string, vscode.Diagnostic[]>()
      files.set(file, [...(files.get(file) ?? []), diagnostic])
      byTask.set(key, files)
      schedule()
    },
    clearAll: () => {
      clearTimeout(pending)
      pending = undefined
      parsers.clear()
      byTask.clear()
      collection.clear()
    },
    dispose: () => { clearTimeout(pending); collection.dispose() },
  }
}
