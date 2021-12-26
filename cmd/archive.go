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
	"archive/tar"
	"fmt"
	"os"

	executor "github.com/geaaru/tar-formers/pkg/executor"
	specs "github.com/geaaru/tar-formers/pkg/specs"

	"github.com/spf13/cobra"
)

func newArchiveCommand(config *specs.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "archive <tarball> <dir1> ... <dirN> [OPTIONS]",
		Short:   "Archive one or more directories to a tarball.",
		Aliases: []string{"h"},
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				fmt.Println("Missing mandatory arguments")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var s *specs.SpecFile = nil
			var err error

			spec, _ := cmd.Flags().GetString("specs")

			// Check instance
			tarformers := executor.NewTarFormers(config)

			archiveFile := args[0]
			dirs := args[1:]
			if spec != "" {
				s, err = specs.NewSpecFileFromFile(spec)
				if err != nil {
					fmt.Println(fmt.Sprintf(
						"Error on read file %s: %s",
						spec, err.Error()))
					os.Exit(1)
				}
			} else {
				s = specs.NewSpecFile()
				s.SameChtimes = true
			}

			tarformers.TaskWriter = s

			// Create the tarball
			out, err := os.Create(archiveFile)
			if err != nil {
				fmt.Println(fmt.Sprintf(
					"Error on create file %s: %s", archiveFile, err.Error()))
				os.Exit(1)
			}
			defer out.Close()

			tw := tar.NewWriter(out)
			defer tw.Close()

			for _, d := range dirs {
				err := tarformers.InjectDir2Writer(tw, d)
				if err != nil {
					fmt.Println(
						fmt.Sprintf("Error on inject directory %s: %s",
							d, err.Error()))
					os.Exit(1)
				}
			}

			fmt.Println("Operation completed.")
		},
	}

	flags := cmd.Flags()
	flags.String("specs", "", "Define a spec file with the rules to follow.")

	return cmd
}
