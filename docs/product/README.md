# Why MonoSmith

These pages are for people who have not used MonoSmith (`smith`) and want to
decide whether it fits their monorepo: a team lead, an engineering manager, a
developer picking a tool. They say what smith does, what it does not do, and how
it compares with the other tools you may be looking at.

They make two promises.

- **Every claim about smith was checked** against the repository, the
  reference pages in [`docs/reference`](../reference/cli.md), the output of
  `smith --help`, or a run of smith in a scratch directory. Where a feature is
  partial, the page says so. Where something could not be checked, it says
  "not verified".
- **Every claim about another tool comes from that tool's own documentation
  or repository.** Each comparison page ends with its sources and the date they
  were read. A cell in the comparison tables is marked "absent" (❌) for a
  competitor only where a documentation page says so, or describes a
  different mechanism for the same job and a search found nothing else, and "not verified" (❔)
  where the sources were searched and say nothing either way.

smith is young. Its repository's first commit is dated 2026-10-15 and this
folder was written on 2026-10-25, at release 1.62.0. The
[evaluator FAQ](faq-for-evaluators.md) answers the maturity question directly.

## Read in this order

| Page | Read it to... | Time |
|---|---|---|
| [The case for MonoSmith](why-monosmith.md) | see the problems of a multi-language monorepo, what smith does about each, and who it is for and not for | 8 min |
| [MonoSmith in five minutes](how-it-works-in-five-minutes.md) | see the product in commands you can run | 5 min |
| [Migration stories](migration-stories.md) | know what converting from Nx, Turborepo or mnci really does | 8 min |
| [Monorepo tools compared](comparison/README.md) | compare smith with 14 other tools in tables, and read a page on each | 10 min |
| [Choosing a monorepo tool](comparison/choosing.md) | get a decision guide, including the cases where another tool wins | 5 min |
| [FAQ for evaluators](faq-for-evaluators.md) | get answers on cost, lock-in, licence, security, support and the roadmap | 5 min |

## The short version

smith is one program, `smith`, that runs a monorepo in Node, Go, Python, Rust,
.NET, Ruby, PHP, C and C++, Dart, the JVM and Bicep. It works out the projects
and their dependencies from the files you already have, runs their tasks in
order from a content-addressed cache, writes the CI pipeline, and releases from
conventional commits. It is one static binary, licensed MIT.

It is not a replacement for everything. It has no remote execution and no
hermetic sandbox, and its Node plugin runs npm only. If you need those, the
[decision guide](comparison/choosing.md) says which tool to look at.

## For engineers who want the technical documentation

The technical documentation is in [docs/](../README.md): guides, reference,
design and contributing.
