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

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", `config file (default "config.yaml|json|toml|hcl|ini")`)
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", `log level, valid values are "trace", "debug", "info", "warn", "error", "fatal" or "panic"`)

	_ = viper.BindPFlag("logLevel", rootCmd.PersistentFlags().Lookup("log-level"))

	rootCmd.AddCommand(serverCmd)
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		workingDir, err := os.Getwd()
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		viper.AddConfigPath(workingDir)
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Info().Msgf("using config file %s", viper.ConfigFileUsed())
	}
}

func Execute() {
	_ = rootCmd.Execute()
}
