package manifestfixes

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// EmitDeclarationsDescription names the fix on every package it applies to,
// so the host folds them into one.
const EmitDeclarationsDescription = "emit type declarations (.d.ts) in the build"

// declarationConfigs are the tsconfig files a declaration emit can use, the
// package's own build config first.
var declarationConfigs = []string{"tsconfig.build.json", "tsconfig.lib.json", "tsconfig.json"}

// EmitDeclarationsFix appends tsc's declaration emit to the build script of
// the package p, writing into declarationDir (relative to the package, as the
// package.json types entry reads): "esbuild ... && tsc -p tsconfig.build.json
// --emitDeclarationOnly --declaration --declarationDir dist/src". The build
// stays one cached task whose output directory holds both. It returns nil
// when there is no build script to extend or no tsconfig to emit from.
func EmitDeclarationsFix(workspaceRoot string, p protocol.Project, declarationDir string) *protocol.Fix {
	manifest := path.Join(p.Root, "package.json")
	raw, err := readFile(workspaceRoot, manifest)
	if err != nil {
		return nil
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return nil
	}
	scripts := doc.Object("scripts", false)
	if scripts == nil {
		return nil
	}
	build, ok := scripts.String("build")
	if !ok || build == "" {
		return nil
	}
	tsconfig := ""
	for _, name := range declarationConfigs {
		if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), name)); err == nil {
			tsconfig = name
			break
		}
	}
	if tsconfig == "" || declarationDir == "" {
		return nil
	}
	scripts.SetString("build", build+" && tsc -p "+tsconfig+" --emitDeclarationOnly --declaration --declarationDir "+declarationDir)
	return &protocol.Fix{
		Description: EmitDeclarationsDescription,
		AppliesTo:   []string{p.Name},
		Files:       []protocol.FileEdit{{Path: manifest, Content: render(doc, raw)}},
	}
}
