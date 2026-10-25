import type { PluginRef } from '../protocol/index.ts'

/** A requirement on the project's language plugin: `requiresLanguage('node')`. */
export function requiresLanguage (name: string): PluginRef {
  return { level: 'language', name }
}

/** A plugin that must not be in the same project. */
export function conflictsWith (name: string): PluginRef {
  return { name }
}

/** Whether the reference names nothing: every member of a reference is optional, and at least one must be set. */
export function isEmptyRef (ref: PluginRef): boolean {
  return (ref.level ?? '' as string) === '' && (ref.name ?? '') === '' && (ref.slot ?? '') === ''
}

/** How a message names it: `language plugin node`, `plugin providing test`. */
export function describeRef (ref: PluginRef): string {
  const level: string = ref.level ?? ''
  const parts = [level === '' ? 'plugin' : `${level} plugin`]
  if (ref.name !== undefined && ref.name !== '') parts.push(ref.name)
  if (ref.slot !== undefined && ref.slot !== '') parts.push(`providing ${ref.slot}`)

  return parts.join(' ')
}
