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
package cmd

import (
	"fmt"
	"os"
	"strings"

	specs "github.com/geaaru/tar-formers/pkg/specs"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cliName = `Copyright (c) 2021 - Daniele Rondina

Tar-formers - A golang tool to control tar flows
`

	TARFORMERS_VERSION = `0.4.0`
)

// Build time and commit information. This code is get from: https://github.com/mudler/luet/
//
// ⚠️ WARNING: should only be set by "-ldflags".
var (
	BuildTime   string
	BuildCommit string
)

func initConfig(config *specs.Config) {
	// Set env variable
	config.Viper.SetEnvPrefix(specs.TARFORMERS_ENV_PREFIX)
	config.Viper.BindEnv("config")
	config.Viper.SetDefault("config", "")

	config.Viper.AutomaticEnv()

	// Create EnvKey Replacer for handle complex structure
	replacer := strings.NewReplacer(".", "__")
	config.Viper.SetEnvKeyReplacer(replacer)

	// Set config file name (without extension)
	config.Viper.SetConfigName(specs.TARFORMERS_CONFIGNAME)

	config.Viper.SetTypeByDefaultValue(true)

}

func initCommand(rootCmd *cobra.Command, config *specs.Config) {
	var pflags = rootCmd.PersistentFlags()

	pflags.StringP("config", "c", "", "Tarformers configuration file")
	pflags.BoolP("debug", "d", config.Viper.GetBool("general.debug"),
		"Enable debug output.")

	config.Viper.BindPFlag("config", pflags.Lookup("config"))
	config.Viper.BindPFlag("general.debug", pflags.Lookup("debug"))

	rootCmd.AddCommand(
		newDockerExportCommand(config),
		newPortalCommand(config),
	)
}

func Execute() {
	// Create Main Instance Config object
	var config *specs.Config = specs.NewConfig(nil)

	initConfig(config)

	var rootCmd = &cobra.Command{
		Short:        cliName,
		Version:      fmt.Sprintf("%s-g%s %s", TARFORMERS_VERSION, BuildCommit, BuildTime),
		Args:         cobra.OnlyValidArgs,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			var v *viper.Viper = config.Viper

			v.SetConfigType("yml")
			if v.Get("config") == "" {
				config.Viper.AddConfigPath(".")
			} else {
				v.SetConfigFile(v.Get("config").(string))
			}

			// Parse configuration file
			err = config.Unmarshal()
			if err != nil {
				panic(err)
			}
		},
	}

	initCommand(rootCmd, config)

	// Start command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
