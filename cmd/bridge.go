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
	"bufio"
	"fmt"
	"io"
	"os"

	executor "github.com/geaaru/tar-formers/pkg/executor"
	specs "github.com/geaaru/tar-formers/pkg/specs"
	"github.com/geaaru/tar-formers/pkg/tools"

	"github.com/spf13/cobra"
)

func newBridgeCommand(config *specs.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "bridge [OPTIONS]",
		Short: "Extract a stdin flow or an input tarball and bridge it to tar output stream or file.",
		Long: `Manage input and output tar streams together.

Apply a filter to a tar stream and write the stream to a file
compressed:

$> tar -cpf - /mydir | tar-formers bridge --stdin --to /tmp/file.tar.gz --out spec.yaml

Apply a filter to a tar stream and write the stream to a file
compressed with both input and output filters:

$> tar -cpf - /mydir | tar-formers bridge --stdin --to /tmp/file.tar.xz --out spec.yaml --in spec-reader.yaml

Apply a filter an input tarball file and write the stream to a file
compressed with both input and output filters:

$> tar-formers bridge --stdin --file /input.tar --to /tmp/file.tar.xz --out spec.yaml --in spec-reader.yaml

`,
		Aliases: []string{"b"},
		PreRun: func(cmd *cobra.Command, args []string) {
			to, _ := cmd.Flags().GetString("to")
			if to == "" {
				fmt.Println("No target file or pipe defined.")
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
			var sReader *specs.SpecFile = nil
			var sWriter *specs.SpecFile = nil
			var err error

			to, _ := cmd.Flags().GetString("to")
			specIn, _ := cmd.Flags().GetString("in")
			specOut, _ := cmd.Flags().GetString("out")
			stdin, _ := cmd.Flags().GetBool("stdin")
			file, _ := cmd.Flags().GetString("file")
			compression, _ := cmd.Flags().GetString("compression")

			// Check instance
			tarformers := executor.NewTarFormers(config)

			// Parse input spec file
			if specIn != "" {
				sReader, err = specs.NewSpecFileFromFile(specIn)
				if err != nil {
					fmt.Println(fmt.Sprintf(
						"Error on read file %s: %s",
						specIn, err.Error()))
					os.Exit(1)
				}
			} else {
				sReader = specs.NewSpecFile()
				sReader.IgnoreFiles = append(sReader.IgnoreFiles, "/.dockerenv")
			}

			// Parse output spec file
			if specOut != "" {
				sWriter, err = specs.NewSpecFileFromFile(specOut)
				if err != nil {
					fmt.Println(fmt.Sprintf(
						"Error on read file %s: %s",
						specOut, err.Error()))
					os.Exit(1)
				}
			} else {
				sWriter = specs.NewSpecFile()
				sWriter.SameChtimes = true
				sWriter.Writer = specs.NewWriter()
			}

			// Prepare the writer
			opts := tools.NewTarCompressionOpts(compression == "")
			if compression != "" {
				opts.Mode = tools.ParseCompressionMode(compression)
			}
			defer opts.Close()

			err = tools.PrepareTarWriter(to, opts)
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

			// Prepare the reader
			var reader io.Reader

			if stdin {
				reader = bufio.NewReader(os.Stdin)
			} else {
				f, err := os.OpenFile(file, os.O_RDONLY, 0666)
				if err != nil {
					fmt.Println(fmt.Sprintf("Error on open file %s: %s",
						file, err.Error()))
					os.Exit(1)
				}
				defer f.Close()
				reader = f
			}
			tarformers.SetReader(reader)

			err = tarformers.RunTaskBridge(sReader, sWriter)
			if err != nil {
				fmt.Println("Error on process tarball :" + err.Error())
				os.Exit(1)
			}

			if to != "-" {
				fmt.Println("Operation completed.")
			}
		},
	}

	flags := cmd.Flags()
	flags.String("in", "", "Define a spec file with the rules to follow for the reader.")
	flags.String("out", "", "Define a spec file with the rules to follow for the writer.")
	flags.Bool("stdin", false, "Read tar flow from stdin.")
	flags.String("file", "", "Read tar flow from specified file.")
	flags.String("to", "", "File where write the tar flow. Use - for stdout.")
	flags.String("compression", "",
		"Specify tarball compression and ignoring extention of the file."+
			" Possible values: gz|gzip|zstd|xz|bz2|bzip2|none.")

	return cmd
}
