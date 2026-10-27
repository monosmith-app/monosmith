// Package dependencies lists and changes the dependency versions a JVM
// workspace declares in its build files: pom.xml (dependencies,
// dependencyManagement and the properties they use) and the string notation
// of Gradle build scripts, so `smith sync` and `smith up` cover the JVM
// without installing any Maven or Gradle plugin. Adding and removing
// dependencies is the build tool's own job and is not offered.
package dependencies
