/*

Copyright (C) 2021  Daniele Rondina <geaaru@sabayonlinux.org>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/
package executor

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func (t *TarFormers) CreateFile(file string, mode os.FileMode, reader *tar.Reader, header *tar.Header) error {

	fmt.Println("Creating file", file, "...")

	err := t.CreateDir(filepath.Dir(file), mode|os.ModeDir|100)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return errors.New(
			fmt.Sprintf("Error on open file %s: %s", file, err.Error()))
	}
	defer f.Close()

	// Copy file content
	nb, err := io.Copy(f, reader)
	if err != nil {
		return errors.New(
			fmt.Sprintf("Error on write file %s: %s", file, err.Error()))
	}
	if nb != header.Size {
		return errors.New(
			fmt.Sprintf("For file %s written file are different %d - %d",
				file, nb, header.Size))
	}

	fmt.Println("Written ", nb)

	// TODO: check if it's needed f.Sync()
	// 	if err := f.Sync(); err != nil {
	//	return err
	//}

	return nil
}

func (t *TarFormers) CreateBlockCharFifo(file string, mode os.FileMode, header *tar.Header) error {
	err := t.CreateDir(filepath.Dir(file), mode|os.ModeDir|100)
	if err != nil {
		return err
	}

	modeDev := uint32(header.Mode & 07777)
	switch header.Typeflag {
	case tar.TypeBlock:
		modeDev |= unix.S_IFBLK
	case tar.TypeChar:
		modeDev |= unix.S_IFCHR
	case tar.TypeFifo:
		modeDev |= unix.S_IFIFO
	}

	dev := int(uint32(unix.Mkdev(uint32(header.Devmajor), uint32(header.Devminor))))
	return unix.Mknod(file, modeDev, dev)
}
