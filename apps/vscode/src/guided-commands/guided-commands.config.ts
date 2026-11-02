import type { GuidedAnswers, GuidedCommand } from './guided-command.contract.ts'

const text = (answers: GuidedAnswers, key: string): string => {
  const value = answers[key]

  return typeof value === 'string' ? value.trim() : ''
}
const list = (answers: GuidedAnswers, key: string): string[] => {
  const value = answers[key]

  return Array.isArray(value) ? value : []
}

/**
 * The commands of smith that have a screen of their own in the extension, beyond the flows that have
 * a folder of their own (add, swap, release...). Each is one palette command, `monosmith.<id>`. A
 * `read` command shows smith's output; a `change` command shows the plan or the consequence, asks,
 * and runs the real command with `--yes`.
 */
export const GUIDED_COMMANDS: readonly GuidedCommand[] = [
  {
    id:    'environmentsList',
    title: 'Environments of the Web Apps',
    mode:  'read',
    steps: [],
    args:  () => ['environments', 'list'],
  },
  {
    id:          'environmentsAdd',
    title:       'Add Environments to a Web App…',
    mode:        'change',
    planFlag:    '--dry-run',
    yes:         true,
    verb:        'Add the environments',
    consequence: 'Each name gets a .env.<name> file, a build:<name> task and an output folder. Nothing is committed.',
    steps:       [
      { key: 'project', kind: 'pick', title: 'MonoSmith: add environments (1/2): which web app?', from: 'web-apps' },
      { key: 'names', kind: 'text', title: 'MonoSmith: add environments (2/2): the names to add', prompt: 'Lowercase words, separated by commas. Edit the names, or write your own (such as qa, stage).', from: 'workspace-environments', validate: (value) => /^[a-z][a-z0-9, -]*$/.test(value.trim()) ? undefined : 'Use lowercase words separated by commas, for example qa, stage.' },
    ],
    args: (a) => ['environments', 'add', text(a, 'project'), ...text(a, 'names').split(/[\s,;]+/).filter(Boolean)],
  },
  {
    id:          'environmentsRemove',
    title:       'Remove Environments from a Web App…',
    mode:        'change',
    planFlag:    '--dry-run',
    yes:         true,
    verb:        'Remove the environments',
    consequence: 'The .env.<name> files stay, because they hold your values. The default environment cannot be removed.',
    steps:       [
      { key: 'project', kind: 'pick', title: 'MonoSmith: remove environments (1/2): which web app?', from: 'web-apps' },
      { key: 'names', kind: 'multi', title: 'MonoSmith: remove environments (2/2): which ones?', placeholder: 'Tick the environments to remove', from: 'environments-of-app' },
    ],
    args: (a) => ['environments', 'remove', text(a, 'project'), ...list(a, 'names')],
  },
  {
    id:    'cacheStatus',
    title: 'Cache: Status',
    mode:  'read',
    steps: [],
    args:  () => ['cache', 'status'],
  },
  {
    id:    'cacheVerify',
    title: 'Cache: Check for Damage',
    mode:  'read',
    steps: [],
    args:  () => ['cache', 'verify'],
  },
  {
    id:          'cacheRepair',
    title:       'Cache: Remove What Is Broken…',
    mode:        'change',
    steps:       [],
    verb:        'Remove what is broken',
    consequence: 'It deletes corrupt blobs, the entries that name them, unused blobs and leftover files. The next run of an affected task is slower.',
    args:        () => ['cache', 'verify', '--fix'],
  },
  {
    id:          'cachePrune',
    title:       'Cache: Trim to Its Limit…',
    mode:        'change',
    planFlag:    '--dry-run',
    steps:       [],
    verb:        'Trim the cache',
    consequence: 'It removes the least recently used entries.',
    args:        () => ['cache', 'prune'],
  },
  {
    id:          'cacheClear',
    title:       'Cache: Delete It All…',
    mode:        'change',
    yes:         true,
    steps:       [],
    verb:        'Delete the cache',
    consequence: 'It deletes the local cache. The remote cache stays, and the next run of every task is slower.',
    args:        () => ['cache', 'clear'],
  },
  {
    id:    'copies',
    title: 'Copies Between Projects',
    mode:  'read',
    steps: [],
    args:  () => ['copies'],
  },
  {
    id:    'toolsList',
    title: 'Tools the Workspace Needs',
    mode:  'read',
    steps: [],
    args:  () => ['tools'],
  },
  {
    id:    'whyMiss',
    title: 'Why Would a Task Run Again?…',
    mode:  'read',
    steps: [{ key: 'task', kind: 'text', title: 'MonoSmith: why would a task run again?', prompt: 'The project and the task, as project:task.', placeholder: 'web:build', validate: (value) => /^[^\s:]+:[^\s:]+$/.test(value.trim()) ? undefined : 'Write project:task, for example web:build.' }],
    args:  (a) => ['why-miss', text(a, 'task')],
  },
  {
    id:          'pipeline',
    title:       'Add a Ready-Made Pipeline…',
    mode:        'change',
    planFlag:    '--dry-run',
    yes:         true,
    verb:        'Write the pipeline',
    consequence: 'The file is yours from then on: smith upgrade never rewrites it. An existing file is kept.',
    steps:       [
      { key: 'template', kind: 'pick', title: 'MonoSmith: add a pipeline (1/2): which one?', from: 'pipeline-templates' },
      { key: 'project', kind: 'text', title: 'MonoSmith: add a pipeline (2/2): which app does it build?', prompt: 'Only a deploy pipeline builds an app. Enter leaves it out.', placeholder: 'web', optional: true },
    ],
    args: (a) => ['pipeline', text(a, 'template'), ...(text(a, 'project') === '' ? [] : ['--project', text(a, 'project')])],
  },
  {
    id:          'tauriIcon',
    title:       'Make the Icons of a Tauri App…',
    mode:        'change',
    planFlag:    '--dry-run',
    yes:         true,
    verb:        'Make the icons',
    consequence: 'It overwrites the icons the plan lists.',
    steps:       [
      { key: 'logo', kind: 'file', title: 'MonoSmith: the square PNG the icons are made from' },
      { key: 'project', kind: 'text', title: 'MonoSmith: which app?', prompt: 'Only when more than one project has a src-tauri. Enter leaves it out.', placeholder: 'desktop', optional: true },
    ],
    args: (a) => ['tauri', 'icon', text(a, 'logo'), ...(text(a, 'project') === '' ? [] : ['-p', text(a, 'project')])],
  },
  {
    id:    'dockerDescriptionPreview',
    title: 'Docker Hub Descriptions: Name the Repositories',
    mode:  'read',
    steps: [],
    args:  () => ['docker', 'description', '--dry-run'],
  },
  {
    id:          'dockerDescription',
    title:       'Set the Docker Hub Descriptions…',
    mode:        'change',
    verb:        'Set the descriptions',
    steps:       [],
    consequence: 'It calls Docker Hub with your login and replaces the description of each app\'s image. "Docker Hub Descriptions: Name the Repositories" shows which ones without calling it.',
    args:        () => ['docker', 'description'],
  },
  {
    id:          'mcpInstall',
    title:       'Connect smith to a Coding Assistant…',
    mode:        'change',
    yes:         true,
    verb:        'Write the files',
    consequence: 'It writes the file that lets the assistant start smith\'s MCP server. Nothing is committed.',
    steps:       [{
      key:         'tools',
      kind:        'multi',
      title:       'MonoSmith: which assistants?',
      placeholder: 'Tick the assistants to connect',
      options:     [
        { label: 'Claude Code', value: '--claude', description: 'writes .mcp.json' },
        { label: 'VS Code', value: '--vscode', description: 'writes .vscode/mcp.json' },
      ],
    }],
    args: (a) => ['mcp', 'install', ...list(a, 'tools')],
  },
]
