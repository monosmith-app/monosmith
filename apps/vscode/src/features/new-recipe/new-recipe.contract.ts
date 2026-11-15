/** One plugin a level can take, as `smith recipe new --list-levels --json` lists it. */
export interface RecipePlugin {
  name:        string
  /** The levels the plugin plays and what it covers. */
  description: string
}

/** A level of a recipe (language, scaffold, a task slot, structure, deploy) and the plugins that can fill it. */
export interface RecipeLevel {
  level:    string
  /** True when the level takes any number of plugins; false when it takes one. */
  additive: boolean
  plugins:  RecipePlugin[]
}

/** What smith offers a recipe: the levels, and for each language the plugins of the other levels that fit it. */
export interface RecipeLevels {
  levels: RecipeLevel[]
  /** `fits[language][level]` are the names of the plugins of the level that fit the language. */
  fits:   Record<string, Record<string, string[]>>
}

/** One question about a level. */
export interface LevelQuestion {
  level:    string
  /** Tick boxes (any number of plugins) when true; a pick list (one plugin) when false. */
  multi:    boolean
  /** The language plugin the recipe fixes, or empty when it fixes none. */
  language: string
  options:  RecipePlugin[]
}

/** The name, the version and the parent recipe the person typed. */
export interface RecipeBasics {
  name:    string
  version: string
  /** The recipe it extends (a path or an https URL); empty for none. */
  extends: string
}

/** Who decides a level: the recipe (one plugin is fixed) or the people who use it (they answer a question). */
export type LevelDecider = 'fixed' | 'chosen'

/** A level the people who use the recipe choose: the question they read, and the plugin each answer gives. */
export interface RecipeChoice {
  question: string
  answers:  Array<{ answer: string, plugin: string }>
}

/** What the person decided: the basics, the plugins the recipe fixes for each level, and the levels its users choose. */
export interface RecipeRequest extends RecipeBasics {
  fixed:   Record<string, string[]>
  /** The levels its users choose, by level name (`smith recipe new --chosen`). */
  chosen?: Record<string, RecipeChoice>
}

/** How writing a recipe ended: written (`file` is the path in the workspace), refused with smith's words, or cancelled by the person. */
export type RecipeOutcome =
  | { kind: 'written', request: RecipeRequest, file: string, lines: string[] } |
  { kind: 'refused', message: string } |
  { kind: 'cancelled' }
