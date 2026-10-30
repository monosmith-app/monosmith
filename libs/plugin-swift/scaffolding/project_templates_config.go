package scaffolding

// templates are every project's: the manifest with the smith header comments
// under swift-tools-version (which must stay first), a source target, an
// XCTest target and a .gitignore for SwiftPM's working directories. An app's
// target is an executable whose main.swift prints a greeting; a lib's (or
// package's) is a library with a public function.
var templates = map[string]string{
	"[[.Dir]]/Package.swift": `// swift-tools-version: 5.9
// smithKind: [[.Kind]]
[[- if .Stack]]
// smithStack: [[.Stack]]
[[- end]]
// smithVersion: 0.0.0
import PackageDescription

let package = Package(
    name: "[[.Module]]",
[[- if ne .Kind "app"]]
    products: [
        .library(name: "[[.Module]]", targets: ["[[.Module]]"])
    ],
[[- end]]
    targets: [
[[- if eq .Kind "app"]]
        .executableTarget(name: "[[.Module]]"),
[[- else]]
        .target(name: "[[.Module]]"),
[[- end]]
        .testTarget(name: "[[.Module]]Tests", dependencies: ["[[.Module]]"]),
    ]
)
`,
	"[[.Dir]]/Tests/[[.Module]]Tests/[[.Module]]Tests.swift": `import XCTest
@testable import [[.Module]]

final class [[.Module]]Tests: XCTestCase {
[[- if eq .Kind "app"]]
    func testGreeting() {
        XCTAssertEqual(greeting(), "Hello from [[.Name]]")
    }
[[- else]]
    func testGreet() {
        XCTAssertEqual(greet("world"), "Hello, world!")
    }
[[- end]]
}
`,
	"[[.Dir]]/.gitignore": ".build/\n.swiftpm/\n",
}

// appTemplates are an executable target's entry point.
var appTemplates = map[string]string{
	"[[.Dir]]/Sources/[[.Module]]/main.swift": `func greeting() -> String {
    "Hello from [[.Name]]"
}

print(greeting())
`,
}

// libTemplates are a library target's public function.
var libTemplates = map[string]string{
	"[[.Dir]]/Sources/[[.Module]]/[[.Module]].swift": `/// Greets someone by name.
///
/// - Parameter who: The name to greet.
/// - Returns: The greeting.
public func greet(_ who: String) -> String {
    "Hello, \(who)!"
}
`,
}
