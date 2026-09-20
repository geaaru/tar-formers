/*

Copyright © 2021-2026 Daniele Rondina <geaaru@macaronios.org>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.:s

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	executor "github.com/geaaru/tar-formers/pkg/executor"
	specs "github.com/geaaru/tar-formers/pkg/specs"

	"github.com/spf13/cobra"
)

func deleteContainer(name string) {
	deleteargs := []string{"rm", name}

	fmt.Println(":whale: deleting container with name " + name)
	_, err := exec.Command("docker", deleteargs...).CombinedOutput()
	if err == nil {
		fmt.Println("Container " + name + " removed.")
	}
}

func flatDockerContainer(tarformers *executor.TarFormers,
	sourceImage string, args *DockerImportArgs,
	specExporter, specImporter string,
	summary bool) error {

	// Create container from the image in order to use docker export command.
	createargs := []string{
		"create", sourceImage,
		"-c", "sleep", "1",
	}
	fmt.Println(":whale: Creating container from image " + sourceImage)

	// Creating a fake container to use for the export.
	out, err := exec.Command("docker", createargs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed creating container for image %s: %s",
			sourceImage, err.Error())
	}
	idcontainer := strings.TrimRight(string(out), "\n")
	fmt.Println(":whale: Container for image " + sourceImage + " (id " + idcontainer + ") created.")
	defer deleteContainer(idcontainer)

	var sexp, simp *specs.SpecFile = nil, nil

	ctx := context.TODO()

	if specExporter != "" {

		sexp, err = specs.NewSpecFileFromFile(specExporter)
		if err != nil {
			return fmt.Errorf("Error on read file %s: %s", specExporter, err.Error())
		}

	} else {

		// Prepare specfile for exporter reader
		sexp := specs.NewSpecFile()
		sexp.IgnoreFiles = []string{
			"/.dockerenv",
		}
		sexp.MapEntities = false
		sexp.SameChtimes = false
		sexp.SameOwner = true
		sexp.BrokenLinksFatal = true

	}

	if summary {
		sexp.Summary = summary
	}

	if specImporter != "" {

		simp, err = specs.NewSpecFileFromFile(specImporter)
		if err != nil {
			return fmt.Errorf("Error on read file %s: %s", specImporter, err.Error())
		}

	} else {
		// Prepare specfile for importer writer
		simp := specs.NewSpecFile()
		simp.MapEntities = false
		simp.SameChtimes = false
		simp.SameOwner = true
		simp.BrokenLinksFatal = true
		simp.Writer = specs.NewWriter()
	}

	// Prepare docker export command
	exportCmd := exec.CommandContext(ctx,
		"docker", "export", idcontainer)

	importArgs := []string{
		"import", "-", args.ImageTag,
	}

	if args.Message != "" {
		importArgs = append(importArgs, []string{
			"--message", args.Message,
		}...)
	}

	if args.Platform != "" {
		importArgs = append(importArgs, []string{
			"--platform", args.Platform,
		}...)
	}

	if len(args.Changes) > 0 {
		for _, c := range args.Changes {
			importArgs = append(importArgs, []string{
				"-c", c,
			}...)
		}
	}

	// Prepare docker import command
	importCmd := exec.CommandContext(ctx,
		"docker", importArgs...)

	// TODO: Add --platform, --message, --change

	exportReader, err := exportCmd.StdoutPipe()
	if err != nil {
		return err
	}

	importWriter, err := importCmd.StdinPipe()
	if err != nil {
		return err
	}

	tarformers.SetReader(exportReader)
	tarformers.SetWriter(importWriter)

	if err = importCmd.Start(); err != nil {
		return fmt.Errorf("docker import: %w", err)
	}

	if err = exportCmd.Start(); err != nil {
		_ = importWriter.Close()
		_ = importCmd.Process.Kill()
		_ = importCmd.Wait()

		return fmt.Errorf("docker export: %w", err)
	}

	err = tarformers.RunTaskBridge(sexp, simp)
	if err != nil {
		_ = importWriter.Close()

		_ = exportCmd.Process.Kill()
		_ = importCmd.Process.Kill()

		_ = exportCmd.Wait()
		_ = importCmd.Wait()

		return fmt.Errorf("tar bridge: %w", err)
	}

	if err = importWriter.Close(); err != nil {
		_ = exportCmd.Process.Kill()
		_ = importCmd.Process.Kill()

		_ = exportCmd.Wait()
		_ = importCmd.Wait()

		return fmt.Errorf("close docker import stdin: %w", err)
	}

	if err = exportCmd.Wait(); err != nil {
		_ = importCmd.Process.Kill()
		_ = importCmd.Wait()

		return fmt.Errorf("docker export: %w", err)
	}

	if err = importCmd.Wait(); err != nil {
		return fmt.Errorf("docker import: %w", err)
	}

	fmt.Println(fmt.Sprintf(":spouting_whale: Flatten image %s generated.", args.ImageTag))

	return nil
}

func newDockerFlatCommand(config *specs.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "docker-flat [source-image] [dest-image]",
		Short: "Create a flatten docker image from a source image.",
		Long: `Create a flat image from an existing image:

$> tar-formers docker-flat <source-image-tag> <flatten-image-tag>

`,
		Aliases: []string{"df", "flat"},
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				fmt.Println("Required arguments missed")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			summary, _ := cmd.Flags().GetBool("summary")
			specExporter, _ := cmd.Flags().GetString("specs-exporter")
			specImporter, _ := cmd.Flags().GetString("specs-importer")
			message, _ := cmd.Flags().GetString("message")
			platform, _ := cmd.Flags().GetString("platform")
			changes, _ := cmd.Flags().GetStringArray("change")

			// Check instance
			tarformers := executor.NewTarFormers(config)

			diargs := &DockerImportArgs{
				Message:  message,
				Platform: platform,
				Changes:  changes,
				ImageTag: args[1],
			}

			err := flatDockerContainer(
				tarformers, args[0], diargs,
				specExporter, specImporter, summary)

			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if summary {
				sum, _ := tarformers.GetSummary().YAML()
				fmt.Println(string(sum))
			}

		},
	}

	flags := cmd.Flags()
	flags.String("specs-exporter", "",
		"Define a spec file with the rules to follow for exporter flow.")
	flags.String("specs-importer", "",
		"Define a spec file with the rules to follow for the importer flow.")
	flags.Bool("summary", false, "Generate summary of the elaboration to stdout.")
	flags.StringP("message", "m", "",
		"Set commit message for imported image")
	flags.String("platform", "",
		"Set platform if server is multi-platform capable")
	flags.StringArray("change", []string{},
		"Apply Dockerfile instruction to the created image.")

	return cmd
}
