package compose

import (
	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// offer is a choice of a fake plugin.
func offer(plugin string, c protocol.Choice) choiceselect.Offer {
	if c.Language == "" {
		c.Language = "node"
	}
	return choiceselect.Offer{Plugin: plugin, Choice: c}
}

var scaffoldRole = []protocol.Role{protocol.RoleScaffold}
var toolRole = []protocol.Role{protocol.RoleToolchain}

// fakeOffers are the choices of three fake plugins: a web starter, a desktop
// shell, a function host, and the tools they need.
func fakeOffers() []choiceselect.Offer {
	return []choiceselect.Offer{
		offer("vite", protocol.Choice{Axis: "framework", Name: "react", Title: "React", Aliases: []string{"web", "ui"}, Description: "a React app", Roles: append(append([]protocol.Role{}, scaffoldRole...), protocol.RoleToolchain), Needs: []string{"lint", "test"}}),
		offer("electron", protocol.Choice{Axis: "framework", Name: "electron", Title: "Electron", Layer: "shell", Aliases: []string{"desktop", "app"}, Description: "a desktop app", Roles: scaffoldRole, Needs: []string{"lint", "test"}, Conflicts: []string{"azure-functions"}}),
		offer("functions", protocol.Choice{Axis: "framework", Name: "azure-functions", Title: "Azure Functions", Layer: "host", Aliases: []string{"serverless", "cloud"}, Description: "an Azure function app", Roles: scaffoldRole, Needs: []string{"trigger", "lint", "test"}}),
		offer("functions", protocol.Choice{Axis: "trigger", Name: "http-trigger", Title: "HTTP trigger", Layer: "host", Suggested: true, Description: "answers a request"}),
		offer("functions", protocol.Choice{Axis: "trigger", Name: "timer-trigger", Title: "Timer trigger", Layer: "host", Description: "runs on a schedule"}),
		offer("eslint", protocol.Choice{Axis: "lint", Name: "eslint", Title: "ESLint", Roles: toolRole, Default: true, Description: "lint"}),
		offer("vitest", protocol.Choice{Axis: "test", Name: "vitest", Title: "Vitest", Roles: toolRole, Default: true, Description: "test"}),
		offer("vite", protocol.Choice{Axis: "style", Name: "bootstrap", Description: "css"}),
		offer("py", protocol.Choice{Axis: "framework", Name: "azure-functions", Language: "python", Title: "Azure Functions", Layer: "host", Aliases: []string{"serverless"}, Roles: scaffoldRole, Description: "python function app"}),
	}
}
