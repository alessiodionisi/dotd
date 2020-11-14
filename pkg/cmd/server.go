package cmd

import (
	"github.com/adnsio/dotd/pkg/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverCmd = &cobra.Command{
	Use: "server",
	Aliases: []string{
		"serve",
	},
	Short: "Starts the DNS server",
	Long:  "Starts DotD DNS server, set it as primary resolver to start using it.",
	Run:   runServer,
}

func init() {
	serverCmd.Flags().StringP("address", "a", "[::1]:53", "listening address")
	serverCmd.Flags().StringSliceP("upstreams", "u", []string{"https://1.1.1.1/dns-query", "https://1.0.0.1/dns-query"}, "upstream addresses")
	serverCmd.Flags().StringSlice("blocklist", []string{}, "blocked domains")
	serverCmd.Flags().StringToString("resolve", map[string]string{}, "custom resolve list")

	_ = viper.BindPFlag("address", serverCmd.Flags().Lookup("address"))
	_ = viper.BindPFlag("upstreams", serverCmd.Flags().Lookup("upstreams"))
	_ = viper.BindPFlag("blocklist", serverCmd.Flags().Lookup("blocklist"))
	_ = viper.BindPFlag("resolve", serverCmd.Flags().Lookup("resolve"))
}

func runServer(cmd *cobra.Command, args []string) {
	address := viper.GetString("address")
	upstreams := viper.GetStringSlice("upstreams")
	resolve := viper.GetStringMapString("resolve")

	server, err := server.New(address, upstreams, resolve)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Send()
	}
}
