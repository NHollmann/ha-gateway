package main

import (
	"fmt"
	"log"

	"github.com/NHollmann/ha-gateway/hagateway"
	"github.com/spf13/cobra"
)

func runCreateKey(cmd *cobra.Command, args []string) {
	log.Println("Generate new API key and hash...")

	apiKey, hash, err := hagateway.ClientGenerateKey()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println()
	fmt.Printf("API KEY: %s\n", apiKey)
	fmt.Printf("HASH:    %s\n", hash)
	fmt.Println()

	log.Println("Done")
}
