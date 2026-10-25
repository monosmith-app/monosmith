#!/usr/bin/env sh
# An end-to-end test of the plugin in a real workspace: declare it, create
# projects with it and run their tasks twice, expecting a cache hit.
# Needs smith and Node.js 24 on the path (run `npm install` first, outside this
# repository, so that @monosmith/plugin-sdk is installed). Prints "ok" and
# exits 0 when everything holds.
set -eu

here=$(cd "$(dirname "$0")" && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

plugin="$here/plugin.ts"
# In Git Bash on Windows, smith needs a Windows path in smith.toml.
if command -v cygpath >/dev/null 2>&1; then plugin=$(cygpath -m "$plugin"); fi

cd "$tmp"
smith new ws >/dev/null 2>&1
cd ws
cat >> smith.toml <<EOF

[plugins.external.widget]
command = ["node", "$plugin"]
EOF

smith add lib core --lang widget >/dev/null
smith add app shop --lang widget >/dev/null
sed -i 's/"license": "MIT"/"license": "MIT",\n  "uses": ["core"]/' apps/shop/widget.json

smith projects | grep -q '^shop .*app .*widget' || { echo "shop is not discovered as a widget app"; exit 1; }
smith graph | grep -q 'core' || { echo "the shop -> core edge is missing"; exit 1; }

smith run build test >/dev/null
smith run build test 2>&1 | grep -q 'from cache' || { echo "the second run was not served from the cache"; exit 1; }

echo ok
