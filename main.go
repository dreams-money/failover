package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dreams-money/failover/config"
	"github.com/dreams-money/failover/health"
	"github.com/dreams-money/failover/routers"
)

func main() {
	// Load config
	configuration, err := config.LoadProgramConfiguration()
	if err != nil {
		log.Panic(err)
	}

	router, err := routers.Make(configuration)
	if err != nil {
		log.Panic(err)
	}

	// Set Router Auth
	router.SetAuthorization(configuration)
	err = router.SimpleCall(configuration)
	if err != nil {
		log.Println("Authorization failed", err)
		os.Exit(1)
		return
	}

	// Failover endpoint
	http.HandleFunc("/failover", func(w http.ResponseWriter, r *http.Request) {
		err := router.Failover(configuration)
		if err != nil {
			log.Println("Router Failover failed", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	// Health check endpoint
	http.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		err = router.SimpleCall(configuration)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	// Loop to check peers
	go startCheckPeerJob(configuration)

	log.Println("Server listening on http://localhost:" + configuration.AppPort)
	err = http.ListenAndServe(":"+configuration.AppPort, nil)
	if err != nil {
		log.Panicf("Error starting server: %s\n", err)
	}
}

func startCheckPeerJob(cfg config.Config) {
	heartBeatInterval := time.Tick(cfg.HeartBeatInterval)

	for range heartBeatInterval {
		checkPeers(cfg)
	}
}

func checkPeers(cfg config.Config) {
	var err error
	for peer, config := range cfg.Peers {
		if !config.CheckHealth {
			continue
		}

		err = health.CheckPeer(peer, config.Address+"/heartbeat")
		if err != nil {
			log.Println("Health check failed to execute!", err)
		}
	}
}
