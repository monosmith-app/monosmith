# The bring-in lab

The bring-in lab holds the promise of `smith adopt` and `smith import`: a real repository
comes into a workspace and works. It adopts one public sample repository into a new
workspace, imports the other samples into that workspace, and checks each one. The lab is in
`tools/bring-in-lab/` and the workflow is `.github/workflows/bring-in-lab.yml` (issue #1155).

## What it checks

The first sample is adopted with `smith adopt --init --yes --commit` and the flags its line
gives (`--move`, `--tooling smith`). Each other sample is imported with
`smith import <url> --ref <sha> --history keep --yes --commit` and the flags of its line.
After each bring-in the lab runs these steps in the workspace:

| Step | Command |
|---|---|
| `install` | `smith install` |
| `build` | `smith run build -p <project>` |
| `test` | `smith run test -p <project>` |
| `lint` | `smith run lint -p <project>` |
| `doctor` | `smith doctor` |
| `ignored` | `git ls-files` and `git status`: in the project's folder, no `node_modules`, `dist`, `.venv`, `obj` or `bin/Debug` file is committed or shows as untracked |

A step gets one of these results:

| Result | Meaning |
|---|---|
| `pass` | the command exits 0 |
| `left to do` | the command fails, and the "Left to do" list of that bring-in names the step (an item that says build, test, lint or install; for `doctor`, any item) |
| `known` | the command fails, and `known-failing.tsv` lists the step with an open issue |
| `FAIL` | any other failure, or a crash (`panic:`, "bug in smith") in any step |
| `skipped` | the bring-in failed, so the steps did not run |

The lab exits 1 when one step is `FAIL`. So it is green on what works and red on a
regression. Lint findings in a sample's own code are allowed only when the "Left to do" list
names lint.

After the `ignored` step the lab adds the build folders to `.git/info/exclude` and commits
what the steps changed (a lockfile). The next import with `--history keep` needs a clean
tree. The workspace's own files stay as smith wrote them.

## Run it

The lab needs git, Node.js 24, Python 3.12 with uv, the .NET 10 SDK and network access.
Build the smith you want to test, then run the script with bash (Git Bash on Windows):

```sh
go build -o /tmp/smith ./apps/smith
SMITH=/tmp/smith bash tools/bring-in-lab/run.sh ~/lab/bring-in
```

- The argument is the lab folder. The lab keeps it, with the workspace in `ws/` and one log
  per step in `logs/` (`<scenario>-<step>.log`). Without the argument the lab uses a
  temporary folder and removes it at the end (`LAB_KEEP=1` keeps it).
- `LAB_ONLY=express-api,astar` runs the adopt scenario and only these imports.
- The lab prints a table of the results. In GitHub Actions the same table goes to the job
  summary, and the logs are an artifact of the run.

In GitHub, the workflow runs every Tuesday at 05:00 UTC and on demand (Actions,
bring-in-lab, Run workflow). It builds smith from the commit. It does not run on pull
requests, because it takes too long. Run it on demand after a change to adopt, import,
the tooling step or a language plugin's identification.

## Add a repository

1. Choose a small public repository with a licence that allows a copy (MIT, Apache 2.0). It
   must show a case the lab does not have: a language, a framework, a layout.
2. Pin it to a commit: `gh api repos/<owner>/<repo>/commits/HEAD --jq .sha`.
3. Add one line to `tools/bring-in-lab/repos.tsv`, with tabs between the fields: the
   scenario id, `import`, the URL, the SHA, the project name and the flags. Give `--name`,
   so the project name is known, and `--as`, because the lab has no terminal. Add `--path`
   for a project in a sub-folder and `--manifest <language>` for sources without a
   manifest. Only the first line is `adopt`.
4. Run the lab with `LAB_ONLY=<id>`.
5. For each `FAIL`, find the cause. When the sample itself is broken (its tests fail before
   smith touches it), choose another commit or another repository. When smith is the cause,
   open an issue (label `bug` or `ux`, refer to #1156) and add a line to
   `tools/bring-in-lab/known-failing.tsv`: the scenario id, the step and the issue number.

When an issue is fixed, the lab prints "passes now" for its step. Remove the line from
`known-failing.tsv` in the same pull request as the fix.

## The samples

| Id | Repository | Case |
|---|---|---|
| `mitt` | developit/mitt | a TypeScript package at the repository root, adopted with smith's tooling |
| `express-api` | rzgry/Express-REST-API-Template | an Express API in JavaScript, with Jest |
| `vite-react` | joaopaulomoraes/reactjs-vite-tailwindcss-boilerplate | a Vite React app with Vitest and a pnpm lockfile |
| `sampleproject` | pypa/sampleproject | a Python package with `pyproject.toml` and pytest |
| `astar` | roy-t/AStar | a .NET library (`netstandard2.0`) in a sub-folder, beside its tests, that packs the `LICENSE` above its folder (`--outside-files copy`) |
| `file-analyzer` | mraza007/python-file-analyzer | one Python file and no manifest |

The commits are in `tools/bring-in-lab/repos.tsv`.
