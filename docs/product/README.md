# Why MonoSmith

**Run a monorepo in eleven languages with one tool.** `smith add` scaffolds a project that already
lints, tests and builds, in the language you pick. One dependency graph, read from the files you
already have, orders the work and says what a change reaches. Plugins in any language extend it.
CI, a dev container and releases are generated from the same model.

These pages are for people who have not used MonoSmith (`smith`) and want to
decide whether it fits their monorepo: a team lead, an engineering manager, a
developer picking a tool. They say what smith does, what it does not do, and how
it compares with the other tools you may be looking at.

## Three reasons to choose it, each with its proof

| You want | MonoSmith gives you | The proof, tool by tool |
|---|---|---|
| **A new Rust, Python, .NET or Go project that works on day one** | a scaffold, the workspace around it, tasks, CI, a dev container and a release target for eleven ecosystems, with no BUILD files or per-language rules | [Languages and ready-made environments](comparison/languages-and-environments.md) |
| **One dependency graph for every language** | edges read from manifests, imports and shared schemas; `smith graph` as a tree, JSON, DOT, HTML or a terminal browser; blast radius and diff between revisions | [The dependency graph](comparison/dependency-graph.md) |
| **To extend the tool in a language your team already uses** | plugins in five roles over JSON-RPC, SDKs for Go, TypeScript, Python and C#, `smith plugin check`, stacks and recipes for team standards, a trust gate for plugins a repository names | [Plugins and flexibility](comparison/plugins-and-flexibility.md) |

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
folder was written on 2026-10-25, at release 1.62.0; the three deep-dive pages were checked
against release 1.65.0 the same day. The
[evaluator FAQ](faq-for-evaluators.md) answers the maturity question directly.

## Read in this order

| Page | Read it to... | Time |
|---|---|---|
| [The case for MonoSmith](why-monosmith.md) | see the problems of a multi-language monorepo, what smith does about each, and who it is for and not for | 8 min |
| [MonoSmith in five minutes](how-it-works-in-five-minutes.md) | see the product in commands you can run | 5 min |
| [Migration stories](migration-stories.md) | know what converting from Nx, Turborepo or mnci really does | 8 min |
| [Monorepo tools compared](comparison/README.md) | compare smith with 14 other tools in tables, and read a page on each | 10 min |
| [Languages and ready-made environments](comparison/languages-and-environments.md) | see what you get on day one in each of eleven languages, and what the other tools make you write | 8 min |
| [The dependency graph](comparison/dependency-graph.md) | see how the graph is built, what you can ask it, and where Bazel, Pants and Nx are ahead | 8 min |
| [Plugins and flexibility](comparison/plugins-and-flexibility.md) | see how each tool is extended, in what language, and how safely | 8 min |
| [Everything compared](comparison/everything-compared.md) | see one table of every smith capability against 14 tools, what only smith documents and where it is behind | 10 min |
| [Choosing a monorepo tool](comparison/choosing.md) | get a decision guide, including the cases where another tool wins | 5 min |
| [FAQ for evaluators](faq-for-evaluators.md) | get answers on cost, lock-in, licence, security, support and the roadmap | 5 min |

## The short version

smith is one program, `smith`, that runs a monorepo in Node, Go, Python, Rust,
.NET, Ruby, PHP, C and C++, Dart, the JVM and Bicep. It works out the projects
and their dependencies from the files you already have, runs their tasks in
order from a content-addressed cache, writes the CI pipeline, and releases from
conventional commits. It is one static binary, licensed MIT.

It is not a replacement for everything. It has no remote execution and no
hermetic sandbox, its graph is per project and not per file, its Node plugin runs
npm only, and it has no registry of third-party plugins yet (only a marketplace index). If you need those, the
[decision guide](comparison/choosing.md) says which tool to look at.

## For engineers who want the technical documentation

The technical documentation is in [docs/](../README.md): guides, reference,
design and contributing.
