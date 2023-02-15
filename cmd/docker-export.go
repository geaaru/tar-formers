/*

Copyright (C) 2021-2023 Daniele Rondina <geaaru@gmail.com>

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
	"fmt"
	"os"
	"os/exec"

	executor "github.com/geaaru/tar-formers/pkg/executor"
	specs "github.com/geaaru/tar-formers/pkg/specs"
	"github.com/geaaru/tar-formers/pkg/tools"

	"github.com/spf13/cobra"
)

func exporDockerContainer(tarformers *executor.TarFormers,
	cid, dir, file, spec, specOut string) error {
	var s *specs.SpecFile = nil
	var sWriter *specs.SpecFile = nil
	var err error

	cmds := []string{
		"/bin/bash", "-c",
		"docker export " + cid,
	}

	if spec != "" {
		s, err = specs.NewSpecFileFromFile(spec)
		if err != nil {
			return fmt.Errorf(
				"Error on read file %s: %s",
				spec, err.Error())
		}
	} else {
		s = specs.NewSpecFile()
		s.IgnoreFiles = append(s.IgnoreFiles, ".dockerenv")
	}

	if specOut != "" {
		sWriter, err = specs.NewSpecFileFromFile(specOut)
		if err != nil {
			return fmt.Errorf(
				"Error on read file %s: %s",
				specOut, err.Error())
		}
	} else {
		sWriter = specs.NewSpecFile()
		sWriter.SameChtimes = true
		sWriter.Writer = specs.NewWriter()
	}

	hostCommand := exec.Command(cmds[0], cmds[1:]...)
	hostCommand.Stderr = os.Stderr

	outReader, err := hostCommand.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Error on get stdout pipe :" + err.Error())
	}

	tarformers.SetReader(outReader)

	err = hostCommand.Start()
	if err != nil {
		return fmt.Errorf("Error on start " + err.Error())
	}

	if dir != "" {
		err = tarformers.RunTask(s, dir)
		if err != nil {
			return fmt.Errorf("Error on process tarball :" + err.Error())
		}
	} else {
		// Prepare the writer
		opts := tools.NewTarCompressionOpts(true)
		defer opts.Close()

		err = tools.PrepareTarWriter(file, opts)
		if err != nil {
			return fmt.Errorf("Error on prepare writer: %s",
				err.Error())
		}

		if opts.CompressWriter != nil {
			tarformers.SetWriter(opts.CompressWriter)
		} else {
			tarformers.SetWriter(opts.FileWriter)
		}

		err = tarformers.RunTaskBridge(s, sWriter)
		if err != nil {
			return fmt.Errorf("Error on process tarball :" + err.Error())
		}

	}

	err = hostCommand.Wait()
	if err != nil {
		return fmt.Errorf("Error on wait " + err.Error())
	}

	res := hostCommand.ProcessState.ExitCode()
	if res != 0 {
		fmt.Println("Exporting exit with ", res)
	} else if file != "-" {
		fmt.Println("Operation completed.")
	}

	return nil
}

func newDockerExportCommand(config *specs.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "docker-export [container-id]",
		Short: "Export the files a docker container to a specified directory or to a file.",
		Long: `Export docker container files to a specified directory:

$> tar-formers docker-export <container-id> --todir ./out

Export docker container files, apply filter and generate a new tarball.

$> tar-formers docker-export <container-id> --to /mycontainer.tar.gz --specs spec.yml

Export docker container files, apply filter on both reader and writer
and generate a new tarball.

$> tar-formers docker-export <container-id> --to /mycontainer.tar.gz --specs spec.yml \
   --out specs-writer.yml

`,
		Aliases: []string{"de"},
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("No container id present.")
				os.Exit(1)
			}

			todir, _ := cmd.Flags().GetString("todir")
			to, _ := cmd.Flags().GetString("to")
			if todir == "" && to == "" {
				fmt.Println(
					"No export directory or target file defined.",
				)
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			to, _ := cmd.Flags().GetString("to")
			todir, _ := cmd.Flags().GetString("todir")
			specfile, _ := cmd.Flags().GetString("specs")
			out, _ := cmd.Flags().GetString("out")

			// Check instance
			tarformers := executor.NewTarFormers(config)

			err := exporDockerContainer(
				tarformers, args[0], todir,
				to, specfile, out)

			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

		},
	}

	flags := cmd.Flags()
	flags.String("todir", "", "Export directory where untar files.")
	flags.String("to", "", "Target tarball file or stream with the container files.")
	flags.String("specs", "", "Define a spec file with the rules to follow.")
	flags.String("out", "",
		"Define a spec file with the rules to follow for the writer. Only used with --to.")

	return cmd
}
