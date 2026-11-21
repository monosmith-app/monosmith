# ADR 0109: smith offers trusted publishing

- Status: accepted
- Date: 2026-11-21
- Amends [ADR 0031](0031-marketplace-publishing-with-entra-id.md).

## Context

Three `[release]` settings let the GitHub workflow publish with its OIDC token instead of a
long-lived secret: `nuget_trusted_publishing`, `rubygems_trusted_publishing` and `vsce_entra`.
Nothing offered them. A person had to know the key and edit smith.toml. The owner's rule is
"ask, never impose": a decision is a presented choice, with the options explained.

## Decision

1. The three settings can be absent (not decided), `true` or `false`. `false` records the answer
   "keep the secret", so smith does not ask again. Absent and `false` generate the same
   pipelines.
2. `smith ci generate` collects the secrets the pipelines need. For each of `NUGET_API_KEY`,
   `GEM_HOST_API_KEY` and `VSCE_PAT` whose setting is absent, when every CI provider is GitHub,
   it asks. The question names the secret, lists the one-time set-up on the registry with the
   owner, repository, workflow file and branches smith knows, and offers trusted publishing (the
   default, recommended) or the API key secret. smith writes the answer to smith.toml before it
   shows the plan of the pipeline files.
3. A run without a person (`--yes`, `--dry-run`, CI, no terminal) does not ask and changes
   nothing. It prints a note for each undecided setting, with the `smith config set` command
   that turns it on. `smith add`, `smith import` and `smith adopt`, which write the pipelines on
   the way, print the same note.
4. Workspaces with a provider other than GitHub are never asked: the sign-ins work on GitHub
   Actions only.

The `trustedpublishing` slice holds the options, the policy (`Pending`) and the question
(`Offer`). `cipipeline.OfferSignIns` connects it to the collected contributions.

## Consequences

- A new NuGet, RubyGems or Marketplace workspace meets the keyless sign-in the first time it
  generates its pipelines.
- The VS Code extension and the TUI run `smith ci generate --yes`, so they show the note and do
  not ask.
- The answer is written to smith.toml even when the person then declines the plan of the
  pipeline files: it is a decision of its own.
