package scaffolding

// templates are every project's: the rockspec (named for its package and
// version, as LuaRocks expects), a module, its busted spec and the luacheck
// rules. The source URL and the license are what luarocks write_rockspec
// leaves for the author to fill in.
var templates = map[string]string{
	"[[.Dir]]/[[.Name]]-0.0.0-1.rockspec": `-- smithKind: [[.Kind]]
[[- if .Stack]]
-- smithStack: [[.Stack]]
[[- end]]
rockspec_format = "3.0"
package = "[[.Name]]"
version = "0.0.0-1"
source = {
  url = "*** please add URL for source tarball, zip or repository here ***",
}
description = {
  summary = "A [[.Kind]] of the [[.Workspace]] workspace.",
  license = "MIT",
}
dependencies = {
  "lua >= 5.1",
}
test_dependencies = {
  "busted",
}
build = {
  type = "builtin",
  modules = {
    ["[[.Name]]"] = "src/[[.Name]].lua",
  },
[[- if eq .Kind "app"]]
  install = {
    bin = { ["[[.Name]]"] = "bin/[[.Name]].lua" },
  },
[[- end]]
}
`,
	"[[.Dir]]/src/[[.Name]].lua": `local M = {}

--- Greets someone by name.
-- @param who string
-- @return string
function M.greet(who)
  return "Hello, " .. who .. "!"
end

return M
`,
	"[[.Dir]]/spec/[[.Name]]_spec.lua": `local mod = require("[[.Name]]")

describe("[[.Name]]", function()
  it("greets", function()
    assert.are.equal("Hello, world!", mod.greet("world"))
  end)
end)
`,
	"[[.Dir]]/.luacheckrc": `std = "max"
files["spec/**/*_spec.lua"] = { std = "+busted" }
`,
}

// appTemplates are an app's script, which the rockspec installs.
var appTemplates = map[string]string{
	"[[.Dir]]/bin/[[.Name]].lua": `local mod = require("[[.Name]]")

print(mod.greet("world"))
`,
}
