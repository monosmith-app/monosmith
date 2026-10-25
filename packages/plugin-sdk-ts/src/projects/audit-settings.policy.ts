import type { AuditSettings } from '../protocol/index.ts'

/** Whether the ecosystem's (`node`, `go`, `python`, `rust`) audit policy checks are on. */
export function auditEnabled (audit: AuditSettings | undefined, policy: string): boolean {
  return audit?.policies?.includes(policy) ?? false
}

/** Whether a failing check of the policy fails the audit: it does unless the workspace made the policy only warn. */
export function auditBlocks (audit: AuditSettings | undefined, policy: string): boolean {
  return !(audit?.warn?.includes(policy) ?? false)
}
