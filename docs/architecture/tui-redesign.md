# Design: the TUI redesign

Status: **in progress**. Phase 1 (theme and chrome) is built. The other phases are planned.
This page holds the owner's decisions, the colour tokens, the phases and the drawn screens.
The earlier design and its "as built" box are in [tui-and-ide.md](tui-and-ide.md), and the
code layout is in [tui-architecture.md](tui-architecture.md).

## Screen sizes and pieces

A wide screen is 150 x 44 cells and a narrow one 80 x 30. Every screen is drawn only with
Lip Gloss and Bubbles pieces (Bubble Tea 1.3, Bubbles 1.0, Lip Gloss 1.1).

## Decisions

The owner made these decisions for the whole redesign. Later phases do most of them.

| Decision | Phase |
|---|---|
| The palette is the MonoSmith one (table below). Plain mode does not change. | 1 |
| The title bar has the ` smith ` tag, the workspace in bold, the branch muted, the palette hint with its key and the smith version at the right. | 1 |
| The footer is the short style of `bubbles/help`: `key desc • key desc`, the key in the accent and bold, the description muted, the separator in the border colour. | 1 |
| The status bar has the last command (muted), the result counts with their symbols, the time, and the branch, the version and the doctor state at the right. | 1 |
| A wide screen has a rail of views at the left. A narrow screen has a one-line tab strip. One view shows at a time: `Enter` opens, `Esc` goes back. | 2 |
| A Runs view on key `3`. Today the runs show in the right panel of Projects. | 5 |
| The Doctor view is renamed Health, in the TUI and in the VS Code extension. The `smith doctor` command keeps its name. | 6 |
| The doctor fix shows inline in the detail (review, apply, check again), not as a dialog. | 6 |
| One pull request for each phase. | all |

## Tokens

The tokens are in `apps/smith/features/tui/theme/tokens_config.go`. They are hex colours, and
Lip Gloss makes them fit the terminal's colour profile. Each text colour keeps a contrast of
4.5:1 on the terminal's background, the selected row (Highlight), the row under the pointer
(RowHover) and the bars (Bar). `tokens_config_test.go` checks this.

| Token | Dark | Light | Use |
|---|---|---|---|
| Accent, Focus, Info | `#5fb6d1` | `#176a86` | selection, focus, the active tab, running |
| Success | `#7ccf96` | `#1b7537` | passed, cached, ready |
| Warning | `#e3c16f` | `#855600` | needs attention |
| Error | `#ff9a90` | `#b3261e` | failed, blocked |
| Text | `#d0d0d0` | `#2b2b2b` | normal text |
| Muted | `#b0b0b0` | `#5c5c5c` | secondary text, the footer descriptions, the last command |
| Faint | `#a5a5a5` | `#686868` | hints: a dimmer grey that still has 4.5:1 on the selected row |
| Highlight | `#2f3d44` | `#dcebf1` | background of the selected row |
| RowHover | `#263238` | `#eef5f8` | background of the row under the pointer, between the terminal's background and Highlight |
| Bar | `#303030` | `#f2f2f2` | background of the title bar and the status bar |
| Border | `#5a5a5a` | `#b8b8b8` | rules, pane borders, the footer separator |
| Primary button | Accent background, text `#141414`, bold | Accent background, text `#ffffff`, bold | `ButtonActive` and `ButtonOn`; also the ` smith ` tag |
| Secondary button | Highlight background, Text | Highlight background, Text | `Button` and `ButtonText` |
| Button under the pointer | `#3d5560` | `#a9cfdd` | `ButtonHover`, between the secondary and the primary background |

The palette gives `#1d7a3a` for the light Success. That colour has 4.4:1 on the light
Highlight, so the token is `#1b7537`, one step darker, with 4.7:1.

## Phases

1. Theme and chrome: the tokens, the title bar, the footer and the status bar. The layout
   keeps the top tabs. From 120 columns the header stays at the right end of the tab row
   without the palette hint and the version. Phase 2 gives the title bar its own row there.
2. Layout: the rail of views on a wide screen, the tab strip on a narrow one, one view at a
   time.
3. Home.
4. Projects.
5. The Runs view.
6. The command palette, and Health with the inline fix.
7. Then Dependencies, Release, Settings and the dialogs, once they are drawn.

## Screens

The screens below are plain-text renders of the design.

