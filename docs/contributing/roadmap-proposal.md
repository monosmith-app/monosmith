# Roadmap proposal from the comparison gaps

> **This is a proposal, not a plan.** It is written for the maintainers of MonoSmith, from
> the gaps found by [Everything compared](../product/comparison/everything-compared.md) on
> 2026-10-25 (smith 1.63.0). Nothing here is committed, dated or promised, and the product
> pages do not mention it. The owner decides what, if anything, becomes work.

> **Update, 2026-11-01 (smith 2.0 line).** Some gaps below were closed after this page was written,
> as far as the code shows: the import-cycle failure (#558, fixed after 1.65.0; see
> [project discovery](../architecture/project-discovery.md#step-7-the-checks-that-end-discovery)),
> the plugin trust prompt and minimal environment ([ADR 0036](../architecture/decisions/0036-external-plugins-are-trusted-and-get-a-minimal-environment.md)),
> registry checks before a first publish ([ADR 0043](../architecture/decisions/0043-a-release-is-rehearsed-against-its-registries-with-read-only-requests.md)),
> the TypeScript, Python and C# SDKs, and CircleCI and Bitbucket Pipelines generation. SBOM coverage grew to ten
> language plugins. The ranking below is unchanged and was not re-done.

## How the gaps were ranked

Every row of the comparison where another tool is ✅ and smith is ❌ or ◐ is a gap, and
every row where smith is not ✅ and no rival is either is a shortfall in smith itself. Each
is ranked on three things:

- **Customer value**: how often a team evaluating or running a polyglot monorepo hits it
  (H, M, L). This is judgement from the comparison and the open issues, not from usage data
  smith does not have.
- **Effort**: S (days), M (a few weeks), L (months or a design decision first).
- **Differentiation**: whether the work *closes* a gap other tools already closed, *extends*
  something smith already leads on, or *removes a risk* that no tool in the set has addressed
  either.

Deliberate non-goals are not gaps and are not ranked: remote and distributed execution,
sandboxed execution by default, a file-level graph, watch mode and downloading toolchains
([Execution and cache, section 9](../architecture/execution-and-cache.md),
[ADR 0020](../architecture/decisions/0020-toolchain-pins-checked-not-downloaded.md)).
Decision 1 below asks whether to keep it that way.

## The gaps, ranked

| # | Gap | Value | Effort | Differentiation | Issue | Horizon |
|---|---|---|---|---|---|---|
| 1 | A project import cycle makes every command fail, including the doctor that reports cycles | H | S | removes a defect | #558 | Now |
| 2 | Plugin security model: a trust prompt, a minimal environment, digest pins | H | M | removes a risk no tool in the set documents solving | #565 (step 1) | Now |
| 3 | Registry checks before the first publish; dry-run uploads; a journal that stops a bad key from half-releasing | H | M | extends a lead | #347, #380 | Now |
| 4 | SBOM and provenance for every ecosystem and for Azure and GitLab | H | M | closes a gap (npm and pnpm have `sbom`) and extends a lead | #348 | Now |
| 5 | Cache: hit-rate report, file outputs, compression, a reference cache server, an opt-in strict-inputs check | M to H | S to M | closes small gaps; the strict check narrows the sandbox gap without a sandbox | #563 | Now |
| 6 | Rules for which projects may depend on which | M to H | M | closes a gap with six tools | #570 | Next |
| 7 | Plugin SDKs beyond Go, TypeScript first | M | M to L | enables the ecosystem; closes a gap | #564 | Next (after #565 step 1) |
| 8 | Dependency tooling beyond Node and Python: `sync`, `up`, `deps-check`, audit | M | M | closes a gap | #572 | Next |
| 9 | CircleCI and Bitbucket Pipelines generation | M | S to M | closes a gap with Nx | #571 | Next |
| 10 | Per-platform release archives for apps that are not Go | M | M | extends a lead | #573 | Next |
| 11 | Trusted publishing (OIDC) for npm, NuGet and crates.io | M to H | S per registry | closes a gap with npm | needs an issue | Next |
| 12 | JetBrains support: a language server first, then the plugin | M to H | L | closes a gap with five tools | #303 | Next, started with the language server |
| 13 | Adopt at the full level for plain repositories | M | M | extends a lead | #553, #548 | Next |
| 14 | Remote Execution API for caching: first check which servers work over HTTP | M | S for the note, L for gRPC | closes a gap | #345 | Next as research, Later as code |
| 15 | Task-level affected selection | M | L | closes a gap with Nx and Bazel | #574 | Later, after Decision 2 |
| 16 | Plugin packager and an install-from-GitHub path; a registry only when third-party plugins exist | M | M to L | closes a gap with Nx | #565 (steps 2 and 3), #164 | Later |
| 17 | OS package managers and signing and notarisation | L to M | M | closes a gap | #225, #124 | Later |
| 18 | Importers for Lerna, Rush, Bazel and Gradle workspaces | L | L | no rival documents one either | #553 is the nearest | Later, if asked |

## Themes, in the order suggested

### Now: trust and release safety

The work that protects a user who already runs smith, and the first thing a security review
asks. Items 1 to 5. The theme is "nothing surprising happens when you clone, run or
release". Item 2 comes before the SDK work, because more third-party plugins (item 7) widen
the surface #565 describes: a cloned repository can make smith start a program of its
choosing, and every plugin sees the whole environment, secrets included.

Housekeeping that costs little: #343 (schema edges and the blast-radius query) is still open
although the feature shipped ([ADR 0028](../architecture/decisions/0028-graph-edges-from-shared-schemas.md),
`smith graph --blast-radius`), and the checklist in the #563 epic still lists pruning (#559,
closed) and a `smith cache` command (shipped in 1.63.0) as open. Closing and ticking them
makes the open list match the product.

### Next: breadth that evaluators ask about

Items 6 to 14. Each closes a row of the comparison where one or more tools are ahead. Item 6
(boundary rules) has the clearest customer story: architecture governance as a repository
grows, which six of the fourteen tools already document. Items 8 and 9 widen what "one tool for
the whole chain" covers. Item 12 should start with the language server in #303, since the
same server serves VS Code, JetBrains and other editors.

### Later: scale and ecosystem

Items 15 to 18. These are large, or wait on a decision, or wait for demand. A real plugin
registry needs namespace ownership, moderation and signing, and the project has one main
maintainer (#565 proposes a curated index and install-from-GitHub first).

## Decisions the owner must make

1. **Keep the non-goals?** Remote execution, a default sandbox, a file-level graph and watch
   mode each cost smith the biggest-repository evaluations (Bazel, Pants, Buck2 and Nx win
   them). Keeping them is what keeps smith a single binary that runs `npm`, `dotnet` and
   `cargo` unmodified. The cheap middle path is the opt-in strict-inputs check in #563.
2. **Project-level or task-level affected?** Task-level selection (Nx, moon, Bazel) runs less
   in a big project and costs a more detailed model of each task's inputs. Item 15 waits on
   this answer.
3. **A hosted service, or only a reference cache server?** Four tools sell a hosted cache
   and analytics. Today smith has none. #563 proposes `smith cache serve`; a hosted service is
   a business decision, not an engineering one.
4. **Which SDK languages, and in what order?** #564 proposes TypeScript, then Python, then
   C#. The order decides which plugin authors can start without writing the protocol by hand.
5. **Registry now, or "registry-lite"?** #565 recommends install-from-GitHub with a
   PR-reviewed index, and a real registry only with third-party demand.
6. **JetBrains: language server first, or a native plugin?** The language server is shared
   work; a native plugin is a second codebase.
7. **How much maturity work before breadth?** smith's first commit is dated 2026-10-15 and
   one maintainer carries it. Items 1 to 5 are the maturity work. Items 6 to 14 are breadth.
   Doing both at once is a staffing question.
8. **Which CI providers?** Jenkins is the one provider Nx documents as unsupported; CircleCI
   and Bitbucket Pipelines are the ones it generates. Item 9 should follow demand.

## Issues for this proposal

The epic [#575](https://github.com/monosmith-app/monosmith/issues/575) links this page and the issues
above. The five gaps that had no issue each have one: boundary rules (#570), dependency
tooling beyond Node and Python (#572), CircleCI and Bitbucket Pipelines (#571), release
archives for apps that are not Go (#573) and task-level affected (#574). Trusted
publishing for more registries (item 11) has no issue yet.

See also: [Everything compared](../product/comparison/everything-compared.md),
[Working on smith](development.md), [Decision records](../architecture/decisions/README.md).
