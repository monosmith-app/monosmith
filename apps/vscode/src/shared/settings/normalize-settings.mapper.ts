import { DEBOUNCE_LIMITS, type MonoSmithSettings } from './settings.contract.ts'

/** A setting as the user wrote it: anything, since `settings.json` is edited by hand. */
export type RawSettings = Partial<Record<keyof MonoSmithSettings, unknown>>

/**
 * Turns what the user wrote into settings the extension can rely on: a wrong
 * type falls back to the default, a number is held between its limits.
 */
export function normalizeSettings (raw: RawSettings): MonoSmithSettings {
  return {
    refreshAuto:       flag(raw.refreshAuto, true),
    refreshDebounceMs: clamp(raw.refreshDebounceMs, DEBOUNCE_LIMITS.min, DEBOUNCE_LIMITS.max, DEBOUNCE_LIMITS.fallback),
    parallel:          Math.floor(clamp(raw.parallel, 0, 256, 0)),
    problems:          flag(raw.problems, true),
    bail:              flag(raw.bail, false),
    noCache:           flag(raw.noCache, false),
  }
}

function flag (value: unknown, fallback: boolean): boolean {
  return typeof value === 'boolean' ? value : fallback
}

function clamp (value: unknown, min: number, max: number, fallback: number): number {
  return typeof value !== 'number' || Number.isNaN(value) ? fallback : Math.min(max, Math.max(min, value))
}
