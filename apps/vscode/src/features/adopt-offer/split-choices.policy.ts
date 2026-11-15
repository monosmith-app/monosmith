import { say } from '../words/index.ts'

/** What the person decided about the split of workspace and project. */
export interface SplitAnswer {
  /** Development dependencies sent to the workspace manifest. */
  hoist:  readonly string[]
  /** What happens to the project's own ESLint setup; undefined when it was not asked. */
  eslint: 'shared' | 'keep' | undefined
}

/** One answer of a question, as the pick shows it. */
export interface SplitChoice<T extends string> {
  value:       T
  label:       string
  description: string
}

/** What a question of the split asks: where a dependency goes, or what happens to ESLint. */
export interface SplitQuestion<T extends string> {
  title:   string
  choices: readonly SplitChoice<T>[]
}

/**
 * The question about one development dependency smith cannot place. The
 * first answer changes the least: the dependency stays in the project.
 * @param name - the package
 */
export function dependencyQuestion (name: string): SplitQuestion<'project' | 'workspace'> {
  return {
    title:   say('vscode.adopt-offer.dependency-question', { name }),
    choices: [
      { value: 'project', label: say('vscode.adopt-offer.dependency-project'), description: say('vscode.adopt-offer.dependency-project-detail') },
      { value: 'workspace', label: say('vscode.adopt-offer.dependency-workspace'), description: say('vscode.adopt-offer.dependency-workspace-detail') },
    ],
  }
}

/** The question about the project's own ESLint setup. The first answer changes the least: it stays. */
export function lintQuestion (): SplitQuestion<'keep' | 'shared'> {
  return {
    title:   say('vscode.adopt-offer.lint-question'),
    choices: [
      { value: 'keep', label: say('vscode.adopt-offer.lint-keep'), description: say('vscode.adopt-offer.lint-keep-detail') },
      { value: 'shared', label: say('vscode.adopt-offer.lint-shared'), description: say('vscode.adopt-offer.lint-shared-detail') },
    ],
  }
}

/**
 * Turns the answers into a SplitAnswer.
 * @param dependencies - each dependency asked about with where it goes
 * @param lint - the answer about ESLint, or undefined when the project has none
 */
export function toSplitAnswer (dependencies: ReadonlyArray<{ name: string, goesTo: 'project' | 'workspace' }>, lint: 'keep' | 'shared' | undefined): SplitAnswer {
  return { hoist: dependencies.filter((d) => d.goesTo === 'workspace').map((d) => d.name), eslint: lint }
}
