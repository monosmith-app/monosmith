import type { WorkspaceProject } from '../workspace/index.ts'
import type { RunOptions, RunPlan, RunTarget } from './run-plan.contract.ts'

/**
 * Turns what the person asked to run into the plan smith is given.
 * @param target - what was asked for
 * @param projects - the projects of the folder, to know each one's tasks
 * @param options - bail and no-cache
 * @returns the plan, or undefined when there is nothing to run
 */
export function toRunPlan (target: RunTarget, projects: readonly WorkspaceProject[], options: RunOptions): RunPlan | undefined {
  const flags = { ...(options.bail && { bail: true }), ...(options.noCache && { noCache: true }) }
  switch (target.kind) {
    case 'task': {
      return { tasks: [target.task], projects: [target.project], ...flags }
    }
    case 'projects': {
      const chosen = projects.filter((project) => target.projects.includes(project.name))
      const tasks = [...new Set(chosen.flatMap((project) => project.tasks))]

      return tasks.length === 0 ? undefined : { tasks, projects: chosen.filter((project) => project.tasks.length > 0).map((project) => project.name), ...flags }
    }
    case 'affected': {
      return { tasks: [target.task], affected: true, ...flags }
    }
    case 'all': {
      return { tasks: [target.task], ...flags }
    }
  }
}

/** Says what a plan runs, in a few words, for messages and menus. */
export function describePlan (plan: RunPlan): string {
  const where = plan.affected === true
    ? 'affected projects'
    : (plan.projects === undefined ? 'all projects' : plan.projects.join(', '))
  if (plan.tasks.length === 1 && plan.projects?.length === 1 && plan.affected !== true) return `${plan.projects[0]}:${plan.tasks[0]}`
  const what = plan.tasks.length === 1 ? plan.tasks[0] : `${plan.tasks.length} tasks`

  return `${what} in ${where.length > 60 ? `${plan.projects?.length ?? 0} projects` : where}`
}
