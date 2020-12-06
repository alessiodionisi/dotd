package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "v0.0.0-development"
	commit  = "development"
)

var (
	rootCmd = &cobra.Command{
		Use:     "dotd",
		Long:    "DotD is a Simple and flexible DNS over HTTPS proxy with custom resolver and other perks.",
		Version: fmt.Sprintf("%s (%s)", version, commit),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			parsedLogLevel, err := zerolog.ParseLevel(viper.GetString("logLevel"))
			if err != nil {
				return err
			}

			zerolog.SetGlobalLevel(parsedLogLevel)

			return nil
		},
	}
	configFile string
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file to use")
	rootCmd.PersistentFlags().StringP("loglevel", "l", "info", `log level, valid values are "trace", "debug", "info", "warn", "error", "fatal" or "panic"`)

	_ = viper.BindPFlags(rootCmd.PersistentFlags())

	rootCmd.AddCommand(serverCmd)
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath("$HOME/.config/dotd")
		viper.SetConfigName("dotd")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("dotd")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal().Msg(err.Error())
		}
	}

	if viper.ConfigFileUsed() != "" {
		log.Info().Msgf(`using config file "%s"`, viper.ConfigFileUsed())
	}
}

func Execute() {
	_ = rootCmd.Execute()
}
