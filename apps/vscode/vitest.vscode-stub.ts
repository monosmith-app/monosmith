// The unit specs run in Node, where the `vscode` module does not exist. A slice index that also exports a
// prompt client imports it, so the specs of the slices that read that index load this empty stand-in.
export {}
