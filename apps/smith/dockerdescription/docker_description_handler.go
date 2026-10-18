package dockerdescription

import (
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith docker description`.
func Command(version string) *cobra.Command {
	docker := &cobra.Command{
		Use:   "docker",
		Short: "Docker image commands",
	}
	var req Request
	description := &cobra.Command{
		Use:   "description",
		Short: "Set the Docker Hub description of the apps' images from their description_from file",
		Long: `description sets the Docker Hub repository description of each app whose
registries include a docker.io entry and name a description_from file, with the
registry's credentials (DOCKERHUB_USERNAME and DOCKERHUB_TOKEN, or the variables
its entry names). Editing a description needs a token with Read, Write & Delete
permission.

The call is an idempotent PATCH, so it can run on every push to the release
branch: the generated pipeline does, after the release, so a docs-only commit
that changes the file reaches Docker Hub without a new version. A failure is a
warning (the images are already published); --strict makes it an error.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			return Run(s.Root, s.Graph.Projects, req, cmd.OutOrStdout())
		},
	}
	description.Flags().StringSliceVarP(&req.Projects, "project", "p", nil, "only these apps")
	description.Flags().BoolVar(&req.DryRun, "dry-run", false, "name the repositories without calling Docker Hub")
	description.Flags().BoolVar(&req.Strict, "strict", false, "exit non-zero when a description could not be set")
	docker.AddCommand(description)
	return docker
}
