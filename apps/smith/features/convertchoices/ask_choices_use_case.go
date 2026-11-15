package convertchoices

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Ask puts each question the user has not answered on the command line, as a
// numbered list in the style of `smith doctor --fix`; an empty answer takes
// the suggested option (the first). It returns the choices with the answers
// in, and an error when an answer is not one of the options.
func Ask(in *bufio.Reader, out io.Writer, questions []Question, c Choices, g Given) (Choices, error) {
	given := map[string]bool{"lint": g.Lint, "test": g.Test, "build": g.Build, "ci": g.CI}
	for _, q := range questions {
		if given[q.Key] || len(q.Options) < 2 {
			if len(q.Options) == 1 && !given[q.Key] {
				c.Set(q.Key, q.Options[0].Value)
			}
			continue
		}
		fmt.Fprintf(out, "\n%s\n", q.Question)
		for i, o := range q.Options {
			suggested := ""
			if i == 0 {
				suggested = " (suggested)"
			}
			fmt.Fprintf(out, "  %d) %s%s\n", i+1, o.Label, suggested)
		}
		fmt.Fprint(out, "Choose [1]: ")
		line, _ := in.ReadString('\n')
		line = strings.TrimSpace(line)
		pick := 1
		if line != "" {
			n, err := strconv.Atoi(line)
			if err != nil || n < 1 || n > len(q.Options) {
				return c, fmt.Errorf("%q is not one of the options 1 to %d for %q", line, len(q.Options), q.Question)
			}
			pick = n
		}
		c.Set(q.Key, q.Options[pick-1].Value)
	}
	return c, nil
}
