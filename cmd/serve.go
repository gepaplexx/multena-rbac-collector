/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/gepaplexx/multena-rbac-collector/server"
	"github.com/gepaplexx/multena-rbac-collector/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/cobra"
)

var (
	level int
	port  int
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts continues RBAC collection",
	Long:  `Starts continues RBAC collection and updates the ConfigMap with the RBAC data.`,
	Run: func(cmd *cobra.Command, args []string) {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		log.Info().Int("port", port).Msg("")
		logCommit()
		initializeKubernetesClient()
		log.Info().Msg("Starting RBAC analyzer server...")
		server.Serve(clientset, port, util.Config{
			CMName:      cmName,
			CMNamespace: cmNamespace,
		})
	},
}

func init() {
	cobra.OnInitialize(updateLogLevel)
	rootCmd.AddCommand(serveCmd)
	serveCmd.PersistentFlags().IntVarP(&level, "level", "l", 1, "Set log level between 0 and 5")
	serveCmd.PersistentFlags().IntVarP(&port, "port", "p", 8080, "Set port to listen on")
}

func updateLogLevel() {
	zerolog.SetGlobalLevel(zerolog.Level(level))
}
