import type { SmithConnection } from '../../integrations/smith-client/index.ts'
import type { PluginEntry, RecipeEntry, SettingsSections, WorkspaceSetting } from './settings-sections.contract.ts'

/** Why a part could not be read, or nothing when it was. */
function why (result: PromiseSettledResult<unknown>): string[] {
  if (result.status === 'fulfilled') return []

  return [result.reason instanceof Error ? result.reason.message : String(result.reason)]
}

/**
 * Asks smith for what the Settings view shows: every setting of smith.toml (`config.list`), the
 * plugins (`plugins`) and the recipes of the workspace (`recipe.list`). It writes nothing. A part
 * that cannot be read leaves its section empty and says why.
 * @param connection - the folder's smith serve
 * @returns the sections
 */
export async function readSettingsSections (connection: SmithConnection): Promise<SettingsSections> {
  const [settings, plugins, recipes] = await Promise.allSettled([
    connection.request<WorkspaceSetting[]>('config.list'),
    connection.request<PluginEntry[]>('plugins'),
    connection.request<{ recipes?: RecipeEntry[] }>('recipe.list'),
  ])

  return {
    settings: settings.status === 'fulfilled' && Array.isArray(settings.value) ? settings.value : [],
    plugins:  plugins.status === 'fulfilled' && Array.isArray(plugins.value) ? plugins.value : [],
    recipes:  recipes.status === 'fulfilled' ? (recipes.value.recipes ?? []) : [],
    errors:   [...why(settings), ...why(plugins), ...why(recipes)],
  }
}
