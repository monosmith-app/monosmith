# python

The smallest smith **language plugin**, in Python with the standard library only and no SDK: it
discovers `widget.json` projects and gives each a cached `hello` task. Discovery only. The same
program in [`go`](../go) and [`node`](../node); a plugin with more methods and tests in
[`structure-layout-python`](../structure-layout-python).

```text
PS> smith plugin check .\plugin.cmd
```

Walk-through: [Writing a plugin in Go, Node.js or Python](../../../plugins/creating/plugin-sdks.md).
