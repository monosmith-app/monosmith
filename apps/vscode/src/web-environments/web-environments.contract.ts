/** How a web app is built once per environment: what `smith add` asks and the flags answer. */
export interface WebEnvironments {
  names:    string[]
  /** The environment the task `build` builds, which CI and releases use. */
  default:  string
  /** `workspace`: dist/apps/<app>/<env>; `project`: dist/<env> in the app. */
  output:   'workspace' | 'project'
  /** `commit`: git keeps .env.<env>; `example`: git keeps only .env.example. */
  envFiles: 'commit' | 'example'
}

/** The parts of a request to add a project that tell whether it makes a web app. */
export interface WebAppChoice {
  role:     string
  language: string
  with:     string[]
  stack?:   string
}
