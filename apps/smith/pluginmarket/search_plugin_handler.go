package pluginmarket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// SearchCommand is `smith plugin search`. configured returns the [plugins]
// marketplace value of the workspace smith.toml ("" outside a workspace or
// when unset); client downloads the index.
func SearchCommand(client *http.Client, configured func() (string, error)) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "search [term]",
		Short: "Search the plugin marketplace",
		Long: `Lists the plugins of the marketplace whose name, description, roles, kinds or
languages contain every word of the term; without a term, lists them all. Each
line shows the newest version. Install one with "smith plugin install <name>".

The marketplace is the index of github.com/monosmith-app/plugins unless
[plugins] marketplace in smith.toml or the SMITH_MARKETPLACE environment
variable names another (an https URL or a file path; the variable wins). For a
private index smith uses GITHUB_TOKEN, GH_TOKEN or "gh auth token", and sends
the token to GitHub only.`,
		Example: "  smith plugin search\n  smith plugin search deploy go",
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := configured()
			if err != nil {
				return err
			}
			location := Location(value)
			idx, err := FetchIndex(cmd.Context(), client, location, GitHubToken)
			if err != nil {
				return err
			}
			found := Search(idx, strings.Join(args, " "))
			w := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				if found == nil {
					found = []Plugin{}
				}
				return enc.Encode(found)
			}
			if len(found) == 0 {
				fmt.Fprintf(w, "No plugin matches in %s.\n", location)
				return nil
			}
			fmt.Fprintf(w, "%-18s %-9s %-22s %s\n", "PLUGIN", "VERSION", "ROLES", "DESCRIPTION")
			for _, p := range found {
				fmt.Fprintf(w, "%-18s %-9s %-22s %s\n", p.Name, Newest(p),dash(strings.Join(p.Roles, ",")), p.Description)
			}
			fmt.Fprintf(w, "\nInstall one with: smith plugin install <name>[@version]\n")
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	return cmd
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
