# smith in the workspace

MonoSmith does not carry smith. It runs the one your team installed, so the editor, the terminal and CI give the same result.

It looks in this order:

1. the `monosmith.path` setting
2. the workspace: `node_modules/.bin/smith` (`npm i -D monosmith`) or `.smith/bin/smith`, in a trusted workspace only
3. `smith` on the `PATH`

The status bar says which one runs and why.
