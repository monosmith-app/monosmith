// Package dependencies lists and changes the NuGet packages a .NET workspace
// declares: each project's and tests project's PackageReference, the
// workspace's Directory.Build.props, and the central versions in
// Directory.Packages.props, so `smith sync`, `smith up` and the one-version
// check cover .NET. Adding and removing packages goes through dotnet itself.
package dependencies
