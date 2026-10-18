package angular

import (
	"maps"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/typescript"
)

// Versions the starter pins; bump them deliberately, Angular's packages
// together. Angular 22 compiles with TypeScript 6.0.
const angularVersion = "^22.2.3"

var devDependencies = map[string]string{
	"@angular/build":        angularVersion,
	"@angular/cli":          angularVersion,
	"@angular/compiler-cli": angularVersion,
	"typescript":            "~6.0.3",
}

// templates are the app: the greeting slice, the manifest, the workspace
// file the CLI builds from, and a root component that greets.
var templates = withGreetingSlice(map[string]string{
	"[[.Dir]]/package.json": `{
  "name": "[[.Package]]",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "smith": { "kind": "app"[[if .Stack]], "stack": "[[.Stack]]"[[end]] },
  "scripts": {
    "build": "ng build",
    "typecheck": "ngc -p tsconfig.app.json",
    "start": "ng serve"
  },
  "dependencies": {
    "@angular/common": "` + angularVersion + `",
    "@angular/compiler": "` + angularVersion + `",
    "@angular/core": "` + angularVersion + `",
    "@angular/platform-browser": "` + angularVersion + `",
    "rxjs": "^7.8.2",
    "tslib": "^2.8.1"
  }
}
`,
	// The CLI's cache, never a source.
	"[[.Dir]]/.gitignore": `.angular/
`,
	"[[.Dir]]/angular.json": `{
  "$schema": "../../node_modules/@angular/cli/lib/config/schema.json",
  "version": 1,
  "cli": { "analytics": false },
  "newProjectRoot": ".",
  "projects": {
    "[[.Name]]": {
      "projectType": "application",
      "root": "",
      "sourceRoot": "src",
      "architect": {
        "build": {
          "builder": "@angular/build:application",
          "options": {
            "browser": "src/main.ts",
            "index": "index.html",
            "tsConfig": "tsconfig.app.json",
            "outputPath": { "base": "[[.Build]]", "browser": "" }
          },
          "configurations": {
            "production": { "outputHashing": "all" },
            "development": { "optimization": false, "sourceMap": true }
          },
          "defaultConfiguration": "production"
        },
        "serve": {
          "builder": "@angular/build:dev-server",
          "configurations": {
            "production": { "buildTarget": "[[.Name]]:build:production" },
            "development": { "buildTarget": "[[.Name]]:build:development" }
          },
          "defaultConfiguration": "development"
        }
      }
    }
  }
}
`,
	"[[.Dir]]/tsconfig.json": `{
  "compilerOptions": {
    "target": "es2023",
    "lib": ["es2023", "dom", "dom.iterable"],
    "module": "preserve",
    "moduleResolution": "bundler",
    "strict": true,
    "isolatedModules": true,
    "skipLibCheck": true,
    "noEmit": true,
    "allowImportingTsExtensions": true,
    "forceConsistentCasingInFileNames": true
  },
  "angularCompilerOptions": {
    "strictTemplates": true,
    "strictInjectionParameters": true
  },
  "include": ["src/**/*.ts"]
}
`,
	"[[.Dir]]/tsconfig.app.json": `{
  "extends": "./tsconfig.json",
  "files": ["src/main.ts"],
  "include": ["src/**/*.ts"],
  "exclude": ["src/**/*.[[.TestSuffix]].ts"]
}
`,
	"[[.Dir]]/index.html": `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>[[.Name]]</title>
  </head>
  <body>
    <app-root></app-root>
  </body>
</html>
`,
	"[[.Dir]]/src/main.ts": `import { bootstrapApplication } from '@angular/platform-browser'
import { AppComponent } from './app/index.ts'

await bootstrapApplication(AppComponent)
`,
	"[[.Dir]]/src/app/app.component.ts": `import { Component } from '@angular/core'
import { greet } from '../greeting'

/** The app's root component. */
@Component({
  selector: 'app-root',
  template: '<h1>{{ message }}</h1>',
})
export class AppComponent {
  protected readonly message = greet('world').message
}
`,
	"[[.Dir]]/src/app/index.ts": `export { AppComponent } from './app.component.ts'
`,
})

func withGreetingSlice(files map[string]string) map[string]string {
	maps.Copy(files, typescript.GreetingSlice)
	return files
}
