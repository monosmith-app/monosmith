package guide

// addSetupKey is the key of the first question after the name of a new project.
const addSetupKey = "add-setup"

// addSetupStep is the first question after the name of a project: take the
// defaults and get to the plan, or answer each question. It uses the same two
// answers as the setup of a new workspace (setupDefaults, setupCustom). The
// defaults are what the stack suggests, so the plan can show every one of
// them as a flag.
func addSetupStep() Step {
	return Step{
		Key: addSetupKey, Prompt: "How do you want to set it up?",
		Note: "The defaults are the answers the stack suggests: the choices you picked above, or none, and the shape of the other projects of its kind.\n" +
			"A web app is built for lcl, dev, uat and prd: smith run build builds prd, the builds go to dist/apps/<name>/<env>, and the .env files are written for you and kept in git.\n" +
			"The plan shows every default as a flag, and nothing is written before you say so.",
		Options: []Option{
			{Label: "Start with the defaults", Value: setupDefaults, Hint: "go to the plan now; each default is in the command (default)"},
			{Label: "Customise", Value: setupCustom, Hint: "answer the questions one at a time: choices, the shape of the other projects, the environments"},
		},
	}
}
