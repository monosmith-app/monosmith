package projectimport

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// maxTries is how many bad answers AskFolder accepts before it gives up.
const maxTries = 3

// AskFolder asks where the project goes: the app, lib or package folder (the
// one the identification suggests is the default) or another folder. Nothing
// is changed.
//
// Parameters: name is the project's name; suggested is the identified kind;
// in and out are the terminal.
// Returns: the folder and its kind, or a problem after three bad answers or
// when no answer comes.
func AskFolder(name string, suggested protocol.ProjectKind, in io.Reader, out io.Writer) (string, protocol.ProjectKind, error) {
	r := bufio.NewReader(in)
	kinds := []protocol.ProjectKind{protocol.ProjectApp, protocol.ProjectLib, protocol.ProjectPackage}
	def := 1
	fmt.Fprintf(out, "\nWhere should %s go? The folder sets its kind.\n", name)
	for i, k := range kinds {
		mark := ""
		if k == suggested {
			def, mark = i+1, " (suggested)"
		}
		fmt.Fprintf(out, "  %d) %s/%s: %s%s\n", i+1, projectkind.Directory(k), name, kindWords(k), mark)
	}
	fmt.Fprintln(out, "  4) another folder")
	read := func(prompt string) (string, error) {
		fmt.Fprint(out, prompt)
		line, err := r.ReadString('\n')
		answer := strings.TrimSpace(strings.TrimPrefix(line, string(rune(0xFEFF))))
		if err != nil && answer == "" {
			return "", problem.New("no answer was given", "--as app|lib|package or --into <folder> says it without a question")
		}
		return answer, nil
	}
	for try := 0; try < maxTries; try++ {
		answer, err := read(fmt.Sprintf("Choose 1-4 [%d]: ", def))
		if err != nil {
			return "", "", err
		}
		n := def
		if answer != "" {
			if n, err = strconv.Atoi(answer); err != nil || n < 1 || n > 4 {
				fmt.Fprintf(out, "  %q is not a choice; type a number from 1 to 4.\n", answer)
				continue
			}
		}
		if n <= 3 {
			return Folder(kinds[n-1], name, "")
		}
		typed, err := read("Folder, from the workspace root (for example tools/" + name + "): ")
		if err != nil {
			return "", "", err
		}
		// The folder sets the kind; outside apps/, libs/ and packages/ the
		// suggested kind holds.
		dir, kind, err := Folder("", name, typed)
		if err != nil {
			dir, kind, err = Folder(suggested, name, typed)
		}
		if err != nil {
			fmt.Fprintln(out, "  "+err.Error())
			continue
		}
		return dir, kind, nil
	}
	return "", "", problem.New("no valid answer after three tries; nothing was changed", "--as app|lib|package or --into <folder> says it without a question")
}

func kindWords(k protocol.ProjectKind) string {
	switch k {
	case protocol.ProjectApp:
		return "an app, a program that runs and deploys"
	case protocol.ProjectLib:
		return "a lib, code the workspace's projects share"
	}
	return "a package, published to a registry"
}
