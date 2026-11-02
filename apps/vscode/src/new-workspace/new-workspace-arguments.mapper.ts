import type { ChoiceLabel, NewWorkspaceChoices } from './new-workspace.contract.ts'

/** What "Start with the defaults" chooses: the same defaults as `smith new --yes` and the TUI. */
export const DEFAULT_CHOICES: Omit<NewWorkspaceChoices, 'name'> = {
  ci:           ['github'],
  language:     'later',
  nodeStack:    'node-jest',
  architecture: 'vertical-slices',
  agent:        true,
  devcontainer: true,
  commit:       true,
}

/** The CI systems `--ci` accepts. */
export const CI_CHOICES: ChoiceLabel[] = [
  { value: 'github', label: 'GitHub Actions', description: 'pipelines in .github/workflows' },
  { value: 'azure', label: 'Azure Pipelines', description: 'azure-pipelines.yml' },
  { value: 'gitlab', label: 'GitLab CI', description: '.gitlab-ci.yml' },
  { value: 'circleci', label: 'CircleCI', description: '.circleci/config.yml' },
  { value: 'bitbucket', label: 'Bitbucket Pipelines', description: 'bitbucket-pipelines.yml' },
]

/** The main languages `--language` accepts. */
export const LANGUAGE_CHOICES: ChoiceLabel[] = [
  { value: 'later', label: 'Decide later', description: 'no language yet; you pick one for each project you add' },
  { value: 'node', label: 'TypeScript / JavaScript', description: 'node' },
  { value: 'python', label: 'Python', description: 'python' },
  { value: 'go', label: 'Go', description: 'go' },
  { value: 'dotnet', label: 'C# / .NET', description: 'dotnet' },
  { value: 'jvm', label: 'Java / Kotlin', description: 'jvm' },
  { value: 'rust', label: 'Rust', description: 'rust' },
]

/** The tests and lint stacks of Node projects. */
export const NODE_STACK_CHOICES: ChoiceLabel[] = [
  { value: 'node-jest', label: 'Jest', description: 'node-jest (the default)' },
  { value: 'node-vitest', label: 'Vitest', description: 'node-vitest' },
  { value: 'node-ts', label: 'TypeScript only', description: 'node-ts' },
]

/** The code layouts `--architecture` accepts. */
export const ARCHITECTURE_CHOICES: ChoiceLabel[] = [
  { value: 'vertical-slices', label: 'Vertical slices', description: 'one folder per outcome, files named by role (the default)' },
  { value: 'hexagonal', label: 'Hexagonal', description: 'ports and adapters' },
  { value: 'clean', label: 'Clean', description: 'clean architecture layers' },
  { value: 'ddd', label: 'Domain-driven design', description: 'domain, application, infrastructure' },
  { value: 'modular-monolith', label: 'Modular monolith', description: 'one module per business area' },
  { value: 'mvc', label: 'MVC', description: 'models, views, controllers' },
  { value: 'package-by-layer', label: 'Package by layer', description: 'one folder per technical layer' },
  { value: 'none', label: 'No layout check', description: 'smith checks nothing about the folders' },
]

function labelOf (choices: readonly ChoiceLabel[], value: string): string {
  return choices.find((choice) => choice.value === value)?.label ?? value
}

/**
 * The arguments of `smith new` for the choices. Every question has a flag, so smith
 * asks nothing more and decides nothing in silence. `--here` creates the workspace in
 * the (empty) folder the editor has open.
 * @param choices - what the person decided
 * @param preview - add `--dry-run`, which lists the files and writes nothing
 * @returns the arguments after `smith`
 */
export function toCreateArguments (choices: NewWorkspaceChoices, preview = false): string[] {
  return [
    'new', choices.name, '--here', '--yes',
    '--ci', choices.ci.join(','),
    '--language', choices.language,
    ...(choices.language === 'node' ? ['--default-stack', `node=${choices.nodeStack}`] : []),
    '--architecture', choices.architecture,
    choices.agent ? '--agent' : '--no-agent',
    choices.devcontainer ? '--devcontainer' : '--no-devcontainer',
    ...(choices.commit ? [] : ['--no-commit']),
    ...(preview ? ['--dry-run'] : []),
  ]
}

/**
 * What the choices mean, one line each, for the summary shown before smith
 * writes anything. The name line says where the name comes from.
 * @param choices - what the person decided
 * @param folderName - the name of the open folder
 * @returns the lines of the summary
 */
export function describeChoices (choices: NewWorkspaceChoices, folderName: string): string[] {
  return [
    `Name: ${choices.name}${choices.name === folderName ? ' (the name of this folder)' : ''}`,
    `CI: ${choices.ci.map((value) => labelOf(CI_CHOICES, value)).join(', ')}`,
    `Main language: ${labelOf(LANGUAGE_CHOICES, choices.language)}`,
    ...(choices.language === 'node' ? [`Tests and lint for Node projects: ${labelOf(NODE_STACK_CHOICES, choices.nodeStack)}`] : []),
    `Code layout smith checks: ${labelOf(ARCHITECTURE_CHOICES, choices.architecture)}`,
    `Agent files for Claude Code and VS Code: ${choices.agent ? 'yes' : 'no'}`,
    `Dev container: ${choices.devcontainer ? 'yes' : 'no'}`,
    `First commit of smith's own files: ${choices.commit ? 'yes, smith makes it' : 'no, the files stay uncommitted'}`,
  ]
}
