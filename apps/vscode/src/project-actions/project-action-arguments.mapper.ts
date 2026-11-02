import type { ProjectKind } from './project-action.contract.ts'

/**
 * The plan of removing a project: `smith remove <project> --dry-run`.
 * @param project - the project
 * @returns the arguments, without the program name
 */
export function removePlanArguments (project: string): string[] {
  return ['remove', project, '--dry-run']
}

/**
 * Removing a project for the answer of the person. smith never deletes files git cannot bring
 * back by surprise, so the answer is always named: keep a copy or delete.
 * @param project - the project
 * @param keep - `keep-copy` moves the folder to .smith/removed; `delete` deletes it for good
 * @returns the arguments, without the program name
 */
export function removeArguments (project: string, keep: 'keep-copy' | 'delete'): string[] {
  return ['remove', project, keep === 'delete' ? '--delete' : '--keep-copy', '--yes']
}

/**
 * The plan of changing the kind of a project: `smith kind <project> <kind> --dry-run`.
 * @param project - the project
 * @param kind - the kind it becomes
 * @returns the arguments, without the program name
 */
export function kindPlanArguments (project: string, kind: ProjectKind): string[] {
  return ['kind', project, kind, '--dry-run']
}

/**
 * Changing the kind of a project after the plan was shown.
 * @param project - the project
 * @param kind - the kind it becomes
 * @returns the arguments, without the program name
 */
export function kindArguments (project: string, kind: ProjectKind): string[] {
  return ['kind', project, kind, '--yes']
}

/**
 * Unlinking a library of this workspace from a project. `workspace:` names the library, so
 * smith never takes it for a registry package.
 * @param project - the project that stops using the library
 * @param library - the library
 * @returns the arguments, without the program name
 */
export function unlinkArguments (project: string, library: string): string[] {
  return ['uninstall', '-w', project, `workspace:${library}`, '--yes']
}

/**
 * The kinds a project can become: all but its own.
 * @param current - the kind of the project now (`app`, `lib` or `package`)
 * @returns the other kinds, in the order app, lib, package
 */
export function otherKinds (current: string): ProjectKind[] {
  return (['app', 'lib', 'package'] as const).filter((kind) => kind !== current)
}
