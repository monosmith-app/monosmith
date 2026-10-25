#!/usr/bin/env sh
# An end-to-end test of the plugin in a real workspace: declare it, create
# projects with it and run their tasks twice, expecting a cache hit.
# Needs smith and the .NET SDK (8.0 or later). Builds the plugin first. Prints "ok"
# and exits 0 when everything holds.
set -eu

here=$(cd "$(dirname "$0")" && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

(cd "$here" && dotnet build -c Release -nologo -v q >/dev/null)
plugin="$here/bin/Release/net8.0/LanguageWidget.dll"
# In Git Bash on Windows, smith needs a Windows path in smith.toml.
if command -v cygpath >/dev/null 2>&1; then
  plugin=$(cygpath -m "$plugin")
fi

cd "$tmp"
smith new ws >/dev/null 2>&1
cd ws
cat >> smith.toml <<TOML

[plugins.external.widget]
command = ["dotnet", "$plugin"]
TOML

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
