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

	executor "github.com/geaaru/tar-formers/pkg/executor"
	specs "github.com/geaaru/tar-formers/pkg/specs"
	"github.com/geaaru/tar-formers/pkg/tools"

	"github.com/spf13/cobra"
)

func newPortalCommand(config *specs.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "portal [OPTIONS]",
		Short:   "Extract a stdin flow or a tar file to a specified directory.",
		Aliases: []string{"h"},
		PreRun: func(cmd *cobra.Command, args []string) {
			to, _ := cmd.Flags().GetString("to")
			if to == "" {
				fmt.Println("No export directory defined.")
				os.Exit(1)
			}

			stdin, _ := cmd.Flags().GetBool("stdin")
			file, _ := cmd.Flags().GetString("file")
			if stdin && file != "" {
				fmt.Println("You can use --stdin or --file. Not both.")
				os.Exit(1)
			}

			if !stdin && file == "" {
				fmt.Println("You need specied --file option or --stdin option.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var s *specs.SpecFile = nil
			var err error

			to, _ := cmd.Flags().GetString("to")
			spec, _ := cmd.Flags().GetString("specs")
			stdin, _ := cmd.Flags().GetBool("stdin")
			file, _ := cmd.Flags().GetString("file")
			compression, _ := cmd.Flags().GetString("compression")

			// Check instance
			tarformers := executor.NewTarFormers(config)

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
				s.IgnoreFiles = append(s.IgnoreFiles, "/.dockerenv")
			}

			opts := tools.NewTarReaderCompressionOpts(compression == "")
			if compression != "" {
				opts.Mode = tools.ParseCompressionMode(compression)
			}

			if stdin {
				file = "-"
			}
			err = tools.PrepareTarReader(file, opts)
			if err != nil {
				fmt.Println("Error on prepare reader:", err.Error())
				os.Exit(1)
			}

			if opts.CompressReader != nil {
				tarformers.SetReader(opts.CompressReader)
			} else {
				tarformers.SetReader(opts.FileReader)
			}

			err = tarformers.RunTask(s, to)
			opts.Close()
			if err != nil {
				fmt.Println("Error on process tarball :" + err.Error())
				os.Exit(1)
			}

			fmt.Println("Operation completed.")
		},
	}

	flags := cmd.Flags()
	flags.String("to", "", "Export directory where untar files.")
	flags.String("specs", "", "Define a spec file with the rules to follow.")
	flags.Bool("stdin", false, "Read tar flow from stdin.")
	flags.String("file", "", "Read tar flow from specified file.")
	flags.String("compression", "",
		"Specify tarball compression and ignoring extention of the file."+
			" Possible values: gz|gzip|zstd|xz|bz2|bzip2|none.")

	return cmd
}
