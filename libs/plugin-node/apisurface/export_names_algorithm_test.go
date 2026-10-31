package apisurface

import (
	"slices"
	"testing"
)

func TestParseExportsReadsEveryTopLevelExportForm(t *testing.T) {
	src := `
import { helper } from './helper'
// export const commented = 1
/* export const blocked = 1 */
const text = "export const inString = 1"
export const answer = 42
export let counter = 0, second = 1
export var legacy
export function run() { const inner = 1; return { export: inner } }
export async function load() {}
export function* walk() {}
export class Service { export() {} }
export abstract class Base {}
export interface Options { a: string }
export type Mode = 'a' | 'b'
export enum Color { Red }
export const enum Flag { On }
export declare const ambient: number
export namespace Tools { export const hidden = 1 }
export default class {}
export { helper, internal as renamed, type OnlyType }
export type { Shape }
export { thing as default } from './thing'
export * as ns from './ns'
export * from './more'
export = legacyStyle
`
	got := ParseExports("src/index.ts", src)
	want := []string{
		"Base", "Color", "Flag", "Mode", "Options", "OnlyType", "Service", "Shape", "Tools", "ambient", "answer",
		"counter", "default", "export=", "helper", "legacy", "load", "ns", "renamed", "run", "walk",
	}
	// "second" (a second declarator) and the exports inside a namespace, a
	// class and a function body are not part of the surface.
	if !slices.Equal(got.Names, sortedCopy(want)) {
		t.Errorf("names = %v\nwant    %v", got.Names, sortedCopy(want))
	}
	if !slices.Equal(got.Stars, []string{"./more"}) {
		t.Errorf("stars = %v, want [./more]", got.Stars)
	}
}

func TestParseExportsReadsCommonJSOutsideDeclarationFiles(t *testing.T) {
	src := `"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
Object.defineProperty(exports, "viaDefine", { enumerable: true, get: function () { return x; } });
exports.alpha = exports.beta = void 0;
module.exports.gamma = 3;
`
	got := ParseExports("dist/index.js", src)
	if want := []string{"alpha", "beta", "gamma", "viaDefine"}; !slices.Equal(got.Names, want) {
		t.Errorf("names = %v, want %v", got.Names, want)
	}
	obj := ParseExports("index.cjs", "module.exports = { one, two: 2, 'skip-me': 3 }\n")
	if want := []string{"one", "two"}; !slices.Equal(obj.Names, want) {
		t.Errorf("object names = %v, want %v", obj.Names, want)
	}
	whole := ParseExports("index.js", "module.exports = function () {}\n")
	if want := []string{"module.exports"}; !slices.Equal(whole.Names, want) {
		t.Errorf("whole names = %v, want %v", whole.Names, want)
	}
	if got := ParseExports("index.d.ts", "declare const exports: any\nexports.nope = 1\n"); len(got.Names) != 0 {
		t.Errorf("a declaration file has no CommonJS exports, got %v", got.Names)
	}
}

func TestParseExportsIgnoresStringsAndAnUnterminatedQuote(t *testing.T) {
	src := "const re = /it's/\nexport const after = 1\nconst t = `export const nope = 1 ${ '{' }`\nexport const last = 2\n"
	got := ParseExports("a.ts", src)
	if want := []string{"after", "last"}; !slices.Equal(got.Names, want) {
		t.Errorf("names = %v, want %v", got.Names, want)
	}
}

func sortedCopy(in []string) []string {
	out := slices.Clone(in)
	slices.Sort(out)
	return out
}
