import type { SmithQuestion } from '../smith-questions/index.ts'
import type { NewWorkspaceChoices } from './new-workspace.contract.ts'

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

function labelOf (questions: readonly SmithQuestion[], key: string, value: string): string {
  return questions.find((question) => question.key === key)?.answers.find((answer) => answer.value === value)?.label ?? value
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
 * @param questions - the questions of `smith new`, which give each answer its label
 * @returns the lines of the summary
 */
export function describeChoices (choices: NewWorkspaceChoices, folderName: string, questions: readonly SmithQuestion[]): string[] {
  return [
    `Name: ${choices.name}${choices.name === folderName ? ' (the name of this folder)' : ''}`,
    `CI: ${choices.ci.map((value) => labelOf(questions, 'ci', value)).join(', ')}`,
    `Main language: ${labelOf(questions, 'language', choices.language)}`,
    ...(choices.language === 'node' ? [`Tests and lint for Node projects: ${labelOf(questions, 'node-stack', choices.nodeStack)}`] : []),
    `Code layout smith checks: ${labelOf(questions, 'architecture', choices.architecture)}`,
    `Agent files for Claude Code and VS Code: ${choices.agent ? 'yes' : 'no'}`,
    `Dev container: ${choices.devcontainer ? 'yes' : 'no'}`,
    `First commit of smith's own files: ${choices.commit ? 'yes, smith makes it' : 'no, the files stay uncommitted'}`,
  ]
}
