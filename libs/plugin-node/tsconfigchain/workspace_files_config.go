package tsconfigchain

// BaseFile is the content of the root tsconfig.base.json a new workspace gets.
const BaseFile = `{
  "extends": "` + Preset + `"
}
`

// NodeRoot are the root files of a Node project's chain: the base. A stack adds
// them to its template set; a file that exists is never overwritten.
func NodeRoot() map[string]string {
	return map[string]string{Base: BaseFile}
}
