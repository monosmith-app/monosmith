package pluginusage

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/spf13/cobra"
)

// GuardDisable makes `smith plugin disable` count the projects the plugins
// find. With any, the command refuses unless --force is given; with --force it
// says the count and goes on. When the workspace cannot be loaded the count is
// unknown and the command goes on as before.
//
// Parameters: plugin is the `smith plugin` command; version is smith's version.
func GuardDisable(plugin *cobra.Command, version string) {
	for _, sub := range plugin.Commands() {
		if sub.Name() != "disable" {
			continue
		}
		run := sub.RunE
		sub.Flags().Bool("force", false, "switch the plugin off even when projects use it")
		sub.RunE = func(c *cobra.Command, args []string) error {
			used, err := Projects(".", version, args)
			if err == nil {
				if warning := Warning(used, args); warning != "" {
					if force, _ := c.Flags().GetBool("force"); !force {
						return problem.New("switching off hides projects: "+warning,
							"smith plugin disable "+strings.Join(args, " ")+" --force switches it off anyway")
					}
					fmt.Fprintf(c.ErrOrStderr(), "hiding projects, as --force says: %s\n", warning)
				}
			}
			return run(c, args)
		}
	}
}
