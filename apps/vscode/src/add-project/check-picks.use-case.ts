import type { SmithRun } from '../smith-run/index.ts'
import { longFlag } from '../smith-flags/index.ts'
import type { AddOptions } from './add-options.contract.ts'
import { fitsKind } from './add-options.mapper.ts'

/** What a layer holds, for the sentence that asks for one (the words of the terminal). */
const LAYER_WORDS: Record<string, string> = { framework: 'a framework', shell: 'a desktop shell', host: 'a cloud host' }

/** What the check of the picks is asked about. */
export interface PicksAsked {
  language:   string
  picks:      readonly string[]
  /** What is being built (`web`, `function`...); absent on the long way. */
  outcomeId?: string
  /** The kind of project: `app`, `lib` or `package`. */
  role?:      string
}

/**
 * Turns a sentence of `smith add` (which talks of `--with`) into one for a list
 * where nothing is typed as a flag; the same words as the terminal form.
 * @param text - what smith said is missing
 * @returns the sentence without flags
 */
export function plainWords (text: string): string {
  return text.replace(' to --with', '').replace('give no --with to start', 'untick everything to start')
}

/** What the check of smith holds: a conflict, what is missing, and whether it asks a question first. */
interface PicksCheckReport {
  conflict: string
  missing:  string
  asks:     boolean
}

function readCheck (stdout: string): PicksCheckReport | undefined {
  try {
    const root = JSON.parse(stdout) as { check?: { conflict?: unknown, missing?: unknown, question?: unknown } }
    const check = root.check
    if (check === undefined) return undefined

    return { conflict: typeof check.conflict === 'string' ? check.conflict : '', missing: typeof check.missing === 'string' ? check.missing : '', asks: check.question !== undefined && check.question !== null }
  } catch {
    return undefined
  }
}

/**
 * Checks the ticked choices before the list is left, as the terminal form does.
 * smith checks the set (`smith compose --json --with`, the solver of `smith add`):
 * lint and test choices alone are complete, because the stack's own scaffold is
 * filled in. What smith cannot know without the kind is whether this kind of
 * project needs a framework, a desktop shell or a cloud host: the outcome says
 * which layer, and the choices of that layer that fit the kind are named.
 * Nothing ticked is fine: the stack's defaults are used.
 * @param run - runs smith in the workspace folder
 * @param options - what smith lists
 * @param asked - the language, the picks, what is being built and the kind of project
 * @returns what is wrong or missing in plain words, or '' when the set is fine or smith cannot say (smith add has the last word)
 */
export async function checkPicks (run: (args: string[]) => Promise<SmithRun>, options: AddOptions, asked: PicksAsked): Promise<string> {
  if (asked.picks.length === 0) return ''
  const result = await run(['compose', longFlag.json, longFlag.lang, asked.language, '--with', asked.picks.join(',')])
  if (result.code !== 0) return ''
  const check = readCheck(result.stdout)
  if (check === undefined) return ''
  if (check.conflict !== '') return check.conflict
  if (check.missing !== '') return plainWords(check.missing)
  if (check.asks) return ''

  const layer = options.outcomes.find((outcome) => outcome.id === asked.outcomeId)?.items[0]?.layer
  if (layer === undefined || layer === '') return ''
  const have = options.choices.filter((choice) => (choice.language === undefined || choice.language === asked.language) && (choice.layer ?? choice.axis) === layer && fitsKind(choice, asked.role))
  const ticked = have.some((choice) => asked.picks.includes(choice.name))

  return ticked ? '' : `this kind of project needs ${LAYER_WORDS[layer] ?? `a ${layer}`}: tick one of ${have.map((choice) => choice.name).join(', ')}, or untick everything`
}
