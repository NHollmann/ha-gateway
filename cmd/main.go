package main

import (
	"log"

	"github.com/NHollmann/ha-gateway/hagateway"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "ha-gateway",
	Short: "A HomeAssistant API Gateway",
	Long: `ha-gateway is small server that allows you to restrict access to the HomeAssistant API.

Instead of giving out your long living token to different servers, create custom API keys with limited
access to different entities.`,
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run server",
	Long:  "Run the API server",
	Run:   runServer,
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping HomeAssistant",
	Long:  "Try to ping HomeAssistant to check its accessible",
	Run:   runPing,
}

var createKeyCmd = &cobra.Command{
	Use:   "create-key",
	Short: "Create a new API Key",
	Long:  "Create a new random API Key and the associated hash",
	Run:   runCreateKey,
}

type config struct {
	ListenAddr string `mapstructure:"listener_address"`
	HAURL      string `mapstructure:"ha_url"`
	HAToken    string `mapstructure:"ha_token"`
	Clients    []hagateway.Client
}

func init() {
	rootCmd.AddCommand(serverCmd, pingCmd, createKeyCmd)
	setupConfig()
}

func setupConfig() {
	viper.SetConfigName("gateway-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/ha-gateway")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
