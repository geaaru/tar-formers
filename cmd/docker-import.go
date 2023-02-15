/*

Copyright (C) 2021-2023  Daniele Rondina <geaaru@sabayonlinux.org>

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
	"archive/tar"
	"fmt"
	"os"
	"os/exec"
	"strings"

	executor "github.com/geaaru/tar-formers/pkg/executor"
	specs "github.com/geaaru/tar-formers/pkg/specs"

	"github.com/spf13/cobra"
)

type DockerImportArgs struct {
	Message  string
	Platform string
	Changes  []string
	ImageTag string
}

func importDockerContainer(tarformers *executor.TarFormers,
	args *DockerImportArgs, dir, file, spec string) error {
	var s *specs.SpecFile = nil
	var sReader *specs.SpecFile = nil
	var err error

	cmds := []string{
		"docker", "import",
		"-", args.ImageTag,
	}

	if args.Message != "" {
		cmds = append(cmds, []string{
			"--message", args.Message,
		}...)
	}

	if args.Platform != "" {
		cmds = append(cmds, []string{
			"--platform", args.Platform,
		}...)
	}

	if len(args.Changes) > 0 {
		for _, c := range args.Changes {
			cmds = append(cmds, []string{
				"-c", c,
			}...)
		}
	}

	if spec != "" {
		s, err = specs.NewSpecFileFromFile(spec)
		if err != nil {
			return fmt.Errorf("Error on read file %s: %s",
				spec, err.Error())
		}
	} else {
		s = specs.NewSpecFile()
		s.SameChtimes = true
		s.Writer = specs.NewWriter()
		if dir != "" {
			s.Writer.ArchiveDirs = []string{dir}

			hf := func(path, newpath string,
				header *tar.Header, tw *tar.Writer,
				opts *executor.TarFileOperation,
				t *executor.TarFormers) error {

				if dir != "./" {
					// Nothing to do if the path is equal to ./
					newpath = strings.Replace(newpath, dir, "./", 1)
					opts.Rename = true
					opts.NewName = newpath
				}

				return nil
			}

			tarformers.SetFileWriterHandler(hf)

		} else {
			// Open file to read
			f, err := os.OpenFile(file, os.O_RDONLY, 0666)
			if err != nil {
				return fmt.Errorf("Error on open file %s: %s",
					file, err.Error())
			}
			defer f.Close()
			tarformers.SetReader(f)

			sReader = specs.NewSpecFile()
			sReader.IgnoreFiles = append(sReader.IgnoreFiles, ".dockerenv")
		}
	}

	hostCommand := exec.Command(cmds[0], cmds[1:]...)
	hostCommand.Stderr = os.Stderr
	hostCommand.Stdout = os.Stdout

	inReader, err := hostCommand.StdinPipe()
	if err != nil {
		return fmt.Errorf("Error on get stdin pipe :" + err.Error())
	}

	tarformers.SetWriter(inReader)

	err = hostCommand.Start()
	if err != nil {
		return fmt.Errorf("Error on start " + err.Error())
	}

	if file != "" {
		err = tarformers.RunTaskBridge(sReader, s)
		if err != nil {
			return fmt.Errorf("Error on process tarball :" + err.Error())
		}
	} else {
		err = tarformers.RunTaskWriter(s)
		if err != nil {
			return fmt.Errorf("Error on process tarball :" + err.Error())
		}
	}

	// NOTE: The reader must be closed before the wait
	//       to avoid starvation.
	inReader.Close()

	err = hostCommand.Wait()
	if err != nil {
		return fmt.Errorf("Error on wait " + err.Error())
	}

	res := hostCommand.ProcessState.ExitCode()
	if res != 0 {
		fmt.Println("Importing exit with ", res)
	} else {
		fmt.Println("Operation completed.")
	}

	return nil
}

func newDockerImportCommand(config *specs.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "docker-import <image-tag> [OPTIONS]",
		Short: "Create a docker image from a directory or a tarball.",
		Long: `Create a docker image from a directory:

$> tar-formers di geaaru/tar-formers:latest --dir ./tmp --platform amd64 -m "My image" \
   --change 'ENTRYPOINT ["/bin/sh"]'

Create a docker image from a tarball file:

$> tar-formers di geaaru/tar-formers:latest --file ./alpine.tar --platform amd64 -m "My image" \
   --change 'ENTRYPOINT ["/bin/sh"]'

`,

		Aliases: []string{"di"},
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("No image tag present.")
				os.Exit(1)
			}

			dir, _ := cmd.Flags().GetString("dir")
			file, _ := cmd.Flags().GetString("file")
			if file == "" && dir == "" {
				fmt.Println("You need to use --file or --dir")
				os.Exit(1)
			} else if file != "" && dir != "" {
				fmt.Println("The options --file and --dir could not be used together.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			tarformers := executor.NewTarFormers(config)

			specs, _ := cmd.Flags().GetString("specs")
			dir, _ := cmd.Flags().GetString("dir")
			file, _ := cmd.Flags().GetString("file")
			message, _ := cmd.Flags().GetString("message")
			platform, _ := cmd.Flags().GetString("platform")
			changes, _ := cmd.Flags().GetStringArray("change")

			diargs := &DockerImportArgs{
				Message:  message,
				Platform: platform,
				Changes:  changes,
				ImageTag: args[0],
			}

			err := importDockerContainer(tarformers, diargs, dir, file, specs)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		},
	}

	flags := cmd.Flags()
	flags.String("dir", "", "Define directory to import in the image as /.")
	flags.String("file", "", "Read tar flow from specified file (only .tar is supported).")
	flags.StringP("message", "m", "",
		"Set commit message for imported image")
	flags.String("platform", "",
		"Set platform if server is multi-platform capable")
	flags.StringArray("change", []string{},
		"Apply Dockerfile instruction to the created image.")
	flags.String("specs", "", "Define a spec file with the rules to follow.")

	return cmd
}
