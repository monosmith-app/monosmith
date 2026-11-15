/** One way to get a tool, as `smith tools install --dry-run --json` prints it. */
export interface ToolOption {
  /** What the option does: install with a package manager, use the official page, install it yourself, or skip. */
  kind:     'install' | 'page' | 'myself' | 'skip'
  /** The stable name of the option: `install:winget`, `page`, `myself`, `skip`. */
  value:    string
  /** The package manager of an install option. */
  manager?: string
  /** The install command of an install option: the program and its arguments. */
  command?: string[]
  /** True when the command needs administrator rights, or may ask the system for them. */
  admin?:   boolean
  label:    string
  /** The option and its consequence in plain words. */
  hint:     string
  /** A sentence the person needs before choosing. */
  note?:    string
}

/** A tool that is missing or too old, with the options smith offers for it. */
export interface ToolPlan {
  name:        string
  title:       string
  why:         string
  consequence: string
  download:    string
  state:       'missing' | 'outdated'
  options:     ToolOption[]
  /** The `value` of the marked option. */
  default:     string
}
