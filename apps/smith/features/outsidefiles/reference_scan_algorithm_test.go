package outsidefiles

import (
	"reflect"
	"testing"
)

func TestMSBuildRefsTakesThePathsThatLeaveTheFolder(t *testing.T) {
	text := `<Project Sdk="Microsoft.NET.Sdk">
  <Import Project="..\build\common.props" />
  <PropertyGroup>
    <PackageLicenseFile>LICENSE</PackageLicenseFile>
    <AssemblyOriginatorKeyFile> ../signing.snk </AssemblyOriginatorKeyFile>
    <ApplicationIcon>$(SolutionDir)app.ico</ApplicationIcon>
  </PropertyGroup>
  <ItemGroup>
    <None Include="..\LICENSE">
      <Pack>True</Pack>
    </None>
    <None Include="..\LICENSE" />
    <Compile Include="..\shared\*.cs" />
    <Content Include="local.txt;..\notes.txt" />
    <PackageReference Include="Newtonsoft.Json" Version="13.0.1" />
    <ProjectReference Include="..\Other\Other.csproj" />
  </ItemGroup>
</Project>`
	want := []Ref{
		{Value: `..\build\common.props`, Kind: KindFile},
		{Value: `..\LICENSE`, Kind: KindFile},
		{Value: `..\notes.txt`, Kind: KindFile},
		{Value: `..\Other\Other.csproj`, Kind: KindProject},
		{Value: `../signing.snk`, Kind: KindFile},
	}
	if got := MSBuildRefs(text); !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}
}

func TestPackageJSONRefsTakesTheFilesEntriesThatLeaveTheFolder(t *testing.T) {
	got := PackageJSONRefs([]byte(`{"name":"a","files":["dist","../LICENSE","../docs/*.md"]}`))
	if want := []Ref{{Value: "../LICENSE", Kind: KindFile}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
	if PackageJSONRefs([]byte(`not json`)) != nil {
		t.Fatal("a file that is not JSON names nothing")
	}
}
