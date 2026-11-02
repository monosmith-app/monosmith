/** The language most projects of the workspace use, and the words that say why it comes first. */
export interface MainLanguage {
  language: string
  why:      string
}

/**
 * The language most projects of the workspace use among the ones that fit what is being built.
 * It starts the language list, as the TUI's Add form starts on it; a tie goes to the language that comes first in the list.
 * @param projects - the workspace's projects
 * @param fitting - the languages that fit what is being built
 * @returns the language and the line that says why, or undefined when no project uses a fitting language
 */
export function mainLanguage (projects: ReadonlyArray<{ language: string }>, fitting: readonly string[]): MainLanguage | undefined {
  const counts = new Map<string, number>()
  for (const project of projects) {
    if (fitting.includes(project.language)) counts.set(project.language, (counts.get(project.language) ?? 0) + 1)
  }
  let best: string | undefined
  for (const language of fitting) {
    const count = counts.get(language) ?? 0
    if (count > 0 && count > (counts.get(best ?? '') ?? 0)) best = language
  }

  return best === undefined ? undefined : { language: best, why: `Preselected: ${counts.get(best)} of the ${projects.length} projects here use ${best}.` }
}

/**
 * Puts the main language first, so it is the one the list starts on.
 * @param rows - the rows of the language question
 * @param main - the main language, when there is one
 */
export function startOnMainLanguage<T extends { language: string }> (rows: readonly T[], main: MainLanguage | undefined): T[] {
  return main === undefined ? [...rows] : [...rows.filter((row) => row.language === main.language), ...rows.filter((row) => row.language !== main.language)]
}
