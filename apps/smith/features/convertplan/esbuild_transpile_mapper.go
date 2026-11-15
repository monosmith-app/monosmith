package convertplan

import (
	"fmt"
	"strings"
)

// esbuildTranspileScript is what @nx/esbuild does with bundle: false for
// anything but Node CommonJS: it transpiles the main file and every file the
// project's tsconfig includes (minus what it excludes), each into the same
// place under the output directory, with nothing bundled. TypeScript reads
// the tsconfig (comments, extends, include and exclude, as tsc does), and
// esbuild's API takes the files, since its CLI cannot exclude any. It runs
// as `node -e`, so it works in every shell npm uses.
const esbuildTranspileScript = `const ts=require('typescript'),p=require('node:path'),c=ts.getParsedCommandLineOfConfigFile('%[1]s',{},` +
	`{...ts.sys,onUnRecoverableConfigFileDiagnostic:d=>{throw new Error(ts.flattenDiagnosticMessageText(d.messageText,' '))}});` +
	`require('esbuild').buildSync({entryPoints:[...new Set(['%[2]s',...c.fileNames].map(f=>p.resolve(f)))],` +
	`outdir:'%[3]s',platform:'%[4]s',format:'%[5]s',sourcemap:true,tsconfig:'%[1]s'})`

// EsbuildTranspile is the command for an @nx/esbuild build with bundle:
// false, paths relative to the project. ok is false for a path a shell
// would read.
func EsbuildTranspile(tsconfig, main, outDir, platform, format string) (line string, ok bool) {
	for _, s := range []string{tsconfig, main, outDir, platform, format} {
		if s == "" || strings.ContainsAny(s, "'\"`$%\\!") {
			return "", false
		}
	}
	return `node -e "` + fmt.Sprintf(esbuildTranspileScript, tsconfig, main, outDir, platform, format) + `"`, true
}
