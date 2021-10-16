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
package specs

import (
	"os"
)

type SpecFile struct {
	File string `yaml:"-" json:"-"`

	Ignore []string `yaml:"ignore,omitempty" json:"ignore,omitempty"`

	Rename []RenameRule `yaml:"rename,omitempty" json:"rename,omitempty"`
}

type RenameRule struct {
	Source string `yaml:"source" json:"source"`
	Dest   string `yaml:"dest" json:"dest"`
}

type Link struct {
	Name     string
	Path     string
	Mode     os.FileMode
	Symbolic bool
}
