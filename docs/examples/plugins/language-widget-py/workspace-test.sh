#!/usr/bin/env sh
# An end-to-end test of the plugin in a real workspace: declare it, create
# projects with it and run their tasks twice, expecting a cache hit.
# Needs smith and a Python (3.10 or later) with monosmith-plugin-sdk installed:
# set PYTHON to that interpreter (default: python). Prints "ok" and exits 0 when
# everything holds.
set -eu

here=$(cd "$(dirname "$0")" && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

python=${PYTHON:-python}
plugin="$here/plugin.py"
# In Git Bash on Windows, smith needs a Windows path in smith.toml.
if command -v cygpath >/dev/null 2>&1; then
  plugin=$(cygpath -m "$plugin")
  python=$(cygpath -m "$(command -v "$python")")
fi

cd "$tmp"
smith new ws >/dev/null 2>&1
cd ws
cat >> smith.toml <<EOF

[plugins.external.widget]
command = ["$python", "$plugin"]
EOF

# smith starts an external plugin only once it is trusted (in CI, only when pinned).
smith plugin trust widget --yes --pin >/dev/null

smith add lib core --lang widget >/dev/null
smith add app shop --lang widget >/dev/null
sed -i 's/"license": "MIT"/"license": "MIT",\n  "uses": ["core"]/' apps/shop/widget.json

smith projects | grep -q '^shop .*app .*widget' || { echo "shop is not discovered as a widget app"; exit 1; }
smith graph | grep -q 'core' || { echo "the shop -> core edge is missing"; exit 1; }

smith run build test >/dev/null
smith run build test 2>&1 | grep -q 'from cache' || { echo "the second run was not served from the cache"; exit 1; }

echo ok
