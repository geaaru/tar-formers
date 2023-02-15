/*
Copyright (C) 2021-2022  Daniele Rondina <geaaru@funtoo.org>

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

	log "github.com/geaaru/tar-formers/pkg/logger"
	specs "github.com/geaaru/tar-formers/pkg/specs"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cliName = `Copyright (c) 2021-2023 - Daniele Rondina

Tar-formers - A golang tool to control tar flows
`
	TARFORMERS_VERSION = `0.7.0`
)

var (
	BuildTime      string
	BuildCommit    string
	BuildGoVersion string
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

func version() string {
	ans := fmt.Sprintf("%s-g%s %s", TARFORMERS_VERSION, BuildCommit, BuildTime)
	if BuildGoVersion != "" {
		ans += " " + BuildGoVersion
	}
	return ans
}

func initCommand(rootCmd *cobra.Command, config *specs.Config) {
	var pflags = rootCmd.PersistentFlags()

	pflags.StringP("config", "c", "", "Tarformers configuration file")
	pflags.BoolP("debug", "d", config.Viper.GetBool("general.debug"),
		"Enable debug output.")

	config.Viper.BindPFlag("config", pflags.Lookup("config"))
	config.Viper.BindPFlag("general.debug", pflags.Lookup("debug"))

	rootCmd.AddCommand(
		newBridgeCommand(config),
		newDockerExportCommand(config),
		newDockerImportCommand(config),
		newPortalCommand(config),
		newArchiveCommand(config),
	)
}

func Execute() {
	// Create Main Instance Config object
	var config *specs.Config = specs.NewConfig(nil)

	initConfig(config)

	var rootCmd = &cobra.Command{
		Short:        cliName,
		Version:      version(),
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

			logger := log.NewLogger(config)
			logger.SetAsDefault()
		},
	}

	initCommand(rootCmd, config)

	// Start command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
