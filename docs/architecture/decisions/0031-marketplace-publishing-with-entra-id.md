# ADR 0031: Publish VS Code extensions with Microsoft Entra ID instead of a token

- Status: accepted
- Date: 2026-10-23

## Context

Azure DevOps retires global personal access tokens on 1 December 2026, and the Marketplace
publish used `VSCE_PAT`.

## Options considered

A `VSCE_PAT` secret (the previous behaviour) against an Entra ID app registration signed in
with the workflow's OIDC token (workload identity federation). The same shape had been used
for RubyGems trusted publishing
([#417](https://github.com/monosmith-app/monosmith/pull/417)).

## Decision

`[release] vsce_entra = true` publishes as an Entra ID app. The generated workflow gets
`id-token: write` and a `Sign in to Azure` step (`azure/login`, with the `AZURE_CLIENT_ID`
and `AZURE_TENANT_ID` secrets) before Release, only on a push to the release branch, and the
node plugin runs `vsce publish --azure-credential`. `VSCE_PAT` is neither a workflow secret
nor a preflight requirement with it on. GitHub Actions only, enforced when the configuration
is loaded. This repository turns it on.

## Consequences

One-time setup: an app registration with a federated credential for the release branch, its
Azure DevOps ID added to the publisher as Contributor, and the two repository secrets. Until
they exist the release fails at the Azure sign-in. The pull request states that nothing had
been published with Entra yet when it merged.

## Sources

- Pull request [#429](https://github.com/monosmith-app/monosmith/pull/429)
- [Releasing](../../reference/releasing.md)
