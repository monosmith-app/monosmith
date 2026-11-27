/** One value of a setting with a fixed set of values. */
export interface SettingOption {
  value:        string
  description?: string
  /** The value the setting takes when smith.toml does not set it. */
  default?:     boolean
}

/** A setting of smith.toml, as `config.list` (`smith config list --json`) answers it. */
export interface WorkspaceSetting {
  /** `table.key`, such as `ci.devcontainer`. */
  path:        string
  /** The schema type: `boolean`, `string`, `integer`, `array`. */
  type:        string
  description: string
  /** Why the default is what it is. */
  reason?:     string
  /** The value as smith.toml writes it; absent while it has its default. */
  value?:      string
  set:         boolean
  options?:    SettingOption[]
}

/** A plugin, as `plugins` (`smith plugin list --json`) answers it. */
export interface PluginEntry {
  name:    string
  /** `built-in`, or where an external plugin comes from. */
  source?: string
  roles?:  string[]
}

/** A recipe of the workspace, as `recipe.list` answers it. */
export interface RecipeEntry {
  path:     string
  name?:    string
  version?: string
  /** Set when the file is not a recipe. */
  error?:   string
}

/** What the Settings view shows of one folder. */
export interface SettingsSections {
  settings: WorkspaceSetting[]
  plugins:  PluginEntry[]
  recipes:  RecipeEntry[]
  /** What could not be read, one line each. */
  errors:   string[]
}
