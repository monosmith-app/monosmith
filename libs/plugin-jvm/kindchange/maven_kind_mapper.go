package kindchange

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/fileedits"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/xmltext"
)

var relativePath = regexp.MustCompile(`(<relativePath>)([^<]*)(</relativePath>)`)

// changeMaven writes a Maven module's kind (the smith.kind property) and what
// an app needs that the others do not: the exec plugin's properties.
func changeMaven(edits *fileedits.Set, params protocol.ChangeKindParams, result *protocol.ChangeKindResult) error {
	p := params.Project
	manifest := path.Join(p.Root, "pom.xml")
	src, _ := edits.Read(manifest)

	src = setProperty(src, "smith.kind", string(params.To))
	switch params.To {
	case protocol.ProjectApp:
		if mainClass := mainClassOf(edits, p.Root); mainClass != "" {
			src = setProperty(src, "exec.skip", "false")
			src = setProperty(src, "exec.mainClass", mainClass)
		} else {
			warnf(result, "an app runs a program, and this module has no Main class: add one and set exec.mainClass in %s before running it", manifest)
		}
	default:
		_, program := xmltext.Element(src, "exec.mainClass")
		src, _ = xmltext.RemoveElement(src, "exec.skip")
		src, _ = xmltext.RemoveElement(src, "exec.mainClass")
		if program {
			warnf(result, "this module has a program (a Main class); as a %s it is a dependency of other modules, and the program is kept but not deployed", params.To)
		}
	}
	if params.To == protocol.ProjectPackage {
		for _, missing := range missingPomFields(src) {
			warnf(result, "a Maven repository asks for %s in %s before the first release; add it", missing, manifest)
		}
	}
	edits.Write(manifest, src)
	return nil
}

func missingPomFields(src string) []string {
	var missing []string
	if _, ok := xmltext.Element(src, "description"); !ok {
		missing = append(missing, "a description")
	}
	if !strings.Contains(src, "<licenses>") {
		missing = append(missing, "a license")
	}
	return missing
}

// setProperty sets a pom property, adding the properties block when the pom
// has none.
func setProperty(src, name, value string) string {
	if out, ok := xmltext.AppendElement(src, "properties", name, value); ok {
		return out
	}
	eol := "\n"
	if strings.Contains(src, "\r\n") {
		eol = "\r\n"
	}
	block := fmt.Sprintf("  <properties>%s    <%s>%s</%s>%s  </properties>%s", eol, name, value, name, eol, eol)
	if i := strings.LastIndex(src, "</project>"); i >= 0 {
		return src[:i] + block + src[i:]
	}
	return src
}

// mainClassOf finds the module's Main class from its sources: the package is
// the folders under src/main/java (or kotlin).
func mainClassOf(edits *fileedits.Set, root string) string {
	for _, file := range edits.Walk(root, func(name string) bool { return name == "Main.java" || name == "Main.kt" }) {
		for _, marker := range []string{"/src/main/java/", "/src/main/kotlin/"} {
			_, rest, ok := strings.Cut(file, marker)
			if !ok {
				continue
			}
			class := strings.ReplaceAll(path.Dir(rest), "/", ".") + "." + strings.TrimSuffix(path.Base(rest), path.Ext(rest))
			if path.Ext(rest) == ".kt" {
				class += "Kt"
			}
			return strings.TrimPrefix(class, "..")
		}
	}
	return ""
}
