import { appliesWhen, askableQuestions, type PickAnswer, type SmithQuestion } from '../smith-questions/index.ts'
import type { NewWorkspaceChoices } from './new-workspace.contract.ts'

/**
 * Asks the questions of `smith new` as smith lists them. A question is asked only when the answers
 * so far satisfy its `when`; advanced questions and questions asked at the end are skipped. Each
 * question starts on the value of `start`, or on smith's default for a question the choices do not hold.
 * @param questions - the questions from `smith new --questions --json`
 * @param start - the choices to start from (the defaults and the name)
 * @param pick - draws one question
 * @returns the choices with the answers, or undefined when the person cancelled at any question
 */
export async function askCustomAnswers (questions: readonly SmithQuestion[], start: NewWorkspaceChoices, pick: PickAnswer): Promise<NewWorkspaceChoices | undefined> {
  const answers: Record<string, string> = {
    'ci':           start.ci.join(','),
    'language':     start.language,
    'node-stack':   start.nodeStack,
    'architecture': start.architecture,
    'agent':        start.agent ? 'yes' : 'no',
    'devcontainer': start.devcontainer ? 'yes' : 'no',
  }
  for (const question of askableQuestions(questions)) {
    if (!appliesWhen(question, answers)) continue
    const current = (answers[question.key] ?? question.default).split(',').filter(Boolean)
    const picked = await pick(question, current)
    if (picked === undefined) return undefined
    answers[question.key] = picked.join(',')
  }

  return {
    ...start,
    ci:           (answers.ci ?? '').split(',').filter(Boolean),
    language:     answers.language ?? start.language,
    nodeStack:    answers['node-stack'] ?? start.nodeStack,
    architecture: answers.architecture ?? start.architecture,
    agent:        answers.agent === 'yes',
    devcontainer: answers.devcontainer === 'yes',
  }
}
