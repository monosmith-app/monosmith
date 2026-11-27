/** The sections of the command list, in the order the terminal interface's palette shows them. */
export type PaletteSection = 'start' | 'build' | 'projects' | 'dependencies' | 'release' | 'health' | 'goto' | 'help'

/** Where a command of the extension sits in the command list, and the smith command line it stands for. */
export interface CommandLine {
  section: PaletteSection
  /** The smith command line the command runs (or opens in a terminal); absent for a command of the editor only. */
  line?:   string
}

/** The order of the sections. */
export const PALETTE_SECTIONS: readonly PaletteSection[] = ['start', 'build', 'projects', 'dependencies', 'release', 'health', 'goto', 'help']

/**
 * The smith command line of every command the extension contributes, by command id (without `monosmith.`).
 * A test checks that every command of `package.json` is here, so a new command must say what it runs.
 * `<…>` marks a part the command asks for, or takes from the row it starts on.
 */
export const COMMAND_LINES: Readonly<Record<string, CommandLine>> = {
  // Start
  'new':                       { section: 'start', line: 'smith new --here' },
  'adopt.start':               { section: 'start', line: 'smith' },
  'adopt.preview':             { section: 'start', line: 'smith adopt --dry-run' },
  'convert.choose':            { section: 'start', line: 'smith convert --questions' },
  'copyInstallCommand':        { section: 'start', line: 'npm i -D monosmith' },
  'installAgent':              { section: 'start', line: 'smith agent install --claude --vscode' },
  'installMcp':                { section: 'start', line: 'smith mcp install --vscode' },
  'talkToAgent':               { section: 'start' },
  'agentFile':                 { section: 'start', line: 'smith agent file' },
  'mcpInstall':                { section: 'start', line: 'smith mcp install <assistants>' },
  'commit':                    { section: 'start', line: 'smith commit' },
  'undo':                      { section: 'start', line: 'smith undo' },
  'try':                       { section: 'start', line: 'smith try <command>' },
  'bootstrap':                 { section: 'start', line: 'smith bootstrap' },
  'setUpTabCompletion':        { section: 'start', line: 'smith completion install' },
  'installTools':              { section: 'start', line: 'smith tools install' },
  'toolsList':                 { section: 'start', line: 'smith tools' },
  'info':                      { section: 'start', line: 'smith info' },
  'openReport':                { section: 'start', line: 'smith report' },
  'commands':                  { section: 'start' },
  // Build
  'buildAll':                  { section: 'build', line: 'smith run build' },
  'lintAll':                   { section: 'build', line: 'smith run lint' },
  'testChanged':               { section: 'build', line: 'smith run lint test --affected' },
  'runTask':                   { section: 'build', line: 'smith run <task> -p <project>' },
  'runSelected':               { section: 'build', line: 'smith run <task> -p <project>' },
  'runAllTasks':               { section: 'build', line: 'smith run <tasks> -p <project>' },
  'runAffected':               { section: 'build', line: 'smith run <task> --affected' },
  'runAll':                    { section: 'build', line: 'smith run <task>' },
  'rerunLast':                 { section: 'build', line: 'smith run <the last run again>' },
  'cancelRun':                 { section: 'build' },
  'showTaskOutput':            { section: 'build' },
  'runs.showLog':              { section: 'build' },
  'runs.open':                 { section: 'build' },
  'runs.errorsOnly.on':        { section: 'build' },
  'runs.errorsOnly.off':       { section: 'build' },
  'whyMiss':                   { section: 'build', line: 'smith why-miss <task>' },
  'cacheSetup':                { section: 'build', line: 'smith cache setup' },
  'cacheStatus':               { section: 'build', line: 'smith cache status' },
  'cacheVerify':               { section: 'build', line: 'smith cache verify' },
  'cacheRepair':               { section: 'build', line: 'smith cache verify --fix' },
  'cachePrune':                { section: 'build', line: 'smith cache prune' },
  'cacheClear':                { section: 'build', line: 'smith cache clear' },
  // Projects
  'addProject':                { section: 'projects', line: 'smith add' },
  'composeStack':              { section: 'projects', line: 'smith compose' },
  'removeProject':             { section: 'projects', line: 'smith remove <project>' },
  'changeKind':                { section: 'projects', line: 'smith kind <project> <kind>' },
  'projectSettings':           { section: 'projects', line: 'smith config set -p <project> <setting> <value>' },
  'swap':                      { section: 'projects', line: 'smith swap <project> <choice>' },
  'newRecipe':                 { section: 'projects', line: 'smith recipe new' },
  'explainProject':            { section: 'projects', line: 'smith explain <project>' },
  'showGraph':                 { section: 'projects', line: 'smith graph --focus <project>' },
  'why':                       { section: 'projects', line: 'smith why <file or setting>' },
  'filter':                    { section: 'projects' },
  'clearFilter':               { section: 'projects' },
  'affectedOnly.on':           { section: 'projects', line: 'smith affected' },
  'affectedOnly.off':          { section: 'projects' },
  'environmentsList':          { section: 'projects', line: 'smith environments list' },
  'environmentsAdd':           { section: 'projects', line: 'smith environments add <app> <names>' },
  'environmentsRemove':        { section: 'projects', line: 'smith environments remove <app> <names>' },
  'pipeline':                  { section: 'projects', line: 'smith pipeline <name>' },
  'tauriIcon':                 { section: 'projects', line: 'smith tauri icon <image>' },
  'tauriAdd':                  { section: 'projects', line: 'smith tauri add <app>' },
  'dockerDescriptionPreview':  { section: 'projects', line: 'smith docker description --dry-run' },
  'dockerDescription':         { section: 'projects', line: 'smith docker description' },
  'architecture':              { section: 'projects', line: 'smith architecture <name>' },
  'ciGenerate':                { section: 'projects', line: 'smith ci generate' },
  'copies':                    { section: 'projects', line: 'smith copies' },
  // Dependencies
  'dependencies.refresh':      { section: 'dependencies', line: 'smith deps up --json' },
  'dependencies.up':           { section: 'dependencies', line: 'smith deps up' },
  'dependencies.update':       { section: 'dependencies', line: 'smith deps up <packages>' },
  'dependencies.updateSafe':   { section: 'dependencies', line: 'smith deps up <every patch and minor release>' },
  'dependencies.repair':       { section: 'dependencies', line: 'smith deps repair' },
  'dependencies.openRegistry': { section: 'dependencies' },
  'addDependency':             { section: 'dependencies', line: 'smith install -p <project> <package>' },
  'installPackage':            { section: 'dependencies', line: 'smith install -p <project> <package>' },
  'addLibrary':                { section: 'dependencies', line: 'smith install -p <project> workspace:<library>' },
  'unlinkLibrary':             { section: 'dependencies', line: 'smith uninstall -p <project> <library>' },
  'syncVersions':              { section: 'dependencies', line: 'smith sync' },
  // Release
  'release':                   { section: 'release', line: 'smith release' },
  'release.refresh':           { section: 'release', line: 'smith release --dry-run --json' },
  'release.dryRun':            { section: 'release', line: 'smith release --dry-run' },
  'release.notes':             { section: 'release', line: 'smith release --dry-run --notes' },
  // Health
  'doctor':                    { section: 'health', line: 'smith doctor' },
  'health.refresh':            { section: 'health', line: 'smith doctor' },
  'health.fix':                { section: 'health', line: 'smith doctor --fix' },
  'health.fixFinding':         { section: 'health', line: 'smith doctor --fix --only <check>' },
  'clearProblems':             { section: 'health' },
  'derive.explain':            { section: 'health', line: 'smith derive --explain=<file>' },
  'derive.apply':              { section: 'health', line: 'smith derive --replace=<file>' },
  'derive.adopt':              { section: 'health', line: 'smith derive --adopt=<file>' },
  'derive.eject':              { section: 'health', line: 'smith derive --eject=<file>' },
  'deriveCheck':               { section: 'health', line: 'smith derive --check' },
  'derivePreview':             { section: 'health', line: 'smith derive --dry-run' },
  'deriveUpdate':              { section: 'health', line: 'smith derive' },
  'divergences':               { section: 'health', line: 'smith divergences' },
  // Go to: the views, and the settings of smith.toml
  'settings.refresh':          { section: 'goto', line: 'smith config list --json' },
  'settings.change':           { section: 'goto', line: 'smith config set <setting> <value>' },
  'settings.reset':            { section: 'goto', line: 'smith config unset <setting>' },
  'settings.plugin':           { section: 'goto', line: 'smith plugin enable <plugin> · smith plugin disable <plugin>' },
  'settings.openRecipe':       { section: 'goto' },
  'configSet':                 { section: 'goto', line: 'smith config set <setting> <value>' },
  'configUnset':               { section: 'goto', line: 'smith config unset <setting>' },
  'pluginEnable':              { section: 'goto', line: 'smith plugin enable <plugin>' },
  'pluginDisable':             { section: 'goto', line: 'smith plugin disable <plugin>' },
  'pluginInstall':             { section: 'goto', line: 'smith plugin install <plugin>' },
  // Help: the extension itself
  'refresh':                   { section: 'help', line: 'smith projects --json' },
  'restart':                   { section: 'help', line: 'smith serve' },
  'showMenu':                  { section: 'help' },
  'showOutput':                { section: 'help' },
  'openSettings':              { section: 'help' },
  'openWalkthrough':           { section: 'help' },
}
