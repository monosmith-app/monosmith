# Contributing

For people who change MonoSmith itself.

| Page | Holds |
|---|---|
| [Working on smith](development.md) | the repository layout, the slices of `apps/smith`, how to build, test and lint with smith itself, how releases of this repository work |
| [Documentation audit, October 2026](documentation-audit.md) | what the October 2026 fact-check of these docs found, page by page |
| [Roadmap proposal from the comparison gaps](roadmap-proposal.md) | a proposal, not a plan: the gaps found by the tool comparison, ranked, with the decisions the owner must make |

Read the [architecture manual](../architecture/README.md) first for how smith is put together and
the [decision records](../architecture/decisions/README.md) for why. The code follows the vertical-slice
layout described in [the repository's own slice architecture](../architecture/slice-architecture.md).

Commits follow [Conventional Commits](https://www.conventionalcommits.org): a `feat:` or `fix:`
commit on `main` releases.
