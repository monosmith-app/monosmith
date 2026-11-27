import * as vscode from 'vscode'
import { say } from '../../shared/words/index.ts'

/**
 * Asks whether to switch a plugin on or off; each answer shows the smith command it runs. Nothing
 * is chosen for the person: Escape answers nothing.
 * @param plugin - the plugin
 * @returns the command to run (`monosmith.pluginEnable` or `monosmith.pluginDisable`), or undefined
 */
export async function pickPluginSwitch (plugin: string): Promise<string | undefined> {
  const picked = await vscode.window.showQuickPick([
    { label: say('settingspage.switch.on'), detail: `$ smith plugin enable ${plugin}`, command: 'monosmith.pluginEnable' },
    { label: say('settingspage.switch.off'), detail: `$ smith plugin disable ${plugin}`, command: 'monosmith.pluginDisable' },
  ], { title: say('vscode.settings-view.switch-plugin', { plugin }), placeHolder: say('vscode.settings-view.switch-plugin-hint') })

  return picked?.command
}
