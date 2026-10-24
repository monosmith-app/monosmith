const PACKAGE_NAME = /^[\w@][\w./@-]*$/

/**
 * The arguments that update one package with `smith up`, or undefined when the
 * name could be read as an option or is not a package name at all. A name is
 * one argument, never a flag.
 */
export function updateArguments (name: string): string[] | undefined {
  return PACKAGE_NAME.test(name) ? ['up', name] : undefined
}

/** The arguments that let a person pick the updates in the terminal. */
export function pickArguments (): string[] {
  return ['up']
}
