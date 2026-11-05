package commands

import "github.com/monosmith-app/monosmith/libs/words/words"

// The keymap: the one list the keys, the footer hints, the help (F1 and ?)
// and the command palette are all drawn from, so a label cannot go stale.
// Every key is a single key, nothing needs a modifier, and none works while a
// text field is editing (see ByKey).

// Group names, in the order of the help.
var (
	GroupMoving    = words.Say("commands.group.moving")
	GroupStart     = words.Say("commands.group.start")
	GroupRun       = words.Say("commands.group.run")
	GroupWorkspace = words.Say("commands.group.workspace")
	GroupProjects  = words.Say("commands.group.projects")
	GroupRelease   = words.Say("commands.group.release")
	GroupViews     = words.Say("commands.group.views")
	GroupHelp      = words.Say("commands.group.help")
)

// GroupOrder is the order of the groups in the help.
var GroupOrder = []string{GroupMoving, GroupStart, GroupRun, GroupWorkspace, GroupProjects, GroupRelease, GroupViews, GroupHelp}

// The views, as slices to scope a command to (see Command.Views).
var (
	vHome     = []string{TabHome}
	vProjects = []string{TabProjects}
	vDoctor   = []string{TabDoctor}
	vRelease  = []string{TabRelease}
	vDeps     = []string{TabDeps}
	vGuide    = []string{TabGuide}
	vSettings = []string{TabSettings}
	// vReloadable are the views that read something and can read it again.
	vReloadable = []string{TabHome, TabDoctor, TabRelease, TabDeps, TabSettings}
	// vWorkspace are the views of a workspace where the run keys are shown.
	vWorkspace = []string{TabHome, TabProjects, TabDoctor, TabDeps, TabRelease}
	vCopy      = []string{TabProjects, TabDoctor, TabDeps}
	// vSplit are the views whose selection has a detail pane beside the list.
	vSplit = []string{TabProjects, TabDoctor}
	// vSide are the views whose side panel scrolls with Ctrl+U and Ctrl+D: the
	// detail of Projects and Doctor, the policy of Deps and the option of Settings.
	vSide = []string{TabProjects, TabDoctor, TabDeps, TabSettings}
	// vSplitters are the views with a divider between panels.
	vSplitters = []string{TabProjects, TabDoctor, TabSettings}
	// vZoom are the views whose detail can fill the pane.
	vZoom = []string{TabProjects, TabDoctor}
)