### Main

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Home ────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                      │ │                                                                                                                           │
│ ▶ Home             1 │ │  monosmith  42 projects · 7 changed since main · last run 3 min ago                                                       │
│                      │ │                                                                                                                           │
│   Projects         2 │ │  ╭───────────────────────────╮ ╭───────────────────────────╮ ╭───────────────────────────╮ ╭───────────────────────────╮  │
│                      │ │  │ Changed                   │ │ Last run                  │ │ Health                    │ │ Release                   │  │
│   Runs         ✗1  3 │ │  │ ● 7 projects              │ │ ✗ 1 failed                │ │ ⚠ 2 warnings              │ │ ✓ 2 ready                 │  │
│                      │ │  │ since main                │ │ 17 passed · 9 cached      │ │ 0 failed checks           │ │ 1 minor · 1 patch         │  │
│   Health       ⚠2  4 │ │  ╰───────────────────────────╯ ╰───────────────────────────╯ ╰───────────────────────────╯ ╰───────────────────────────╯  │
│                      │ │                                                                                                                           │
│   Dependencies ↑3  5 │ │  Needs you now  5 items                                                                                                   │
│                      │ │                                                                                                                           │
│   Release      ✓2  6 │ │  │ ✗ Failed   test failed in orders-api                                                                  See why enter    │
│                      │ │  │            2 of 140 tests · 4 min ago                                                                                  │
│   Settings         7 │ │                                                                                                                           │
│                      │ │    ● Changed  7 projects changed and not tested                                                         Test changed t    │
│                      │ │               web, kit, orders-api and 4 more                                                                             │
│                      │ │                                                                                                                           │
│ Pinned               │ │    ⚠ Warning  2 health checks warn                                                                            Review 4    │
│                      │ │               both have a fix that needs no decision                                                                      │
│ ✗ orders-api     app │ │                                                                                                                           │
│ ✓ web            app │ │    ↑ Updates  3 dependencies have newer versions                                                              Review 5    │
│ ● kit            lib │ │               1 major · 1 minor · 1 patch                                                                                 │
│                      │ │                                                                                                                           │
│                      │ │    ✓ Release  2 packages are ready to release                                                                Preview 6    │
│                      │ │               kit 1.4.0 · web 2.0.1                                                                                       │
│                      │ │                                                                                                                           │
│                      │ │  ───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────  │
│                      │ │                                                                                                                           │
│                      │ │  Start something                                                                                                          │
│                      │ │                                                                                                                           │
│                      │ │   Build changed b     Lint all l     Run a task… r     Add a project a     Settings 7                                     │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
│                      │ │  Tip: every action here is a smith command. Press y on any row to copy it.                                                │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
╰──────────────────────╯ ╰───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯

 ↑/↓ move • enter see why • t test changed • 1-7 views • ctrl+k commands • ? more • q quit
 $ smith test --changed  ✗ 1 failed  ✓ 17 passed  38s                                                                           main  ·  2.64.1  ·  ⚠2
```

### Projects

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Projects 7/42 ──────────────────────────╮╭─ orders-api ──────────────────────────────────────────────────────────────────╮
│                      │ │                                          ││                                                                               │
│   Home             1 │ │ Filter: changed kind:app▏                ││  apps/orders-api · app · C# (.NET 9)  ● changed         Run… r    Open o      │
│                      │ │ 4 of 12 apps · 3 of 30 libs              ││                                                                               │
│ ▶ Projects         2 │ │                                          ││   Tasks    Dependencies    Runs    Settings                                   │
│                      │ │ APPS                                     ││  ───────────────────────────────────────────────────────────────────────────  │
│   Runs         ✗1  3 │ │ ▶ ✗  orders-api         C#       ● changed│                                                                               │
│                      │ │   ✓  web                React    ● changed│  Task          Last result       Time       Cache                     Key     │
│   Health       ⚠2  4 │ │   ✓  ms-teams-bot       Go       ● changed│  ───────────────────────────────────────────────────────────────────────────  │
│                      │ │   –  billing-fn         Python   ● changed│  build         ✓ passed          4.1s       hit                       b       │
│   Dependencies ↑3  5 │ │ LIBS                                     ││  test          ✗ 2 failed        21.6s      miss · src changed        t       │
│                      │ │   ✓  kit                TS       ● changed│  lint          ✓ passed          2.3s       hit                       l       │
│   Release      ✓2  6 │ │   ✓  core               C#       ● changed│  package       – not run         –          –                         p       │
│                      │ │   ✓  words              Go       ● changed│  start         own terminal      –          never cached              s       │
│   Settings         7 │ │                                          ││                                                                               │
│                      │ │                                          ││  start runs in your terminal: smith leaves the screen and comes back when it en
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││  ╭─ Uses ────────────────────────────╮ ╭─ Used by ─────────────────────────╮  │
│ Pinned               │ │                                          ││  │                                   │ │                                   │  │
│                      │ │                                          ││  │ core · kit · words                │ │ web · billing-fn                  │  │
│ ✗ orders-api     app │ │                                          ││  │                                   │ │ a change here tests them too      │  │
│ ✓ web            app │ │                                          ││  ╰───────────────────────────────────╯ ╰───────────────────────────────────╯  │
│ ● kit            lib │ │                                          ││                                                                               │
│                      │ │                                          ││  Recent runs                                                                  │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││  ✗ test   2 failed   21.6s               4 min ago                            │
│                      │ │                                          ││  ✓ build  passed     4.1s (cache)        4 min ago                            │
│                      │ │                                          ││  ✓ test   passed     19.8s               yesterday                            │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │ ✓ passed  ✗ failed  – not run            ││                                                                               │
╰──────────────────────╯ ╰──────────────────────────────────────────╯╰───────────────────────────────────────────────────────────────────────────────╯

 / filter • enter run test • space select • tab next pane • ctrl+k commands • ? more
 $ smith run test -p orders-api                                                                                                 main  ·  2.64.1  ·  ⚠2
```

