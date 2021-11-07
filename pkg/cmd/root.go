package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//nolint:gochecknoglobals
var (
	version = "v0.0.0-development"
	commit  = "development"
	rootCmd = &cobra.Command{
		Use:     "dotd",
		Long:    "DotD is a Simple and flexible DNS over HTTPS proxy with custom resolver and other perks.",
		Version: fmt.Sprintf("%s (%s)", version, commit),
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			parsedLogLevel, err := zerolog.ParseLevel(viper.GetString("log-level"))
			if err != nil {
				log.Fatal().Err(fmt.Errorf("zerolog: %w", err)).Send()
			}

			zerolog.SetGlobalLevel(parsedLogLevel)
		},
	}
)

//nolint:gochecknoinits
func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file to use")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", `log level, valid values are "trace", "debug", "info", "warn", "error", "fatal" or "panic"`)

	rootCmd.AddCommand(serverCmd)

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.Fatal().Err(fmt.Errorf("viper: %w", err)).Send()
	}

	viper.RegisterAlias("loglevel", "log-level")
}

func initConfig() {
	configFile := viper.GetString("config")

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath("$HOME/.config/dotd")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("dotd")

	if err := viper.ReadInConfig(); err != nil {
		//nolint:errorlint
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal().Err(fmt.Errorf("viper: %w", err)).Send()
		}
	}

	if viper.ConfigFileUsed() != "" {
		log.Info().Msgf(`using config file "%s"`, viper.ConfigFileUsed())
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
