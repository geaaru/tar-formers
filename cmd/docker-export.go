/*

Copyright (C) 2021  Daniele Rondina <geaaru@sabayonlinux.org>

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

	"github.com/spf13/cobra"
)

func exporDockerContainer(tarformers *executor.TarFormers, cid, dir string) {
	cmds := []string{
		"/bin/bash", "-c",
		"docker export " + cid,
	}

	hostCommand := exec.Command(cmds[0], cmds[1:]...)
	hostCommand.Stderr = os.Stderr

	outReader, err := hostCommand.StdoutPipe()
	if err != nil {
		fmt.Println("Error on get stdout pipe :" + err.Error())
		os.Exit(1)
	}

	tarformers.SetReader(outReader)

	err = hostCommand.Start()
	if err != nil {
		fmt.Println("Error on start " + err.Error())
		os.Exit(1)
	}

	err = tarformers.RunTask(
		&specs.SpecFile{},
		dir,
	)
	if err != nil {
		fmt.Println("Error on process tarball :" + err.Error())
		os.Exit(1)
	}

	err = hostCommand.Wait()
	if err != nil {
		fmt.Println("Error on wait " + err.Error())
		os.Exit(1)
	}

	res := hostCommand.ProcessState.ExitCode()
	fmt.Println("Exiting with ", res)

	os.Exit(0)
}

func newDockerExportCommand(config *specs.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "docker-export [container-id]",
		Short:   "Export a docker container files to a specified directory.",
		Aliases: []string{"h"},
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("No container id present.")
				os.Exit(1)
			}

			to, _ := cmd.Flags().GetString("to")
			if to == "" {
				fmt.Println("No export directory defined.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			to, _ := cmd.Flags().GetString("to")

			// Check instance
			tarformers := executor.NewTarFormers(config)

			exporDockerContainer(tarformers, args[0], to)

		},
	}

	flags := cmd.Flags()
	flags.String("to", "", "Export directory")

	return cmd
}