### Run

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Run · smith test --changed ──────────────────────────────────────────────────────────────────────────────────────────────╮
│                      │ │                                                                                                                           │
│   Home             1 │ │  ⣾ Running 18 tasks  00:24 · about 20s left                                             Copy command y     Stop ctrl+c    │
│                      │ │                                                                                                                           │
│   Projects         2 │ │  ████████████████████████████████████████████████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  67%       │
│                      │ │  ✗ 1 failed    ✓ 9 passed    ⚡ 2 cached    ◐ 2 running    ○ 4 queued                                                      │
│ ▶ Runs         ✗1  3 │ │                                                                                                                           │
│                      │ │───────────────────────────────────────┬───────────────────────────────────────────────────────────────────────────────────│
│   Health       ⚠2  4 │ │                                       │                                                                                   │
│                      │ │ Failed                                │ orders-api:test    Errors only   Full log   / search                              │
│   Dependencies ↑3  5 │ │ ├─ ✗ orders-api:test           21.6s  │                                                                                   │
│                      │ │ Running                               │   Restored 3 projects (0.4s)                                                    ┃ │
│   Release      ✓2  6 │ │ ├─ ⣾ web:test                    14s  │   Running 140 tests in Orders.Api.Tests                                         ┃ │
│                      │ │ ├─ ⣾ billing-fn:test              6s  │                                                                                 ┃ │
│   Settings         7 │ │ Queued · waits on web:test            │ ✗ OrderTotals_IncludeTax          tests/OrderTotalsTests.cs:42                  ┃ │
│                      │ │ ├─ ○ web-e2e:test                     │     Expected: 119.00                                                            ┃ │
│                      │ │ ├─ ○ docs:test                        │     Actual:   100.00                                                            ┃ │
│                      │ │ Done                                  │                                                                                 ┃ │
│ Pinned               │ │ ├─ ✓ kit:test                   3.2s  │ ✗ Refund_AfterShipping_IsPartial  tests/RefundTests.cs:88                       │ │
│                      │ │ ├─ ⚡ core:test                 cache  │     Expected status Partial, got Full                                           │ │
│ ✗ orders-api     app │ │ ├─ ✓ ms-teams-bot:test          1.9s  │                                                                                 │ │
│ ✓ web            app │ │ ├─ ✓ … 7 more                         │ ✓ 138 passed   ✗ 2 failed   21.6s                                               │ │
│ ● kit            lib │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │ ───────────────────────────────────────────────────────────────────────────────── │
│                      │ │                                       │ Cache miss: 2 files changed in src/Orders since the last pass.                    │
│                      │ │                                       │  Rerun failed R     Open file o                                                   │
╰──────────────────────╯ ╰───────────────────────────────────────┴───────────────────────────────────────────────────────────────────────────────────╯

 ↑/↓ task • e errors only • R rerun failed • y copy command • Y copy log • , . older/newer • esc back
 ◐ running  $ smith test --changed  00:24                                                                                       main  ·  2.64.1  ·  ⚠2
```

### Palette

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Home ────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                      │ │                                                                                                                           │
│ ▶ Home             1 │ │  monosmith  42 projects · 7 changed since main · last run 3 min ago                                                       │
│                      │ │           ╭─ Commands ───────────────────────────────────────────────────────────────╮                                    │
│   Projects         2 │ │  ╭────────│                                                                          │───╮ ╭───────────────────────────╮  │
│                      │ │  │ Changed│  › test orders▏                                                   esc    │   │ │ Release                   │  │
│   Runs         ✗1  3 │ │  │ ● 7 pro│──────────────────────────────────────────────────────────────────────────│   │ │ ✓ 2 ready                 │  │
│                      │ │  │ since m│  Best match                                                              │   │ │ 1 minor · 1 patch         │  │
│   Health       ⚠2  4 │ │  ╰────────│  ▶ Test orders-api                                                enter  │───╯ ╰───────────────────────────╯  │
│                      │ │           │    smith run test -p orders-api                                          │                                    │
│   Dependencies ↑3  5 │ │  Needs you│  ▶ Test orders-api and what uses it                                      │                                    │
│                      │ │           │    smith run test -p orders-api --dependents                             │                                    │
│   Release      ✓2  6 │ │  │ ✗ Faile│  ↻ Rerun the 2 failed tests                                           R  │                   See why enter    │
│                      │ │  │        │    from the last run · 4 min ago                                         │                                    │
│   Settings         7 │ │           │  Projects                                                                │                                    │
│                      │ │    ● Chang│  ✗ orders-api                                                            │                  Test changed t    │
│                      │ │           │    app · C# · apps/orders-api                                            │                                    │
│                      │ │           │  Settings                                                                │                                    │
│ Pinned               │ │    ⚠ Warni│  ⚙ orders-api: test timeout                                              │                        Review 4    │
│                      │ │           │    smith config set -p orders-api tasks.test.timeout                     │                                    │
│ ✗ orders-api     app │ │           │                                                                          │                                    │
│ ✓ web            app │ │    ↑ Updat│                                                                          │                        Review 5    │
│ ● kit            lib │ │           │──────────────────────────────────────────────────────────────────────────│                                    │
│                      │ │           │  ↑/↓ move • enter run • tab edit the command first                       │                                    │
│                      │ │    ✓ Relea╰──────────────────────────────────────────────────────────────────────────╯                       Preview 6    │
│                      │ │               kit 1.4.0 · web 2.0.1                                                                                       │
│                      │ │                                                                                                                           │
│                      │ │  ───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────  │
│                      │ │                                                                                                                           │
│                      │ │  Start something                                                                                                          │
│                      │ │                                                                                                                           │
│                      │ │   Build changed b     Lint all l     Run a task… r     Add a project a     Settings 7                                     │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
│                      │ │  Tip: every action here is a smith command. Press y on any row to copy it.                                                │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
╰──────────────────────╯ ╰───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯

 ↑/↓ move • enter see why • t test changed • 1-7 views • ctrl+k commands • ? more • q quit
 $ smith test --changed  ✗ 1 failed  ✓ 17 passed  38s                                                                           main  ·  2.64.1  ·  ⚠2
```

