# go

The smallest smith **language plugin**, in Go on the plugin SDK: it discovers `widget.json` projects
and gives each a cached `hello` task. Discovery only. For the full set of methods see
[`language-widget`](../language-widget); the same program without an SDK is in [`node`](../node) and
[`python`](../python).

```sh
GOWORK=off go build -o smith-plugin-widget .
smith plugin check ./smith-plugin-widget
```

Walk-through: [Writing a plugin](../../../guides/writing-a-plugin.md) and
[Creating a language plugin](../../../plugins/creating/language-plugin.md).
