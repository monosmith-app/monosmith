import * as vscode from 'vscode'
import { normalizeSettings } from './normalize-settings.mapper.ts'
import type { MonoSmithSettings } from './settings.contract.ts'

/** Reads the settings of the extension as they are now. */
export function readSettings (): MonoSmithSettings {
  const config = vscode.workspace.getConfiguration('monosmith')

  return normalizeSettings({
    refreshAuto:       config.get('refresh.auto'),
    refreshDebounceMs: config.get('refresh.debounceMs'),
    parallel:          config.get('run.parallel'),
    problems:          config.get('output.problems'),
    bail:              config.get('run.bail'),
    noCache:           config.get('run.noCache'),
  })
}