### Health

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Health ───────────────────────────────────╮╭─ Lock file out of date ─────────────────────────────────────────────────────╮
│                      │ │                                            ││                                                                             │
│   Home             1 │ │ ⚠ 2 warnings  · 0 failed · 46 passed       ││  ⚠ Warning                                                                  │
│                      │ │                                            ││                                                                             │
│   Projects         2 │ │ ▶ ⚠ Lock file out of date             fix  ││  package.json in web asks for react 19.2, but package-lock.json still holds 1
│                      │ │     web · package-lock.json                ││  A clean install in CI will fail.                                           │
│   Runs         ✗1  3 │ │                                            ││                                                                             │
│                      │ │   ⚠ TypeScript config drifts          fix  ││   1 Review  ── 2 Apply ── 3 Check again                                     │
│ ▶ Health       ⚠2  4 │ │     applies to 22 projects                 ││                                                                             │
│                      │ │                                            ││  ╭─ apps/web/package-lock.json  +2 −2 ───────────────────────────────────╮  │
│   Dependencies ↑3  5 │ │ ▸ 46 passed checks            p            ││  │                                                                       │  │
│                      │ │                                            ││  │     "node_modules/react": {                                           │  │
│   Release      ✓2  6 │ │                                            ││  │ -     "version": "19.1.0",                                            │  │
│                      │ │                                            ││  │ +     "version": "19.2.0",                                            │  │
│   Settings         7 │ │                                            ││  │ -     "resolved": "…/react-19.1.0.tgz",                               │  │
│                      │ │                                            ││  │ +     "resolved": "…/react-19.2.0.tgz",                               │  │
│                      │ │                                            ││  │       … 1 more line                                                   │  │
│                      │ │                                            ││  │                                                                       │  │
│ Pinned               │ │                                            ││  │                                                                       │  │
│                      │ │                                            ││  ╰───────────────────────────────────────────────────────────────────────╯  │
│ ✗ orders-api     app │ │                                            ││                                                                             │
│ ✓ web            app │ │                                            ││  The same as                                                                │
│ ● kit            lib │ │                                            ││  $ smith doctor --fix lockfile-sync -p web                                  │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │ ╭────────────────────────────────────────╮ ││                                                                             │
│                      │ │ │ Both warnings have a fix that          │ ││                                                                             │
│                      │ │ │ needs no decision.                     │ ││                                                                             │
│                      │ │ │                                        │ ││                                                                             │
│                      │ │ │  Review all fixes x                    │ ││  Nothing is committed. Undo with smith undo.                                │
│                      │ │ ╰────────────────────────────────────────╯ ││                                                                             │
│                      │ │                                            ││   Apply fix f     Not now esc                                               │
╰──────────────────────╯ ╰────────────────────────────────────────────╯╰─────────────────────────────────────────────────────────────────────────────╯

 ↑/↓ check • f apply fix • x review all • c check again • p show passed • ctrl+k commands
 $ smith doctor  ⚠ 2 warnings  46 passed                                                                                        main  ·  2.64.1  ·  ⚠2
```

### Deps

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Dependencies ─────────────────────────────────────────────────────────────────────╮╭─ react ─────────────────────────────╮
│                      │ │                                                                                    ││                                     │
│   Home             1 │ │  ↑ 3 with newer releases  · 1 patch · 1 minor · 1 major · checked 2 min ago        ││  19.1.0 → 19.2.0  ▲ minor           │
│                      │ │                                                                                    ││                                     │
│   Projects         2 │ │   All 3   Patch 1   Minor 1   Major 1                                              ││  Used by                            │
│                      │ │                                                                                    ││  web       apps/web/package.json    │
│   Runs         ✗1  3 │ │  Patch and minor releases: safe to take together                                   ││  kit       libs/kit/package.json    │
│                      │ │                                                                                    ││                                     │
│   Health       ⚠2  4 │ │     Package             Now       Latest    Level     Projects                     ││  One version                        │
│                      │ │  ────────────────────────────────────────────────────────────────────────────────  ││  ✓ both use 19.1.0                  │
│ ▶ Dependencies ↑3  5 │ │  ☐  react               19.1.0    19.2.0    ▲ minor   web, kit                     ││                                     │
│                      │ │  ☑  Newtonsoft.Json     13.0.3    13.0.4    • patch   orders-api, core             ││  Registry                           │
│   Release      ✓2  6 │ │                                                                                    ││  npmjs.com/package/react            │
│                      │ │  Major releases: may break your code, take one at a time                           ││  g opens it · Y copies the link     │
│   Settings         7 │ │                                                                                    ││                                     │
│                      │ │     Package             Now       Latest    Level     Projects                     ││  The same as                        │
│                      │ │  ────────────────────────────────────────────────────────────────────────────────  ││  $ smith up react                   │
│                      │ │  ☐  cobra               1.8.1     2.0.0     ⚠ major   ms-teams-bot                 ││                                     │
│ Pinned               │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│ ✗ orders-api     app │ │                                                                                    ││                                     │
│ ✓ web            app │ │                                                                                    ││                                     │
│ ● kit            lib │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │  ╭──────────────────────────────────────────────────────────────────────────────╮  ││                                     │
│                      │ │  │ 1 marked · Newtonsoft.Json 13.0.4 in 2 projects                              │  ││                                     │
│                      │ │  │                                                                              │  ││                                     │
│                      │ │  │  Update marked u     Update all safe U                                       │  ││                                     │
│                      │ │  ╰──────────────────────────────────────────────────────────────────────────────╯  ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
╰──────────────────────╯ ╰────────────────────────────────────────────────────────────────────────────────────╯╰─────────────────────────────────────╯

 ↑/↓ move • space mark • u update marked • 1-3 level • g release notes • o one version • ctrl+k commands
 $ smith up --dry-run  ↑ 3 newer                                                                                                main  ·  2.64.1  ·  ⚠2
```

### Release

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Release ──────────────────────────────────────────────────────────────────────────╮╭─ kit 1.3.2 → 1.4.0 ─────────────────╮
│                      │ │                                                                                    ││                                     │
│   Home             1 │ │  ✓ 2 packages to release  · 1 minor · 1 patch · from main                          ││  Changes since kit-v1.3.2           │
│                      │ │                                                                                    ││                                     │
│   Projects         2 │ │  Package           Now       Next      Bump       Commits   Publishes to           ││  feat  add a retry option to the clie
│                      │ │  ────────────────────────────────────────────────────────────────────────────────  ││  feat  export the time helpers      │
│   Runs         ✗1  3 │ │  kit               1.3.2     1.4.0     ▲ minor    4         npm                    ││  fix   keep the order of headers    │
│                      │ │  web               2.0.0     2.0.1     • patch    1         container registry     ││  docs  describe the retry option    │
│   Health       ⚠2  4 │ │                                                                                    ││                                     │
│                      │ │  unchanged: orders-api, core, words and 35 more                                    ││  feat → minor bump                  │
│   Dependencies ↑3  5 │ │                                                                                    ││                                     │
│                      │ │  ╭─ Checks before release ──────────────────────────────────────────────────────╮  ││  Tag                                │
│ ▶ Release      ✓2  6 │ │  │                                                                              │  ││  kit-v1.4.0                         │
│                      │ │  │ ✓ Working tree is clean                                                      │  ││                                     │
│   Settings         7 │ │  │ ✓ Tags give the right versions                                               │  ││  The same as                        │
│                      │ │  │ ✓ CI passed on main                                                          │  ││  $ smith release -p kit             │
│                      │ │  │ ○ Registry sign-in: checked when you release                                 │  ││                                     │
│                      │ │  │                                                                              │  ││                                     │
│ Pinned               │ │  ╰──────────────────────────────────────────────────────────────────────────────╯  ││                                     │
│                      │ │                                                                                    ││                                     │
│ ✗ orders-api     app │ │                                                                                    ││                                     │
│ ✓ web            app │ │                                                                                    ││                                     │
│ ● kit            lib │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │  Nothing is published until you confirm the plan.                                  ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │   Release 2 packages x     Dry run d                                               ││                                     │
│                      │ │                                                                                    ││                                     │
╰──────────────────────╯ ╰────────────────────────────────────────────────────────────────────────────────────╯╰─────────────────────────────────────╯

 ↑/↓ package • enter changes • x release • d dry run • ctrl+k commands
 $ smith release --dry-run  ✓ 2 to release                                                                                      main  ·  2.64.1  ·  ⚠2
```

### Settings

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Settings ─────────────────────╮╭─ Workspace · smith.toml ────────────────────────────────────────────────────────────────╮
│                      │ │                                ││                                                                                         │
│   Home             1 │ │ ▶ Workspace                    ││  Filter: release▏                                                                       │
│                      │ │                                ││                                                                                         │
│   Projects         2 │ │   Plugins                      ││  Setting                             Value                   Default         Type       │
│                      │ │                                ││  ─────────────────────────────────────────────────────────────────────────────────────  │
│   Runs         ✗1  3 │ │   Style                        ││  release.nuget_trusted_publishing    ● on                    not decided     switch     │
│                      │ │                                ││  release.channel                     stable ▾                stable          choice     │
│   Health       ⚠2  4 │ │   Recipes                      ││  release.tag_prefix                  {project}-v             {project}-v     text       │
│                      │ │                                ││  release.rubygems_trusted_publishing ○ off                   not decided     switch     │
│   Dependencies ↑3  5 │ │   Forms                        ││                                                                                         │
│                      │ │                                ││  ╭─ release.nuget_trusted_publishing ────────────────────────────────────────────────╮  │
│   Release      ✓2  6 │ │                                ││  │                                                                                   │  │
│                      │ │                                ││  │ Publish NuGet packages with a short-lived key from GitHub OIDC,                   │  │
│ ▶ Settings         7 │ │ Per-project settings are in    ││  │ so the repository keeps no NuGet API key.                                         │  │
│                      │ │ Projects › a project › Settings││  │                                                                                   │  │
│                      │ │                                ││  │ ( ) Not decided    (•) On    ( ) Off                                              │  │
│                      │ │                                ││  │                                                                                   │  │
│ Pinned               │ │                                ││  │ $ smith config set release.nuget_trusted_publishing true                          │  │
│                      │ │                                ││  ╰───────────────────────────────────────────────────────────────────────────────────╯  │
│ ✗ orders-api     app │ │                                ││                                                                                         │
│ ✓ web            app │ │                                ││                                                                                         │
│ ● kit            lib │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││  Writes smith.toml exactly as the command does.                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││   Save enter     Reset to default u                                                     │
╰──────────────────────╯ ╰────────────────────────────────╯╰─────────────────────────────────────────────────────────────────────────────────────────╯

 ↑/↓ move • tab next pane • enter change • u reset • ctrl+k commands
 $ smith config list                                                                                                            main  ·  2.64.1  ·  ⚠2
```

### Help

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Home ────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                      │ │                                                                                                                           │
│ ▶ Home             1 │ │  monosmith  42 projects · 7 changed since main · last run 3 min ago                                                       │
│                      │ │ ╭─ Keys · Home ────────────────────────────────────────────────────────────────────────────────╮                          │
│   Projects         2 │ │ │                                                                                              │───────────────────────╮  │
│                      │ │ │  Everywhere                     Lists                          Runs                          │ease                   │  │
│   Runs         ✗1  3 │ │ │                                                                                              │ ready                 │  │
│                      │ │ │  1-7      open a view           ↑/↓ j/k  move                  ctrl+c   stop                 │inor · 1 patch         │  │
│   Health       ⚠2  4 │ │ │  ctrl+k   commands              enter    open or run           r        run again            │───────────────────────╯  │
│                      │ │ │  tab      next pane             /        filter                R        rerun failed         │                          │
│   Dependencies ↑3  5 │ │ │  ?        these keys            space    select several        e        errors only          │                          │
│                      │ │ │  m        mouse on / off        y        copy the command      Y        copy the log         │                          │
│   Release      ✓2  6 │ │ │  q        quit                  esc      back                  , .      older / newer run    │         See why enter    │
│                      │ │ │                                                                                              │                          │
│   Settings         7 │ │ │                                                                                              │                          │
│                      │ │ │──────────────────────────────────────────────────────────────────────────────────────────────│        Test changed t    │
│                      │ │ │                                                                                              │                          │
│                      │ │ │  Every action is a smith command. The status bar shows the last one, and y copies it.        │                          │
│ Pinned               │ │ │                                                                                              │              Review 4    │
│                      │ │ │  Mouse: click a row, a tab or a key hint. Drag a divider to resize. Wheel scrolls.           │                          │
│ ✗ orders-api     app │ │ │  Plain mode (NO_COLOR or SMITH_THEME=plain) keeps every symbol, so no state is told by colour alone.                    │
│ ✓ web            app │ │ │                                                                                              │              Review 5    │
│ ● kit            lib │ │ │                                                                                              │                          │
│                      │ │ │                                                                                              │                          │
│                      │ │ │                                                                                              │             Preview 6    │
│                      │ │ │                                                                                              │                          │
│                      │ │ │                                                                                              │                          │
│                      │ │ │                                                                                              │────────────────────────  │
│                      │ │ │                                                                                              │                          │
│                      │ │ │                                                                                              │                          │
│                      │ │ │  Find a command by name:  ctrl+k                                                   esc close │                          │
│                      │ │ │                                                                                              │                          │
│                      │ │ ╰──────────────────────────────────────────────────────────────────────────────────────────────╯                          │
│                      │ │                                                                                                                           │
│                      │ │  Tip: every action here is a smith command. Press y on any row to copy it.                                                │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
╰──────────────────────╯ ╰───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯

 ↑/↓ move • enter see why • t test changed • 1-7 views • ctrl+k commands • ? more • q quit
 $ smith test --changed  ✗ 1 failed  ✓ 17 passed  38s                                                                           main  ·  2.64.1  ·  ⚠2
```

### NarrowHome

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Home ───────────────────────────────────────────────────────────────────────╮
│ 42 projects · 7 changed · last run 3 min ago                                 │
│                                                                              │
│ ╭───────────────────────────────────╮ ╭───────────────────────────────────╮  │
│ │ Changed                           │ │ Last run                          │  │
│ │ ● 7 projects                      │ │ ✗ 1 failed                        │  │
│ ╰───────────────────────────────────╯ ╰───────────────────────────────────╯  │
│ ╭───────────────────────────────────╮ ╭───────────────────────────────────╮  │
│ │ Health                            │ │ Release                           │  │
│ │ ⚠ 2 warnings                      │ │ ✓ 2 ready                         │  │
│ ╰───────────────────────────────────╯ ╰───────────────────────────────────╯  │
│                                                                              │
│ Needs you now 1/5                                                            │
│                                                                              │
│ │ ✗ test failed in orders-api                                         enter  │
│ │   2 of 140 tests · 4 min ago                                               │
│   ● 7 projects changed, not tested                                        t  │
│     web, kit, orders-api +4                                                  │
│   ⚠ 2 health checks warn                                                  4  │
│     both have a fix                                                          │
│   ↑ 3 dependency updates                                                  5  │
│     1 major                                                                  │
│   ✓ 2 packages ready to release                                           6  │
│     kit 1.4.0 · web 2.0.1                                                    │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ move • enter see why • t test changed • ? more
 $ smith test --changed  ✗ 1  ✓ 17  38s                                 2.64.1
```

### NarrowProjects

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Projects 7/42 ──────────────────────────────────────────────────────────────╮
│ Filter: changed▏                                        ✓ pass ✗ fail – none │
│                                                                              │
│ APPS                                                                         │
│ ▶ ✗  orders-api           app   C#        ● changed               test ✗     │
│   ✓  web                  app   React     ● changed                          │
│   ✓  ms-teams-bot         app   Go        ● changed                          │
│   –  billing-fn           app   Python    ● changed                          │
│ LIBS                                                                         │
│   ✓  kit                  lib   TS        ● changed                          │
│   ✓  core                 lib   C#        ● changed                          │
│   ✓  words                lib   Go        ● changed                          │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│ enter opens the project · space selects several to run together              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 / filter • enter open • space select • t test
 $ smith run test -p orders-api                                         2.64.1
```

### NarrowProject

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ ‹ Projects / orders-api ────────────────────────────────────────────────────╮
│ app · C# · apps/orders-api  ● changed                                Run… r  │
│                                                                              │
│  Tasks   Deps   Runs   Settings                                              │
│ ──────────────────────────────────────────────────────────────────────────── │
│ Task        Result          Time     Cache               Key                 │
│ build       ✓ passed        4.1s     hit                 b                   │
│ test        ✗ 2 failed      21.6s    miss · changed      t                   │
│ lint        ✓ passed        2.3s     hit                 l                   │
│ package     – not run       –        –                   p                   │
│ start       own terminal    –        never               s                   │
│                                                                              │
│ Uses      core · kit · words                                                 │
│ Used by   web · billing-fn  (a change here tests them too)                   │
│                                                                              │
│ Recent runs                                                                  │
│                                                                              │
│ ✗ test    2 failed   21.6s   4 min ago                                       │
│ ✓ build   passed     4.1s    4 min ago                                       │
│ ✓ test    passed     19.8s   yesterday                                       │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 esc back • tab section • enter run • r run…
 $ smith run test -p orders-api                                         2.64.1
```

### NarrowRunTasks

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ smith test --changed ───────────────────────────────────────────────────────╮
│ ⣾ 18 tasks · 00:24  about 20s left                              Stop ctrl+c  │
│                                                                              │
│ ██████████████████████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░  67%  │
│ ✗ 1   ✓ 9   ⚡ 2   ◐ 2   ○ 4                                                  │
│                                                                              │
│  Tasks   Log: orders-api:test                                   tab switches │
│──────────────────────────────────────────────────────────────────────────────│
│ Failed                                                                       │
│ ├─ ✗ orders-api:test                                                  21.6s  │
│ Running                                                                      │
│ ├─ ⣾ web:test                                                           14s  │
│ ├─ ⣾ billing-fn:test                                                     6s  │
│ Queued · waits on web:test                                                   │
│ ├─ ○ web-e2e:test                                                            │
│ ├─ ○ docs:test                                                               │
│ Done                                                                         │
│ ├─ ✓ kit:test                                                          3.2s  │
│ ├─ ⚡ core:test                                                        cache  │
│ ├─ ✓ … 8 more                                                                │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ task • enter log • tab log • R rerun failed • esc back
 ◐ running  00:24                                                       2.64.1
```

### NarrowRunLog

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ smith test --changed ───────────────────────────────────────────────────────╮
│ ⣾ 18 tasks · 00:24  about 20s left                              Stop ctrl+c  │
│                                                                              │
│ ██████████████████████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░  67%  │
│ ✗ 1   ✓ 9   ⚡ 2   ◐ 2   ○ 4                                                  │
│                                                                              │
│  Tasks   Log: orders-api:test                                   tab switches │
│──────────────────────────────────────────────────────────────────────────────│
│  Errors only   Full log                                 ✗ 2 · lines 1-11/11  │
│                                                                              │
│ ✗ OrderTotals_IncludeTax                                                    ┃│
│   tests/OrderTotalsTests.cs:42                                              ┃│
│     Expected: 119.00                                                        ┃│
│     Actual:   100.00                                                        ┃│
│                                                                             ┃│
│ ✗ Refund_AfterShipping_IsPartial                                            ┃│
│   tests/RefundTests.cs:88                                                   ││
│     Expected status Partial, got Full                                       ││
│                                                                             ││
│ ✓ 138 passed   ✗ 2 failed   21.6s                                           ││
│                                                                             ││
│                                                                             ││
│                                                                             ││
│ Cache miss: 2 files changed in src/Orders.       Rerun failed R              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ scroll • e errors only • / search • Y copy log • tab tasks
 ◐ running  00:24                                                       2.64.1
```

### NarrowPalette

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Home ───────────────────────────────────────────────────────────────────────╮
│ 42 projects · 7 changed · last run 3 min ago                                 │
│   ╭─ Commands ───────────────────────────────────────────────────────────╮   │
│ ╭─│                                                                      │╮  │
│ │ │ › test orders▏                                                  esc  ││  │
│ │ │──────────────────────────────────────────────────────────────────────││  │
│ ╰─│ ▶ Test orders-api                                             enter  │╯  │
│ ╭─│   smith run test -p orders-api                                       │╮  │
│ │ │ ▶ Test orders-api and dependents                                     ││  │
│ │ │   smith run test -p orders-api --dependents                          ││  │
│ ╰─│ ↻ Rerun the 2 failed tests                                           │╯  │
│   │   from the last run · 4 min ago                                      │   │
│ Ne│ ✗ Open orders-api                                                    │   │
│   │   project · app · C#                                                 │   │
│ │ │ ⚙ orders-api: test timeout                                           │r  │
│ │ │   smith config set -p orders-api …                                   │   │
│   │                                                                      │t  │
│   │                                                                      │   │
│   │──────────────────────────────────────────────────────────────────────│4  │
│   │ ↑/↓ move • enter run • tab edit first                                │   │
│   ╰──────────────────────────────────────────────────────────────────────╯5  │
│     1 major                                                                  │
│   ✓ 2 packages ready to release                                           6  │
│     kit 1.4.0 · web 2.0.1                                                    │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ move • enter see why • t test changed • ? more
 $ smith test --changed  ✗ 1  ✓ 17  38s                                 2.64.1
```

### NarrowHealth

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Health ─────────────────────────────────────────────────────────────────────╮
│ ⚠ 2 warnings  · 0 failed · 46 passed · 1 min ago                             │
│                                                                              │
│ ▶ ⚠ Lock file out of date                                             fix    │
│     web · package-lock.json                                                  │
│                                                                              │
│   ⚠ TypeScript config drifts from the standard                        fix    │
│     applies to 22 projects                                                   │
│                                                                              │
│ ▸ 46 passed checks                                                           │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│ ╭──────────────────────────────────────────────────────────────────────────╮ │
│ │ Both warnings have a fix that needs no decision.                         │ │
│ │                                                                          │ │
│ │  Review all fixes x                                                      │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ check • enter open • x review all • p passed
 $ smith doctor  ⚠ 2  46 passed                                         2.64.1
```

### NarrowHealthFix

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ ‹ Health / Lock file out of date ───────────────────────────────────────────╮
│ ⚠ Warning · web                                                              │
│                                                                              │
│ package.json asks for react 19.2, but package-lock.json                      │
│ still holds 19.1. A clean install in CI will fail.                           │
│                                                                              │
│  1 Review  ── 2 Apply ── 3 Check again                                       │
│                                                                              │
│ ╭─ package-lock.json  +2 −2 ───────────────────────────────────────────────╮ │
│ │                                                                          │ │
│ │ -   "version": "19.1.0",                                                 │ │
│ │ +   "version": "19.2.0",                                                 │ │
│ │ -   "resolved": "…/react-19.1.0.tgz",                                    │ │
│ │ +   "resolved": "…/react-19.2.0.tgz",                                    │ │
│ │     … 1 more line                                                        │ │
│ │                                                                          │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
│                                                                              │
│ Nothing is committed. Undo with smith undo.                                  │
│                                                                              │
│  Apply fix f     Not now esc                                                 │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 esc back • f apply • ↑/↓ scroll • c check again
 $ smith doctor --fix lockfile-sync -p web                              2.64.1
```

### NarrowDeps

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Dependencies ───────────────────────────────────────────────────────────────╮
│ ↑ 3 newer · 1 patch · 1 minor · 1 major                                      │
│                                                                              │
│  All 3   Patch 1   Minor 1   Major 1                                         │
│                                                                              │
│ Safe together                                                                │
│    Package            Now      Latest   Level     Used by                    │
│ ──────────────────────────────────────────────────────────────────────────── │
│ ☐  react              19.1.0   19.2.0   ▲ minor   web +1                     │
│ ☑  Newtonsoft.Json    13.0.3   13.0.4   • patch   orders-api +1              │
│                                                                              │
│ Take one at a time                                                           │
│    Package            Now      Latest   Level     Used by                    │
│ ──────────────────────────────────────────────────────────────────────────── │
│ ☐  cobra              1.8.1    2.0.0    ⚠ major   ms-teams-bot               │
│                                                                              │
│ 1 marked: Newtonsoft.Json 13.0.4                                             │
│                                                                              │
│  Update marked u     Update all safe U                                       │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ move • space mark • u update • enter details
 $ smith up --dry-run                                                   2.64.1
```

### NarrowRelease

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Release ────────────────────────────────────────────────────────────────────╮
│ ✓ 2 packages to release · 1 minor · 1 patch                                  │
│                                                                              │
│ Package       Now      Next     Bump      To                                 │
│ ──────────────────────────────────────────────────────────────────────────── │
│ kit           1.3.2    1.4.0    ▲ minor   npm                                │
│ web           2.0.0    2.0.1    • patch   containers                         │
│                                                                              │
│ unchanged: orders-api, core, words +35                                       │
│                                                                              │
│ Checks                                                                       │
│                                                                              │
│ ✓ Working tree is clean                                                      │
│ ✓ Tags give the right versions                                               │
│ ✓ CI passed on main                                                          │
│                                                                              │
│                                                                              │
│ Nothing is published until you confirm the plan.                             │
│                                                                              │
│  Release 2 x     Dry run d                                                   │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ package • enter changes • x release • d dry run
 $ smith release --dry-run                                              2.64.1
```

### NarrowSettings

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Settings ───────────────────────────────────────────────────────────────────╮
│                                                                              │
│ ▶ Workspace                                                                  │
│   every smith.toml setting, 42                                               │
│                                                                              │
│   Plugins                                                                    │
│   6 on · 2 external, trusted                                                 │
│                                                                              │
│   Style                                                                      │
│   theme auto · symbols unicode                                               │
│                                                                              │
│   Recipes                                                                    │
│   1 recipe, valid                                                            │
│                                                                              │
│   Forms                                                                      │
│   add or remove a project…                                                   │
│                                                                              │
│                                                                              │
│ Per-project settings: Projects › a project › Settings                        │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ move • enter open • esc back
 $ smith config list                                                    2.64.1
```

### NarrowSetting

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ ‹ Settings / Workspace ─────────────────────────────────────────────────────╮
│ Filter: release▏                                                             │
│                                                                              │
│ release.nuget_trusted_publishing                              ● on           │
│ release.channel                                               stable ▾       │
│ release.tag_prefix                                            {project}-v    │
│ release.rubygems_trusted_publishing                           ○ off          │
│                                                                              │
│ ╭─ release.nuget_trusted_publishing ───────────────────────────────────────╮ │
│ │                                                                          │ │
│ │ Publish NuGet packages with a short-lived key from                       │ │
│ │ GitHub OIDC, so the repository keeps no NuGet API key.                   │ │
│ │                                                                          │ │
│ │ ( ) Not decided   (•) On   ( ) Off                                       │ │
│ │                                                                          │ │
│ │ Default: not decided · switch                                            │ │
│ │                                                                          │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
│                                                                              │
│  Save enter     Reset u                                                      │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ setting • enter change • u reset • esc back
 $ smith config set release.nuget_trusted_publishing true               2.64.1
```

