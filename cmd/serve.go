package cmd

import (
	"FaRyuk/config"
	"FaRyuk/internal"

	"github.com/spf13/cobra"
)

var (
	wapiaddr string
	wapiport int
)

// rootCmd represents the base command when called without any subcommands
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve API",
	Run:   LaunchServe,
}

// LaunchServe : launch api and web server
func LaunchServe(cmd *cobra.Command, args []string) {
	config.Init()
	internal.MainServer()
}

func init() {
	rootCmd.AddCommand(serveCmd)
	// global
	serveCmd.PersistentFlags().StringVarP(&wapiaddr, "listen-api-addr", "L", config.Cfg.Server.Addr, "Listen address")
	serveCmd.PersistentFlags().IntVarP(&wapiport, "listen-api-port", "P", config.Cfg.Server.Port, "Listen port")
	err := cobra.MarkFlagRequired(serveCmd.Flags(), "serve")
	if err != nil {
		return
	}
}