// all is the keymap; the order is the order within each group and in the footer.
var all = []Command{
	// The keys of each view's own pane come first: they win over a global key
	// that is the same letter (Deps shows the major releases with M, which no other view uses).
	{ID: "home.next", Group: GroupRun, Title: words.Say("commands.home.next.title"), Keys: []string{"enter"}, Kind: PaneKey, Views: vHome, Footer: words.Say("commands.home.next.footer"), Glyph: words.Say("commands.home.next.glyph")},
	{ID: "home.move", Group: GroupRun, Title: words.Say("commands.home.move.title"), Keys: []string{"down", "up", "left", "right"}, Shortcut: words.Say("commands.home.move.shortcut"), Kind: PaneKey, Views: vHome, Footer: words.Say("commands.home.move.footer"), Glyph: words.Say("commands.home.move.glyph")},
	{ID: "home.scroll", Group: GroupRun, Title: words.Say("commands.home.scroll.title"), Keys: []string{"pgdown", "pgup"}, Shortcut: words.Say("commands.home.scroll.shortcut"), Kind: PaneKey, Views: vHome, Footer: words.Say("commands.home.scroll.footer"), Glyph: words.Say("commands.home.scroll.glyph")},
	{ID: "home.page-up", Group: GroupViews, Title: words.Say("commands.home.page-up.title"), Keys: []string{"ctrl+u"}, Kind: ViewAction, Action: "home-up", Views: vHome},
	{ID: "home.page-down", Group: GroupViews, Title: words.Say("commands.home.page-down.title"), Keys: []string{"ctrl+d"}, Kind: ViewAction, Action: "home-down", Views: vHome},

	// The keys that move the keyboard between the list and the detail come first in
	// the footer: a panel that has the keyboard says how to leave it.
	{ID: "side.focus", Group: GroupViews, Title: words.Say("commands.side.focus.title"), Keys: []string{"tab", "right"}, Shortcut: words.Say("commands.side.focus.shortcut"), Kind: PaneKey, Views: vSplit, NotInSideFocus: true, NeedsSideNext: true, Footer: words.Say("commands.side.focus.footer"), Glyph: words.Say("commands.side.focus.glyph")},
	{ID: "side.back", Group: GroupViews, Title: words.Say("commands.side.back.title"), Keys: []string{"tab", "left"}, Shortcut: words.Say("commands.side.back.shortcut"), Kind: PaneKey, Views: vSplit, NeedsSideFocus: true, Footer: words.Say("commands.side.back.footer"), Glyph: words.Say("commands.side.back.glyph")},
	{ID: "side.move", Group: GroupViews, Title: words.Say("commands.side.move.title"), Keys: []string{"down", "up"}, Shortcut: words.Say("commands.side.move.shortcut"), Kind: PaneKey, Views: vSplit, NeedsSideFocus: true, NotInRunPanel: true, Footer: words.Say("commands.side.move.footer"), Glyph: words.Say("commands.side.move.glyph")},
	{ID: "side.open", Group: GroupViews, Title: words.Say("commands.side.open.title"), Keys: []string{"enter"}, Kind: PaneKey, Views: vSplit, NeedsSideFocus: true, NotInRunPanel: true, Footer: words.Say("commands.side.open.footer"), Glyph: words.Say("commands.side.open.glyph")},

	// The run panel and the list of recent runs sit in the Projects view. They come
	// before the keys of the projects: while a run is open its r means "run again",
	// not "replace the derived file".
	{ID: "run.older", Group: GroupRun, Title: words.Say("commands.run.older.title"), Hint: words.Say("commands.run.older.hint"), Keys: []string{","}, Kind: ViewAction, Action: "run-older", Views: vProjects, NeedsRunPanel: true, NeedsIdle: true, Footer: words.Say("commands.run.older.footer")},
	{ID: "run.newer", Group: GroupRun, Title: words.Say("commands.run.newer.title"), Keys: []string{"."}, Kind: ViewAction, Action: "run-newer", Views: vProjects, NeedsRunPanel: true, NeedsIdle: true, Footer: words.Say("commands.run.newer.footer")},
	{ID: "run.all-output", Group: GroupRun, Title: words.Say("commands.run.all-output.title"), Keys: []string{"esc"}, Kind: PaneKey, Views: vProjects, NeedsRunPanel: true, Footer: words.Say("commands.run.all-output.footer"), TaskFooter: words.Say("commands.run.all-output.footer.task")},
	{ID: "run.scroll", Group: GroupRun, Title: words.Say("commands.run.scroll.title"), Keys: []string{"down", "up", "pgdown", "pgup"}, Shortcut: words.Say("commands.run.scroll.shortcut"), Glyph: words.Say("commands.run.scroll.glyph"), Kind: PaneKey, Views: vProjects, NeedsRunPanel: true, Footer: words.Say("commands.run.scroll.footer")},
	{ID: "run.again", Group: GroupRun, Title: words.Say("commands.run.again.title"), Hint: words.Say("commands.run.again.hint"), Keys: []string{"r"}, Kind: ViewAction, Action: "rerun", Views: vProjects, NeedsRunPanel: true, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.run.again.footer")},
	{ID: "run.again-failed", Group: GroupRun, Title: words.Say("commands.run.again-failed.title"), Hint: words.Say("commands.run.again-failed.hint"), Keys: []string{"R"}, Kind: ViewAction, Action: "rerun-failed", Views: vProjects, NeedsRunPanel: true, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.run.again-failed.footer")}, // words:literal
	{ID: "run.copy-log", Group: GroupRun, Title: words.Say("commands.run.copy-log.title"), Hint: words.Say("commands.run.copy-log.hint"), Keys: []string{"Y"}, Kind: ViewAction, Action: "copy-log", Views: vProjects, NeedsRunPanel: true, Footer: words.Say("commands.run.copy-log.footer")},                                                            // words:literal
	{ID: "run.follow", Group: GroupRun, Title: words.Say("commands.run.follow.title"), Hint: words.Say("commands.run.follow.hint"), Keys: []string{"f"}, Kind: Follow, Views: vProjects, NeedsRunPanel: true, Footer: words.Say("commands.run.follow.footer")},

	{ID: "projects.move", Group: GroupRun, Title: words.Say("commands.projects.move.title"), Keys: []string{"down", "up", "j", "k"}, Shortcut: words.Say("commands.projects.move.shortcut"), Kind: PaneKey, Views: vProjects, NotInRunPanel: true, NotInRecentFocus: true, NotInSideFocus: true},
	{ID: "projects.open", Group: GroupRun, Title: words.Say("commands.projects.open.title"), Keys: []string{"enter"}, Kind: PaneKey, Views: vProjects, NeedsNarrowList: true, NotInRunPanel: true, NotInRecentFocus: true, NotInSideFocus: true, Footer: words.Say("commands.projects.open.footer"), Glyph: words.Say("commands.projects.open.glyph")},
	{ID: "projects.install-package", Group: GroupProjects, Title: words.Say("commands.projects.install-package.title"), Hint: words.Say("commands.projects.install-package.hint"), Keys: []string{"+"}, Kind: OpenPackagePrompt, Action: "install", NeedsWorkspace: true, NeedsIdle: true},
	{ID: "projects.remove-package", Group: GroupProjects, Title: words.Say("commands.projects.remove-package.title"), Hint: words.Say("commands.projects.remove-package.hint"), Keys: []string{"-"}, Kind: OpenPackagePrompt, Action: "uninstall", NeedsWorkspace: true, NeedsIdle: true},
	{ID: "projects.back", Group: GroupRun, Title: words.Say("commands.projects.back.title"), Keys: []string{"esc"}, Kind: PaneKey, Views: vProjects, NeedsDetailOpen: true, NotInRunPanel: true, Footer: words.Say("commands.projects.back.footer")},
	{ID: "run.recent-move", Group: GroupRun, Title: words.Say("commands.run.recent-move.title"), Keys: []string{"down", "up"}, Shortcut: words.Say("commands.run.recent-move.shortcut"), Glyph: words.Say("commands.run.recent-move.glyph"), Kind: PaneKey, Views: vProjects, NeedsRecentFocus: true, Footer: words.Say("commands.run.recent-move.footer")},
	{ID: "run.open-recent", Group: GroupRun, Title: words.Say("commands.run.open-recent.title"), Keys: []string{"enter"}, Kind: PaneKey, Views: vProjects, NeedsRecentFocus: true, Footer: words.Say("commands.run.open-recent.footer"), Glyph: words.Say("commands.run.open-recent.glyph")},
	{ID: "run.recent-focus", Group: GroupRun, Title: words.Say("commands.run.recent-focus.title"), Keys: []string{"tab"}, Kind: PaneKey, Views: vProjects, NotInRunPanel: true, NotInSideFocus: true, NotSideNext: true, Footer: words.Say("commands.run.recent-focus.footer"), OnFooter: words.Say("commands.run.recent-focus.footer.on"), OnTitle: words.Say("commands.run.recent-focus.title.on")},
	{ID: "projects.explain", Group: GroupProjects, Title: words.Say("commands.projects.explain.title"), Hint: words.Say("commands.projects.explain.hint"), Keys: []string{"e"}, Kind: ViewAction, Action: "explain", Views: vProjects, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "projects.why", Group: GroupProjects, Title: words.Say("commands.projects.why.title"), Hint: words.Say("commands.projects.why.hint"), Keys: []string{"o"}, Kind: ViewAction, Action: "why", Views: vProjects, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "projects.graph", Group: GroupProjects, Title: words.Say("commands.projects.graph.title"), Hint: words.Say("commands.projects.graph.hint"), Keys: []string{"g"}, Kind: ViewAction, Action: "graph", Views: vProjects, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "projects.derive-explain", Group: GroupProjects, Title: words.Say("commands.projects.derive-explain.title"), Hint: words.Say("commands.projects.derive-explain.hint"), Keys: []string{"E"}, Kind: ViewAction, Action: "derive-explain", Views: vProjects, NeedsWorkspace: true, NeedsIdle: true}, // words:literal
	{ID: "projects.derive-apply", Group: GroupProjects, Title: words.Say("commands.projects.derive-apply.title"), Hint: words.Say("commands.projects.derive-apply.hint"), Keys: []string{"u"}, Kind: ViewAction, Action: "derive-apply", Views: vProjects, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "projects.derive-adopt", Group: GroupProjects, Title: words.Say("commands.projects.derive-adopt.title"), Hint: words.Say("commands.projects.derive-adopt.hint"), Keys: []string{"K"}, Kind: ViewAction, Action: "derive-adopt", Views: vProjects, NeedsWorkspace: true, NeedsIdle: true}, // words:literal
	{ID: "projects.derive-eject", Group: GroupProjects, Title: words.Say("commands.projects.derive-eject.title"), Hint: words.Say("commands.projects.derive-eject.hint"), Keys: []string{"O"}, Kind: ViewAction, Action: "derive-eject", Views: vProjects, NeedsWorkspace: true, NeedsIdle: true}, // words:literal
	{ID: "projects.derive-replace", Group: GroupProjects, Title: words.Say("commands.projects.derive-replace.title"), Hint: words.Say("commands.projects.derive-replace.hint"), Keys: []string{"r"}, Kind: ViewAction, Action: "derive-replace", Views: vProjects, NotInRunPanel: true, NeedsWorkspace: true, NeedsIdle: true},

	{ID: "doctor.open", Group: GroupWorkspace, Title: words.Say("commands.doctor.open.title"), Keys: []string{"enter"}, Kind: PaneKey, Views: vDoctor, NeedsNarrowList: true, Footer: words.Say("commands.doctor.open.footer"), Glyph: words.Say("commands.doctor.open.glyph")},
	{ID: "doctor.fix", Group: GroupWorkspace, Title: words.Say("commands.doctor.fix.title"), Keys: []string{"f"}, Kind: ViewAction, Action: "doctor-fix", Views: vDoctor, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.doctor.fix.footer")},
	{ID: "ws.fix", Group: GroupWorkspace, Title: words.Say("commands.ws.fix.title"), Hint: words.Say("commands.ws.fix.hint"), Keys: []string{"F"}, Kind: OpenFixWizard, Views: vDoctor, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.ws.fix.footer")}, // words:literal
	{ID: "doctor.back", Group: GroupWorkspace, Title: words.Say("commands.doctor.back.title"), Keys: []string{"esc"}, Kind: PaneKey, Views: vDoctor, NeedsDetailOpen: true, Footer: words.Say("commands.doctor.back.footer")},
	{ID: "doctor.page-up", Group: GroupWorkspace, Title: words.Say("commands.doctor.page-up.title"), Keys: []string{"pgup"}, Kind: ViewAction, Action: "detail-up", Views: vDoctor},
	{ID: "doctor.page-down", Group: GroupWorkspace, Title: words.Say("commands.doctor.page-down.title"), Keys: []string{"pgdown"}, Kind: ViewAction, Action: "detail-down", Views: vDoctor},
	{ID: "doctor.group", Group: GroupWorkspace, Title: words.Say("commands.doctor.group.title"), Keys: []string{"g"}, Kind: PaneKey, Views: vDoctor, Footer: words.Say("commands.doctor.group.footer"), OnFooter: words.Say("commands.doctor.group.footer.on"), OnTitle: words.Say("commands.doctor.group.title.on")},
	{ID: "doctor.foldall", Group: GroupWorkspace, Title: words.Say("commands.doctor.foldall.title"), Keys: []string{"E"}, Kind: PaneKey, Views: vDoctor, Footer: words.Say("commands.doctor.foldall.footer"), OnFooter: words.Say("commands.doctor.foldall.footer.on"), OnTitle: words.Say("commands.doctor.foldall.title.on")}, // words:literal

	{ID: "release.open", Group: GroupRelease, Title: words.Say("commands.release.open.title"), Keys: []string{"enter"}, Kind: PaneKey, Views: vRelease},
	{ID: "release.expand", Group: GroupRelease, Title: words.Say("commands.release.expand.title"), Keys: []string{"right", "left"}, Shortcut: words.Say("commands.release.expand.shortcut"), Kind: PaneKey, Views: vRelease, Footer: words.Say("commands.release.expand.footer"), Glyph: words.Say("commands.release.expand.glyph")},
	{ID: "rel.preview", Group: GroupRelease, Title: words.Say("commands.rel.preview.title"), Hint: words.Say("commands.rel.preview.hint"), Keys: []string{"p"}, Kind: Execute, Argv: []string{"release", "--dry-run"}, Views: vRelease, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.rel.preview.footer")},
	{ID: "rel.notes", Group: GroupRelease, Title: words.Say("commands.rel.notes.title"), Hint: words.Say("commands.rel.notes.hint"), Keys: []string{"n"}, Kind: ViewAction, Action: "release-notes", Views: vRelease, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.rel.notes.footer")},
	{ID: "rel.release", Group: GroupRelease, Title: words.Say("commands.rel.release.title"), Hint: words.Say("commands.rel.release.hint"), Keys: []string{"x"}, Kind: Execute, Argv: []string{"release"}, Confirm: true, Views: vRelease, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.rel.release.footer")},
	{ID: "rel.why", Group: GroupRelease, Title: words.Say("commands.rel.why.title"), Hint: words.Say("commands.rel.why.hint"), Keys: []string{"w"}, Kind: ViewAction, Action: "release-why", Views: vRelease, NeedsWorkspace: true, Yields: true, Footer: words.Say("commands.rel.why.footer")},
	{ID: "rel.first", Group: GroupRelease, Title: words.Say("commands.rel.first.title"), Hint: words.Say("commands.rel.first.hint"), Keys: []string{"c"}, Kind: ViewAction, Action: "release-first", Views: vRelease, NeedsWorkspace: true, NeedsIdle: true, Yields: true, Footer: words.Say("commands.rel.first.footer")},
	{ID: "rel.forced", Group: GroupRelease, Title: words.Say("commands.rel.forced.title"), Hint: words.Say("commands.rel.forced.hint"), Keys: []string{"f"}, Kind: Guided, Guide: words.Say("commands.rel.forced.guide"), Views: vRelease, NeedsWorkspace: true, NeedsIdle: true},

	{ID: "deps.mark", Group: GroupProjects, Title: words.Say("commands.deps.mark.title"), Keys: []string{" "}, Shortcut: words.Say("commands.deps.mark.shortcut"), Kind: PaneKey, Views: vDeps, Footer: words.Say("commands.deps.mark.footer")},
	{ID: "deps.filter-all", Group: GroupProjects, Title: words.Say("commands.deps.filter-all.title"), Keys: []string{"e"}, Kind: PaneKey, Views: vDeps},
	{ID: "deps.filter-patch", Group: GroupProjects, Title: words.Say("commands.deps.filter-patch.title"), Keys: []string{"p"}, Kind: PaneKey, Views: vDeps},
	{ID: "deps.filter-minor", Group: GroupProjects, Title: words.Say("commands.deps.filter-minor.title"), Keys: []string{"n"}, Kind: PaneKey, Views: vDeps},
	{ID: "deps.filter-major", Group: GroupProjects, Title: words.Say("commands.deps.filter-major.title"), Keys: []string{"M"}, Kind: PaneKey, Views: vDeps}, // words:literal
	{ID: "deps.update", Group: GroupProjects, Title: words.Say("commands.deps.update.title"), Hint: words.Say("commands.deps.update.hint"), Keys: []string{"u", "enter"}, Kind: ViewAction, Action: "deps-up", Views: vDeps, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.deps.update.footer")},
	{ID: "deps.open-page", Group: GroupProjects, Title: words.Say("commands.deps.open-page.title"), Hint: words.Say("commands.deps.open-page.hint"), Keys: []string{"g"}, Kind: ViewAction, Action: "deps-open-page", Views: vDeps, NeedsWorkspace: true, Footer: words.Say("commands.deps.open-page.footer")},
	{ID: "deps.copy-page", Group: GroupProjects, Title: words.Say("commands.deps.copy-page.title"), Keys: []string{"Y"}, Kind: ViewAction, Action: "deps-copy-page", Views: vDeps, NeedsWorkspace: true}, // words:literal
	{ID: "deps.install", Group: GroupProjects, Title: words.Say("commands.deps.install.title"), Hint: words.Say("commands.deps.install.hint"), Keys: []string{"i"}, Kind: ViewAction, Action: "deps-install", Views: vDeps, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.deps.install.footer")},
	{ID: "deps.remove", Group: GroupProjects, Title: words.Say("commands.deps.remove.title"), Hint: words.Say("commands.deps.remove.hint"), Keys: []string{"x"}, Kind: ViewAction, Action: "deps-remove", Views: vDeps, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.deps.remove.footer")},
	{ID: "deps.align", Group: GroupProjects, Title: words.Say("commands.deps.align.title"), Hint: words.Say("commands.deps.align.hint"), Keys: []string{"v"}, Kind: ViewAction, Action: "deps-align", Views: vDeps, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.deps.align.footer")},
	{ID: "deps.one-version", Group: GroupProjects, Title: words.Say("commands.deps.one-version.title"), Hint: words.Say("commands.deps.one-version.hint"), Keys: []string{"o"}, Kind: Execute, Argv: []string{"config", "set", "dependencies.one_version", "true"}, Confirm: true, Views: vDeps, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.deps.one-version.footer")},

	{ID: "settings.move", Group: GroupWorkspace, Title: words.Say("commands.settings.move.title"), Keys: []string{"down", "up", "j", "k"}, Shortcut: words.Say("commands.settings.move.shortcut"), Glyph: words.Say("commands.settings.move.glyph"), Kind: PaneKey, Views: vSettings, Footer: words.Say("commands.settings.move.footer")},
	{ID: "settings.change", Group: GroupWorkspace, Title: words.Say("commands.settings.change.title"), Keys: []string{"enter", " "}, Kind: PaneKey, Views: vSettings, Footer: words.Say("commands.settings.change.footer")},
	{ID: "settings.sections", Group: GroupWorkspace, Title: words.Say("commands.settings.sections.title"), Keys: []string{"tab", "left", "right"}, Shortcut: words.Say("commands.settings.sections.shortcut"), Glyph: words.Say("commands.settings.sections.glyph"), Kind: PaneKey, Views: vSettings, Footer: words.Say("commands.settings.sections.footer")},
	{ID: "settings.back", Group: GroupWorkspace, Title: words.Say("commands.settings.back.title"), Keys: []string{"esc"}, Kind: PaneKey, Views: vSettings, Footer: words.Say("commands.settings.back.footer")},
	{ID: "settings.trust", Group: GroupWorkspace, Title: words.Say("commands.settings.trust.title"), Hint: words.Say("commands.settings.trust.hint"), Keys: []string{"t"}, Kind: PaneKey, Views: vSettings, Footer: words.Say("commands.settings.trust.footer")},
	{ID: "settings.forget", Group: GroupWorkspace, Title: words.Say("commands.settings.forget.title"), Hint: words.Say("commands.settings.forget.hint"), Keys: []string{"f"}, Kind: PaneKey, Views: vSettings, Footer: words.Say("commands.settings.forget.footer")},
	{ID: "settings.unset", Group: GroupWorkspace, Title: words.Say("commands.settings.unset.title"), Hint: words.Say("commands.settings.unset.hint"), Keys: []string{"u"}, Kind: PaneKey, Views: vSettings, Footer: words.Say("commands.settings.unset.footer")},

	{ID: "guide.choose", Group: GroupStart, Title: words.Say("commands.guide.choose.title"), Keys: []string{"enter"}, Kind: PaneKey, Views: vGuide, Footer: words.Say("commands.guide.choose.footer"), Glyph: words.Say("commands.guide.choose.glyph")},
	{ID: "guide.back", Group: GroupStart, Title: words.Say("commands.guide.back.title"), Keys: []string{"esc"}, Kind: PaneKey, Views: vGuide, Footer: words.Say("commands.guide.back.footer")},

	{ID: "view.reload", Group: GroupViews, Title: words.Say("commands.view.reload.title"), Hint: words.Say("commands.view.reload.hint"), Keys: []string{"r"}, Kind: ViewAction, Action: "reload", Views: vReloadable, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.view.reload.footer")},

	// The filters of the Doctor headline come after reload in the footer, so a narrow footer drops them first.
	{ID: "doctor.failed", Group: GroupWorkspace, Title: words.Say("commands.doctor.failed.title"), Keys: []string{"x"}, Kind: PaneKey, Views: vDoctor, Footer: words.Say("commands.doctor.failed.footer"), OnFooter: words.Say("commands.doctor.failed.footer.on"), OnTitle: words.Say("commands.doctor.failed.title.on")},
	{ID: "doctor.warnings", Group: GroupWorkspace, Title: words.Say("commands.doctor.warnings.title"), Keys: []string{"w"}, Kind: PaneKey, Views: vDoctor, Footer: words.Say("commands.doctor.warnings.footer"), OnFooter: words.Say("commands.doctor.warnings.footer.on"), OnTitle: words.Say("commands.doctor.warnings.title.on")},
	{ID: "doctor.passed", Group: GroupWorkspace, Title: words.Say("commands.doctor.passed.title"), Keys: []string{"p"}, Kind: PaneKey, Views: vDoctor, Footer: words.Say("commands.doctor.passed.footer"), OnFooter: words.Say("commands.doctor.passed.footer.on"), OnTitle: words.Say("commands.doctor.passed.title.on")},

	// Moving: the keys every pane understands, handled by the pane itself.
	{ID: "move.arrows", Group: GroupMoving, Title: words.Say("commands.move.arrows.title"), Keys: []string{"down", "up", "j", "k"}, Shortcut: words.Say("commands.move.arrows.shortcut"), Kind: PaneKey},
	{ID: "move.enter", Group: GroupMoving, Title: words.Say("commands.move.enter.title"), Keys: []string{"enter"}, Kind: PaneKey},
	{ID: "move.esc", Group: GroupMoving, Title: words.Say("commands.move.esc.title"), Keys: []string{"esc"}, Kind: PaneKey},

	// Run
	{ID: "run.build", Group: GroupRun, Title: words.Say("commands.run.build.title"), Hint: words.Say("commands.run.build.hint"), Keys: []string{"b"}, Kind: RunTask, Task: "build", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"build"}, Footer: words.Say("commands.run.build.footer"), ScopeInFooter: true, ShowIn: vWorkspace},
	{ID: "run.build-all", Group: GroupRun, Title: words.Say("commands.run.build-all.title"), Hint: words.Say("commands.run.build-all.hint"), Keys: []string{"B"}, Kind: RunTask, Task: "build", All: true, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"build"}, Footer: words.Say("commands.run.build-all.footer"), ShowIn: vWorkspace}, // words:literal
	{ID: "run.test", Group: GroupRun, Title: words.Say("commands.run.test.title"), Hint: words.Say("commands.run.test.hint"), Keys: []string{"t"}, Kind: RunTask, Task: "test", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"test"}, Footer: words.Say("commands.run.test.footer"), ScopeInFooter: true, ShowIn: vWorkspace},
	{ID: "run.lint", Group: GroupRun, Title: words.Say("commands.run.lint.title"), Hint: words.Say("commands.run.lint.hint"), Keys: []string{"l"}, Kind: RunTask, Task: "lint", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"lint"}, Footer: words.Say("commands.run.lint.footer"), ScopeInFooter: true, ShowIn: vProjects},
	{ID: "run.typecheck", Group: GroupRun, Title: words.Say("commands.run.typecheck.title"), Hint: words.Say("commands.run.typecheck.hint"), Keys: []string{"c"}, Kind: RunTask, Task: "typecheck", NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"typecheck"}, Footer: words.Say("commands.run.typecheck.footer"), ScopeInFooter: true, ShowIn: vProjects},
	{ID: "run.affected", Group: GroupRun, Title: words.Say("commands.run.affected.title"), Hint: words.Say("commands.run.affected.hint"), Keys: []string{"w"}, Kind: Execute, Argv: []string{"run", "lint", "test", "--affected"}, NeedsWorkspace: true, NeedsIdle: true, Footer: words.Say("commands.run.affected.footer"), ShowIn: []string{TabHome, TabProjects}},
	{ID: "run.test-all", Group: GroupRun, Title: words.Say("commands.run.test-all.title"), Hint: words.Say("commands.run.test-all.hint"), Keys: []string{"T"}, Kind: RunTask, Task: "test", All: true, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"test"}},                          // words:literal
	{ID: "run.lint-all", Group: GroupRun, Title: words.Say("commands.run.lint-all.title"), Hint: words.Say("commands.run.lint-all.hint"), Keys: []string{"L"}, Kind: RunTask, Task: "lint", All: true, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"lint"}},                          // words:literal
	{ID: "run.typecheck-all", Group: GroupRun, Title: words.Say("commands.run.typecheck-all.title"), Hint: words.Say("commands.run.typecheck-all.hint"), Keys: []string{"C"}, Kind: RunTask, Task: "typecheck", All: true, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"typecheck"}}, // words:literal
	{ID: "run.verify", Group: GroupRun, Title: words.Say("commands.run.verify.title"), Hint: words.Say("commands.run.verify.hint"), Keys: []string{"V"}, Kind: Execute, Argv: []string{"run", "lint", "typecheck", "test"}, NeedsWorkspace: true, NeedsIdle: true},                                 // words:literal
	{ID: "run.stop", Group: GroupRun, Title: words.Say("commands.run.stop.title"), Hint: words.Say("commands.run.stop.hint"), Keys: []string{"ctrl+c"}, Kind: Stop, NeedsRunning: true, Footer: words.Say("commands.run.stop.footer"), ShowIn: vWorkspace},

	// Workspace
	{ID: "ws.doctor", Group: GroupWorkspace, Title: words.Say("commands.ws.doctor.title"), Hint: words.Say("commands.ws.doctor.hint"), Keys: []string{"d"}, Kind: OpenDoctor, NeedsWorkspace: true, Presents: []string{"doctor"}, Footer: words.Say("commands.ws.doctor.footer"), ShowIn: []string{TabHome, TabProjects}},
	{ID: "ws.ci", Group: GroupWorkspace, Title: words.Say("commands.ws.ci.title"), Hint: words.Say("commands.ws.ci.hint"), Kind: Execute, Argv: []string{"ci", "generate"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.info", Group: GroupWorkspace, Title: words.Say("commands.ws.info.title"), Hint: words.Say("commands.ws.info.hint"), Kind: Execute, Argv: []string{"info"}, NeedsIdle: true},
	{ID: "ws.undo", Group: GroupWorkspace, Title: words.Say("commands.ws.undo.title"), Hint: words.Say("commands.ws.undo.hint"), Kind: OpenUndo, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"undo"}},
	{ID: "ws.completion", Group: GroupWorkspace, Title: words.Say("commands.ws.completion.title"), Hint: words.Say("commands.ws.completion.hint"), Kind: OpenCompletion, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.completion-later", Group: GroupWorkspace, Title: words.Say("commands.ws.completion-later.title"), Hint: words.Say("commands.ws.completion-later.hint"), Kind: DeclineCompletion, NeedsWorkspace: true},
	{ID: "ws.report", Group: GroupWorkspace, Title: words.Say("commands.ws.report.title"), Hint: words.Say("commands.ws.report.hint"), Kind: OpenReport, NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"report"}},
	{ID: "ws.agent-file", Group: GroupWorkspace, Title: words.Say("commands.ws.agent-file.title"), Hint: words.Say("commands.ws.agent-file.hint"), Kind: Execute, Argv: []string{"agent", "file"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.commit", Group: GroupWorkspace, Title: words.Say("commands.ws.commit.title"), Hint: words.Say("commands.ws.commit.hint"), Kind: Execute, Argv: []string{"commit"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.tools", Group: GroupWorkspace, Title: words.Say("commands.ws.tools.title"), Hint: words.Say("commands.ws.tools.hint"), Kind: Execute, Argv: []string{"tools", "install"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.derive-check", Group: GroupWorkspace, Title: words.Say("commands.ws.derive-check.title"), Hint: words.Say("commands.ws.derive-check.hint"), Kind: Execute, Argv: []string{"derive", "--check"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.derive", Group: GroupWorkspace, Title: words.Say("commands.ws.derive.title"), Hint: words.Say("commands.ws.derive.hint"), Kind: Execute, Argv: []string{"derive"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.derive-preview", Group: GroupWorkspace, Title: words.Say("commands.ws.derive-preview.title"), Hint: words.Say("commands.ws.derive-preview.hint"), Kind: Execute, Argv: []string{"derive", "--dry-run"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.why", Group: GroupWorkspace, Title: words.Say("commands.ws.why.title"), Hint: words.Say("commands.ws.why.hint"), Kind: Guided, Guide: words.Say("commands.ws.why.guide"), NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"why"}},
	{ID: "ws.divergences", Group: GroupWorkspace, Title: words.Say("commands.ws.divergences.title"), Hint: words.Say("commands.ws.divergences.hint"), Kind: Execute, Argv: []string{"divergences"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.try", Group: GroupWorkspace, Title: words.Say("commands.ws.try.title"), Hint: words.Say("commands.ws.try.hint"), Kind: Guided, Guide: words.Say("commands.ws.try.guide"), NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"try"}},
	{ID: "ws.bootstrap", Group: GroupWorkspace, Title: words.Say("commands.ws.bootstrap.title"), Hint: words.Say("commands.ws.bootstrap.hint"), Kind: Execute, Argv: []string{"bootstrap"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.cache-status", Group: GroupWorkspace, Title: words.Say("commands.ws.cache-status.title"), Hint: words.Say("commands.ws.cache-status.hint"), Kind: Execute, Argv: []string{"cache", "status"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.cache-verify", Group: GroupWorkspace, Title: words.Say("commands.ws.cache-verify.title"), Hint: words.Say("commands.ws.cache-verify.hint"), Kind: Execute, Argv: []string{"cache", "verify"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.cache-repair", Group: GroupWorkspace, Title: words.Say("commands.ws.cache-repair.title"), Hint: words.Say("commands.ws.cache-repair.hint"), Kind: Execute, Argv: []string{"cache", "verify", "--fix"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.cache-prune", Group: GroupWorkspace, Title: words.Say("commands.ws.cache-prune.title"), Hint: words.Say("commands.ws.cache-prune.hint"), Kind: Execute, Argv: []string{"cache", "prune"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.cache-clear", Group: GroupWorkspace, Title: words.Say("commands.ws.cache-clear.title"), Hint: words.Say("commands.ws.cache-clear.hint"), Kind: Execute, Argv: []string{"cache", "clear"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.copies", Group: GroupWorkspace, Title: words.Say("commands.ws.copies.title"), Hint: words.Say("commands.ws.copies.hint"), Kind: Execute, Argv: []string{"copies"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.tools-list", Group: GroupWorkspace, Title: words.Say("commands.ws.tools-list.title"), Hint: words.Say("commands.ws.tools-list.hint"), Kind: Execute, Argv: []string{"tools"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.env-list", Group: GroupWorkspace, Title: words.Say("commands.ws.env-list.title"), Hint: words.Say("commands.ws.env-list.hint"), Kind: Execute, Argv: []string{"environments", "list"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "ws.env-add", Group: GroupWorkspace, Title: words.Say("commands.ws.env-add.title"), Hint: words.Say("commands.ws.env-add.hint"), Kind: Guided, Guide: words.Say("commands.ws.env-add.guide"), NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"environments add"}},                // words:literal
	{ID: "ws.env-remove", Group: GroupWorkspace, Title: words.Say("commands.ws.env-remove.title"), Hint: words.Say("commands.ws.env-remove.hint"), Kind: Guided, Guide: words.Say("commands.ws.env-remove.guide"), NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"environments remove"}}, // words:literal
	{ID: "ws.pipeline", Group: GroupWorkspace, Title: words.Say("commands.ws.pipeline.title"), Hint: words.Say("commands.ws.pipeline.hint"), Kind: Guided, Guide: words.Say("commands.ws.pipeline.guide"), NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"pipeline"}},
	{ID: "ws.tauri-icon", Group: GroupWorkspace, Title: words.Say("commands.ws.tauri-icon.title"), Hint: words.Say("commands.ws.tauri-icon.hint"), Kind: Guided, Guide: words.Say("commands.ws.tauri-icon.guide"), NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"tauri icon"}},                                         // words:literal
	{ID: "ws.docker-description", Group: GroupWorkspace, Title: words.Say("commands.ws.docker-description.title"), Hint: words.Say("commands.ws.docker-description.hint"), Kind: Guided, Guide: words.Say("commands.ws.docker-description.guide"), NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"docker description"}}, // words:literal
	{ID: "ws.why-miss", Group: GroupWorkspace, Title: words.Say("commands.ws.why-miss.title"), Hint: words.Say("commands.ws.why-miss.hint"), Kind: Guided, Guide: words.Say("commands.ws.why-miss.guide"), NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"why-miss"}},
	{ID: "ws.mcp-install", Group: GroupWorkspace, Title: words.Say("commands.ws.mcp-install.title"), Hint: words.Say("commands.ws.mcp-install.hint"), Kind: Guided, Guide: words.Say("commands.ws.mcp-install.guide"), NeedsWorkspace: true, NeedsIdle: true, Presents: []string{"mcp install"}}, // words:literal

	{ID: "cli.guide", Group: GroupHelp, Title: words.Say("commands.cli.guide.title"), Hint: words.Say("commands.cli.guide.hint"), Kind: Execute, Argv: []string{"help"}, NeedsIdle: true},

	// Start
	{ID: "file.new", Group: GroupStart, Title: words.Say("commands.file.new.title"), Hint: words.Say("commands.file.new.hint"), Keys: []string{"n"}, Kind: Guided, Guide: words.Say("commands.file.new.guide"), OnlyOutside: true, Footer: words.Say("commands.file.new.footer"), ShowIn: vGuide},
	{ID: "file.add", Group: GroupStart, Title: words.Say("commands.file.add.title"), Hint: words.Say("commands.file.add.hint"), Keys: []string{"a"}, Kind: Guided, Guide: words.Say("commands.file.add.guide"), NeedsWorkspace: true, NeedsIdle: true, OnlyInside: true, Footer: words.Say("commands.file.add.footer"), ShowIn: vProjects},
	{ID: "file.compose", Group: GroupStart, Title: words.Say("commands.file.compose.title"), Hint: words.Say("commands.file.compose.hint"), Keys: []string{"A"}, Kind: OpenCompose, NeedsWorkspace: true, NeedsIdle: true, OnlyInside: true}, // words:literal
	{ID: "file.adopt", Group: GroupStart, Title: words.Say("commands.file.adopt.title"), Hint: words.Say("commands.file.adopt.hint"), Keys: []string{"a"}, Kind: OpenAdoptWizard, Views: vGuide, NeedsIdle: true, OnlyOutside: true, NeedsContent: true, Footer: words.Say("commands.file.adopt.footer")},

	// Projects
	{ID: "proj.install", Group: GroupProjects, Title: words.Say("commands.proj.install.title"), Hint: words.Say("commands.proj.install.hint"), Keys: []string{"i"}, Kind: Execute, Argv: []string{"install"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.affected", Group: GroupProjects, Title: words.Say("commands.proj.affected.title"), Hint: words.Say("commands.proj.affected.hint"), Kind: Execute, Argv: []string{"affected"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.graph", Group: GroupProjects, Title: words.Say("commands.proj.graph.title"), Hint: words.Say("commands.proj.graph.hint"), Kind: Execute, Argv: []string{"graph"}, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.plugins", Group: GroupProjects, Title: words.Say("commands.proj.plugins.title"), Hint: words.Say("commands.proj.plugins.hint"), Kind: Guided, Guide: words.Say("commands.proj.plugins.guide"), NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.upgrade", Group: GroupProjects, Title: words.Say("commands.proj.upgrade.title"), Hint: words.Say("commands.proj.upgrade.hint"), Kind: Execute, Argv: []string{"upgrade"}, Confirm: true, NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.plugin-search", Group: GroupProjects, Title: words.Say("commands.proj.plugin-search.title"), Hint: words.Say("commands.proj.plugin-search.hint"), Kind: Guided, Guide: words.Say("commands.proj.plugin-search.guide"), NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.swap", Group: GroupProjects, Title: words.Say("commands.proj.swap.title"), Hint: words.Say("commands.proj.swap.hint"), Kind: Guided, Guide: words.Say("commands.proj.swap.guide"), NeedsWorkspace: true, NeedsIdle: true},
	{ID: "proj.remove", Group: GroupProjects, Title: words.Say("commands.proj.remove.title"), Hint: words.Say("commands.proj.remove.hint"), Kind: Guided, Guide: words.Say("commands.proj.remove.guide"), NeedsWorkspace: true, NeedsIdle: true},

	// Release

	// Views: the number is the position in the tab row.
	{ID: "view.home", Group: GroupViews, Title: words.Say("commands.view.home.title"), Hint: words.Say("commands.view.home.hint"), Keys: []string{"1"}, Glyph: words.Say("commands.view.home.glyph"), Kind: ShowTab, Tab: TabHome, NeedsWorkspace: true, Footer: words.Say("commands.view.home.footer"), Tail: true},
	{ID: "view.projects", Group: GroupViews, Title: words.Say("commands.view.projects.title"), Keys: []string{"2"}, Kind: ShowTab, Tab: TabProjects, NeedsWorkspace: true, Presents: []string{"projects"}},
	{ID: "view.doctor", Group: GroupViews, Title: words.Say("commands.view.doctor.title"), Keys: []string{"3"}, Kind: ShowTab, Tab: TabDoctor, NeedsWorkspace: true},
	{ID: "view.deps", Group: GroupViews, Title: words.Say("commands.view.deps.title"), Hint: words.Say("commands.view.deps.hint"), Keys: []string{"4"}, Kind: ShowTab, Tab: TabDeps, NeedsWorkspace: true, Presents: []string{"up"}},
	{ID: "view.release", Group: GroupViews, Title: words.Say("commands.view.release.title"), Hint: words.Say("commands.view.release.hint"), Keys: []string{"5"}, Kind: ShowTab, Tab: TabRelease, NeedsWorkspace: true, Presents: []string{"release"}},
	// The Settings tab is where plugins, recipes and smith.toml settings are
	// listed and changed (smith plugin, recipe and config in the terminal).
	{ID: "view.settings", Group: GroupViews, Title: words.Say("commands.view.settings.title"), Hint: words.Say("commands.view.settings.hint"), Keys: []string{"6"}, Kind: ShowTab, Tab: TabSettings, NeedsWorkspace: true, Presents: []string{"plugin", "plugin enable", "plugin disable", "recipe", "config", "config set", "config unset"}}, // words:literal
	{ID: "detail.zoom", Group: GroupViews, Title: words.Say("commands.detail.zoom.title"), Hint: words.Say("commands.detail.zoom.hint"), Keys: []string{"z"}, Kind: ViewAction, Action: "detail-zoom", Views: vZoom, Footer: words.Say("commands.detail.zoom.footer")},
	{ID: "detail.expand", Group: GroupViews, Title: words.Say("commands.detail.expand.title"), Keys: []string{"e"}, Kind: ViewAction, Action: "detail-expand", Views: vDoctor},
	{ID: "split.right", Group: GroupViews, Title: words.Say("commands.split.right.title"), Hint: words.Say("commands.split.right.hint"), Keys: []string{"]"}, Kind: ViewAction, Action: "split-right", Views: vSplitters},
	{ID: "split.left", Group: GroupViews, Title: words.Say("commands.split.left.title"), Hint: words.Say("commands.split.left.hint"), Keys: []string{"["}, Kind: ViewAction, Action: "split-left", Views: vSplitters},
	{ID: "split.select", Group: GroupViews, Title: words.Say("commands.split.select.title"), Hint: words.Say("commands.split.select.hint"), Keys: []string{"\\", "|", "shift+left", "shift+right", "shift+up", "shift+down"}, Shortcut: words.Say("commands.split.select.shortcut"), Kind: PaneKey, Views: vSplitters},
	{ID: "detail.up", Group: GroupViews, Title: words.Say("commands.detail.up.title"), Keys: []string{"ctrl+u"}, Kind: ViewAction, Action: "detail-up", Views: vSide},
	{ID: "detail.down", Group: GroupViews, Title: words.Say("commands.detail.down.title"), Keys: []string{"ctrl+d"}, Kind: ViewAction, Action: "detail-down", Views: vSide},
	{ID: "view.copy", Group: GroupViews, Title: words.Say("commands.view.copy.title"), Hint: words.Say("commands.view.copy.hint"), Keys: []string{"y"}, Kind: CopyFocused, Footer: words.Say("commands.view.copy.footer"), ShowIn: vCopy},
	{ID: "view.mouse", Group: GroupViews, Title: words.Say("commands.view.mouse.title"), Hint: words.Say("commands.view.mouse.hint"), Keys: []string{"m"}, Kind: ToggleMouse, Footer: words.Say("commands.view.mouse.footer"), Tail: true},

	// Help
	{ID: "file.palette", Group: GroupHelp, Title: words.Say("commands.file.palette.title"), Hint: words.Say("commands.file.palette.hint"), Keys: []string{":", "ctrl+p"}, Kind: Palette, WhileEditing: true, Footer: words.Say("commands.file.palette.footer"), Tail: true},
	{ID: "file.find", Group: GroupHelp, Title: words.Say("commands.file.find.title"), Hint: words.Say("commands.file.find.hint"), Keys: []string{"/"}, Kind: Palette},
	{ID: "help.keys", Group: GroupHelp, Title: words.Say("commands.help.keys.title"), Hint: words.Say("commands.help.keys.hint"), Keys: []string{"?", "f1"}, Kind: Help, WhileEditing: true, Footer: words.Say("commands.help.keys.footer"), Tail: true},
	{ID: "file.quit", Group: GroupHelp, Title: words.Say("commands.file.quit.title"), Hint: words.Say("commands.file.quit.hint"), Keys: []string{"q"}, Kind: Quit, Footer: words.Say("commands.file.quit.footer"), Tail: true},
}

// All returns every command.
func All() []Command { return append([]Command(nil), all...) }

// init gives each command the shortcut its keys spell, so the label shown
// beside a control is always the key that triggers it.
func init() {
	for i := range all {
		if all[i].Shortcut == "" {
			all[i].Shortcut = keysLabel(all[i].Keys)
		}
	}
}
