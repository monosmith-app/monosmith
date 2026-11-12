package functionhost

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Electron is the name that stands for an Electron app where a function host
// is named: it shares the assemble and pack commands of the function hosts.
const Electron = "electron"

// TriggerHTTP is the trigger whose start command differs from the others'.
const TriggerHTTP = "http"

// AssembleTask is the package.json script that builds an app: the compiler,
// then the smith command that writes the rest of the build folder.
//
// Parameters: host is a host choice name or Electron.
// Returns the script text.
func AssembleTask(host string) string {
	return "tsc -p tsconfig.build.json && " + AssembleCommand(host)
}

// AssembleCommand is the smith command that writes what a build folder holds
// beside the compiled code.
//
// Parameters: host is a host choice name or Electron.
// Returns the command text.
func AssembleCommand(host string) string {
	if host == Electron {
		return "smith electron assemble"
	}
	return "smith func assemble"
}

// PackageTask is the package.json script that packs an app.
//
// Parameters: host is a host choice name or Electron; root is the
// workspace-relative project folder.
// Returns the script text.
func PackageTask(host, root string) string {
	switch host {
	case Electron:
		return "npm run build && cd " + electronAppDir(root) + " && npx electron-forge package"
	case Azure:
		return "smith func package --with-dependencies"
	default:
		return "smith func package"
	}
}

// StartTask is the package.json script that runs an app locally with the tool
// of its host, from the build.
//
// Parameters: host is a host choice name or Electron; root is the
// workspace-relative project folder; trigger is the function's trigger
// (TriggerHTTP, "queue", "blob" or "timer"), unused for Azure and Electron.
// Returns the script text.
func StartTask(host, root, trigger string) string {
	build := protocol.BuildDirFromProject(root)
	switch host {
	case Electron:
		return "npm run build && cd " + electronAppDir(root) + " && npx electron ."
	case Azure:
		return "npm run build && func start --script-root " + build
	case AWS:
		if trigger == TriggerHTTP {
			return "npm run build && sam local start-api"
		}
		return "npm run build && sam local invoke GreetingFunction --event events/greeting.json"
	default:
		target := map[string]string{TriggerHTTP: "hello", "queue": "processMessage"}[trigger]
		if target == "" {
			target = "processObject"
		}
		text := "npm run build && cd " + build + " && npx functions-framework --target=" + target
		if trigger != TriggerHTTP {
			text += " --signature-type=cloudevent"
		}
		return text
	}
}

// electronAppDir is the work folder of an Electron app, seen from the project.
func electronAppDir(root string) string {
	return protocol.FromProject(root, protocol.BuildDir+"/electron/"+strings.Trim(root, "/")+"/app")
}
