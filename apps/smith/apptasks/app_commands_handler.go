package apptasks

import (
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
	"github.com/spf13/cobra"
)

// FuncCommand is `smith func`: the steps that the tasks of a function app run
// (assemble the build folder, pack the zip). Hidden: the tasks call it.
func FuncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "func",
		Short:  "The steps that build and pack a function app (Azure Functions, AWS Lambda, Google Cloud), run by its tasks",
		Hidden: true,
	}
	var withDependencies bool
	pack := &cobra.Command{
		Use:   "package",
		Short: "Build the function app, then zip it into its build folder",
		Long: `Build the function app of the folder you are in, then zip the compiled code, the manifest
and host.json into dist/<project dir>/<name>.zip. A library of the workspace that the app
needs is built and copied into the zip. With --with-dependencies the production dependencies
are installed into the zip too, for a host that does not install them itself (Azure Functions).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := foldersHere()
			if err != nil {
				return err
			}
			zip, err := functionhost.PackFunctionApp(functionhost.PackOptions{ProjectDir: f.Project, BuildDir: f.Build, Name: f.Name, WithDependencies: withDependencies})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), zip)
			return nil
		},
	}
	pack.Flags().BoolVar(&withDependencies, "with-dependencies", false, "install the production dependencies into the zip")
	assemble := &cobra.Command{
		Use:   "assemble",
		Short: "Copy package.json and host.json next to the compiled code in the build folder",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			f, err := foldersHere()
			if err != nil {
				return err
			}
			return functionhost.AssembleFunctionApp(f.Project, f.Build)
		},
	}
	cmd.AddCommand(assemble, pack, pythonStageCommand(), pythonZipCommand())
	return cmd
}

// pythonStageCommand is `smith func stage`, the build task of a Python
// function app.
func pythonStageCommand() *cobra.Command {
	var host, python string
	cmd := &cobra.Command{
		Use:   "stage",
		Short: "Copy the files a Python function app deploys, and the code of the workspace libraries it needs, into the build folder",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			f, err := foldersHere()
			if err != nil {
				return err
			}
			_, err = functionhost.StagePythonApp(functionhost.StageOptions{ProjectDir: f.Project, BuildDir: f.Build, Host: host, Python: python})
			return err
		},
	}
	cmd.Flags().StringVar(&host, "host", "", "the function host: azure-functions, aws-lambda or gcp-functions")
	cmd.Flags().StringVar(&python, "python", "python", "the interpreter of the workspace venv, which finds the workspace libraries")
	_ = cmd.MarkFlagRequired("host")
	return cmd
}

// pythonZipCommand is `smith func zip`, the package task of a Python function
// app.
func pythonZipCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "zip",
		Short: "Zip the staged files of a Python function app into its build folder",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := foldersHere()
			if err != nil {
				return err
			}
			zip, err := functionhost.ZipPythonApp(f.Build, f.Name)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), zip)
			return nil
		},
	}
}

// ElectronCommand is `smith electron`: the step that the build task of an
// Electron app runs. Hidden: the task calls it.
func ElectronCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "electron",
		Short:  "The step that builds an Electron app, run by its build task",
		Hidden: true,
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "assemble",
		Short: "Write the manifest, the Forge configuration and the page next to the compiled code in the work folder",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			f, err := foldersHere()
			if err != nil {
				return err
			}
			return functionhost.AssembleElectronApp(f.Project, f.ElectronApp)
		},
	})
	return cmd
}

// foldersHere is the folders of the project in the current folder.
func foldersHere() (Folders, error) {
	here, err := os.Getwd()
	if err != nil {
		return Folders{}, err
	}
	root, err := workspace.FindRoot(here)
	if err != nil {
		return Folders{}, err
	}
	return AppFolders(root, here)
}
