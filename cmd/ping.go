package main

import (
	"log"

	"github.com/NHollmann/ha-gateway/hagateway"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runPing(cmd *cobra.Command, args []string) {
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("fatal error config file: %v", err)
	}

	var conf config
	err = viper.Unmarshal(&conf)
	if err != nil {
		log.Fatalf("unable to decode into struct: %v", err)
	}

	gateway := hagateway.New(conf.HAURL, conf.HAToken, hagateway.Clients{})
	if err = gateway.Ping(); err != nil {
		log.Fatalf("ping error: %v", err)
	}

	log.Println("Connection OK")
}
