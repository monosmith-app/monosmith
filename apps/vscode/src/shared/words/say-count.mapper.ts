import { say, type WordKey } from './smith-words.config.ts'

/**
 * The text of a plural pair of the catalogue, `<key>.one` for a count of one and `<key>.other`
 * for any other, with `{count}` filled, as `words.Count` does in the terminal interface.
 * @param key - the key without its `.one` or `.other` ending
 * @param count - the number the text counts
 * @param values - the other placeholders, by name
 * @returns the text
 */
export function sayCount (key: string, count: number, values: Record<string, string | number> = {}): string {
  return say(`${key}.${count === 1 ? 'one' : 'other'}` as WordKey, { count, ...values })
}
