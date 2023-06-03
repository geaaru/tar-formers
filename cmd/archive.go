/*

Copyright (C) 2021-2023  Daniele Rondina <geaaru@gmail.com>

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

func newArchiveCommand(config *specs.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "archive <tarball|-> [<dir1> ... <dirN>] [OPTIONS]",
		Short: "Archive one or more directories to a tarball.",
		Long: `Archive one or more directly without changes on tarball file file.tar.gz

$> tar-formers archive /tmp/file.tar.gz /mydir1 /mydir2

Archive directories defined on the spec file to file file.tar.xz

$> tar-formers archive /tmp/file.tar.xz --specs specs.yaml

Archive directories defined on the spec file with eventually filters
to stdout as tar stream:

$> tar-formers archive - --specs specs.yaml | tar -C /target -xvf

Archive directories and filters file to stdout as compressed stream:

$> tar-formers archive - --specs specs.yaml --compression zstd > /tmp/file.tar.zstd

NOTE: Bzip2 compression is experimental.
`,
		Aliases: []string{"a"},
		PreRun: func(cmd *cobra.Command, args []string) {
			spec, _ := cmd.Flags().GetString("specs")
			if len(args) < 1 || (len(args) < 2 && spec == "") {
				fmt.Println("Missing mandatory arguments")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var s *specs.SpecFile = nil
			var err error

			spec, _ := cmd.Flags().GetString("specs")
			compression, _ := cmd.Flags().GetString("compression")

			// Check instance
			tarformers := executor.NewTarFormers(config)

			archiveFile := args[0]
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
				s.Writer = specs.NewWriter()
				s.Writer.ArchiveDirs = args[1:]
			}

			opts := tools.NewTarCompressionOpts(compression == "")
			if compression != "" {
				opts.Mode = tools.ParseCompressionMode(compression)
			}
			defer opts.Close()

			err = tools.PrepareTarWriter(archiveFile, opts)
			if err != nil {
				fmt.Println(fmt.Sprintf(
					"Error on prepare writer: %s",
					err.Error()))
				os.Exit(1)
			}

			if opts.CompressWriter != nil {
				tarformers.SetWriter(opts.CompressWriter)
			} else {
				tarformers.SetWriter(opts.FileWriter)
			}

			err = tarformers.RunTaskWriter(s)
			if err != nil {
				fmt.Println(fmt.Sprintf(
					"Error on create tarball %s: %s",
					archiveFile, err.Error()))
				opts.Close()
				os.Exit(1)
			}

			if archiveFile != "-" {
				fmt.Println("Operation completed.")
			}
		},
	}

	flags := cmd.Flags()
	flags.String("compression", "",
		"Specify tarball compression and ignoring extension of the file."+
			" Possible values: gz|gzip|zstd|xz|bz2|bzip2|none.")
	flags.String("specs", "", "Define a spec file with the rules to follow.")

	return cmd
}
