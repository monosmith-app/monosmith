# smith completion

`smith completion <shell>` prints the autocompletion script for `bash`, `fish`, `powershell`
or `zsh`. It is the standard command of the command-line framework smith is built on; it is
not listed among smith's own commands in `smith commands`, and it takes no flags except
`--help` (and, per shell, `--no-descriptions`).

## When to use it

Once per machine, to complete commands and flags in your shell.

## Synopsis

```sh
smith completion bash | fish | powershell | zsh
```

## Setup

Each shell's own help has the exact lines (`smith completion bash --help`, and so on). Bash
needs the `bash-completion` package:

```sh
source <(smith completion bash)                                  # this session
smith completion bash > /etc/bash_completion.d/smith             # every session, Linux
smith completion bash > $(brew --prefix)/etc/bash_completion.d/smith   # macOS
```

PowerShell:

```sh
smith completion powershell | Out-String | Invoke-Expression
```

To load it in every session, add the output of that command to your PowerShell profile.
Each shell's script takes `--no-descriptions` to leave the descriptions out.

## Exit codes

0.

## Files and environment

Prints to standard output; writes nothing. No workspace is needed.

## Examples

```sh
smith completion bash
```

```text
# bash completion V2 for smith                                -*- shell-script -*-

__smith_debug()
{
...
```

## Related

[`smith commands`](commands.md), [Installing](../getting-started/installing.md).
