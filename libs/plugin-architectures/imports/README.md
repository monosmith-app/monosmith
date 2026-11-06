# imports

Reads the imports of a source file, for the architecture checks.

## Written exception

| Path | Rule waived | Constraint | Owner | Temporary |
|---|---|---|---|---|
| `libs/plugin-architectures/imports` (14 production files) | the 12-file review trigger | The slice has one outcome: the imports of a file, whatever its language. It holds one `<language>_imports_mapper.go` for each of the 11 languages, plus the catalogue, the contract, the project detection and the path algorithm that all of them use. A mapper has no meaning without its siblings, and a package for each language would copy the contract. The file count grows by one with each new language. | the smith maintainers | no |
