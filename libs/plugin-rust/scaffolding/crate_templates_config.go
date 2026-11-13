package scaffolding

// The greeting slice is a worked example of a vertical slice in Rust: a
// module per outcome, a file per role, tests beside the code.
var sliceTemplates = map[string]string{
	"[[.Dir]]/src/greeting/mod.rs": `//! Greets people: a worked example of a slice. Replace it with your own outcome.

mod greet_use_case;

pub use greet_use_case::greet;
`,
	"[[.Dir]]/src/greeting/greet_use_case.rs": `/// Greets someone by name.
pub fn greet(name: &str) -> String {
    format!("Hello, {name}!")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn greets_by_name() {
        assert_eq!(greet("world"), "Hello, world!");
    }
}
`,
}

var manifestTemplate = map[string]string{
	"[[.Dir]]/Cargo.toml": `[package]
name = "[[.Crate]]"
version = "0.0.0"
[[- if .Edition]]
edition = "[[.Edition]]"
[[- else]]
edition.workspace = true
[[- end]]
[[- if eq .Kind "package"]]
description = "[[.Name]]"
license = "MIT"
[[- else]]
publish = false
[[- end]]

[package.metadata.smith]
kind = "[[.Kind]]"
[[- if .Stack]]
stack = "[[.Stack]]"
[[- end]]

[dependencies]
`,
}

var appTemplates = map[string]string{
	"[[.Dir]]/src/main.rs": `mod greeting;

fn main() {
    println!("{}", greeting::greet("world"));
}
`,
}

var libTemplates = map[string]string{
	"[[.Dir]]/src/lib.rs": `//! [[.Name]]: its code lives in the slice modules below, one per outcome.

pub mod greeting;
`,
}
