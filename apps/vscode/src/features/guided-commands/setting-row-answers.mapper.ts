/** The parts of a row of the Settings view that its commands pass on. */
interface SettingRow {
  setting?: unknown
  plugin?:  unknown
}

/**
 * What a row of the Settings view already answers: the setting a change or a reset is about,
 * or the plugin to switch. Those questions are not asked again; the value still is, and the
 * confirmation still shows the command.
 * @param id - the id of the guided command
 * @param argument - what the editor passed to the command
 * @returns the answers by question, empty when the command did not start on a Settings row
 */
export function settingRowAnswers (id: string, argument: unknown): Record<string, string> {
  const row = (typeof argument === 'object' && argument !== null ? argument : {}) as SettingRow
  if ((id === 'configSet' || id === 'configUnset') && typeof row.setting === 'string' && row.setting !== '') return { setting: row.setting }
  if ((id === 'pluginEnable' || id === 'pluginDisable') && typeof row.plugin === 'string' && row.plugin !== '') return { plugin: row.plugin }

  return {}
}
