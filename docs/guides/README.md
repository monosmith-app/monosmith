# Guides

Task-oriented pages for people who already have a workspace.

| Guide | Read it when |
|---|---|
| [Editors](editors.md) | you integrate smith with an editor or an AI assistant: the `--json` contracts, `smith serve` and `smith mcp` |
| [The MonoSmith agent](agent.md) | you want Claude Code or VS Code to work in the workspace through smith: what `smith agent install` writes, what the agent does alone and what it asks |
| [Using derived files](derived-files.md) | config repeats in every project and follows from the graph: let smith write it, keep a record, and report a hand edit instead of overwriting it |
| [Know your files, share the workspace and brief your agents](know-your-files-and-brief-your-agents.md) | you want to know who owns a file (`smith why`), attach the workspace state to a pull request (`smith report`), or give coding agents the rules (`smith agent file`) |
| [Troubleshooting](troubleshooting.md) | something does not work: every doctor check with its remedy, cache surprises, install problems |
| [Private release files](private-release-assets.md) | another repository's CI must download the release files of a private repository: the fine-grained token, the secret and the download step |
| [FAQ](faq.md) | you want a short answer to a common question |
| [Writing a plugin](writing-a-plugin.md) | smith lacks a language or a rule you need: a pointer to [Creating plugins](../plugins/creating/README.md), which has a guide for each kind of plugin, with working examples in [`examples/plugins`](../examples/plugins/README.md) |

Getting started is in [Getting started](../getting-started/README.md); every command is in the
[manual](../manual/README.md).
