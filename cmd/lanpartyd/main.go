package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ffa/lan-party/internal/config"
	"github.com/ffa/lan-party/internal/headscale"
	"github.com/ffa/lan-party/internal/server"
	"github.com/ffa/lan-party/internal/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(version.Version)
		return
	}

	cfg, err := config.LoadServer()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	hs := headscale.New(cfg.HeadscaleURL, cfg.HeadscaleAPIKey)
	srv := server.New(hs)

	log.Printf("lanpartyd %s starting on %s", version.Version, cfg.ListenAddr)
	log.Printf("Headscale API: %s", cfg.HeadscaleURL)

	if err := http.ListenAndServe(cfg.ListenAddr, srv.Routes()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
