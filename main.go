package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/NHollmann/ha-gateway/hagateway"

	"github.com/spf13/viper"
)

type config struct {
	ListenAddr string `mapstructure:"listener_address"`
	HAURL      string `mapstructure:"ha_url"`
	HAToken    string `mapstructure:"ha_token"`
	Clients    []hagateway.Client
}

func main() {
	viper.SetConfigName("gateway-config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/ha-gateway")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("fatal error config file: %v", err)
	}

	var conf config
	err = viper.Unmarshal(&conf)
	if err != nil {
		log.Fatalf("unable to decode into struct: %v", err)
	}

	clients := hagateway.Clients{}
	for _, client := range conf.Clients {
		clients.Add(&client)
	}
	gateway := hagateway.New(conf.HAURL, conf.HAToken, clients)

	mux := http.NewServeMux()
	mux.Handle("/api/states/", gateway)

	runServer(conf.ListenAddr, mux)
}

func runServer(addr string, handler http.Handler) {
	log.Printf("Start server on %s...", addr)
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	log.Println("Graceful shutdown initiated...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Graceful shutdown fauled: %v", err)
	}

	log.Println("Server stopped.")
}
