/** A list as Go's `%v` writes it, `[a b]`, which the details of the Go checks use; the TypeScript checks keep their text the same. */
export function goList (items: readonly unknown[]): string {
  return `[${items.map(String).join(' ')}]`
}

/** A string as Go's `%q` writes it. */
export function goQuote (text: string): string {
  return JSON.stringify(text)
}
