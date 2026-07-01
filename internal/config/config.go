package config

import (
	"fmt"
	"os"
)

type Server struct {
	ListenAddr     string
	HeadscaleURL   string
	HeadscaleAPIKey string
}

type Client struct {
	ServerURL string
}

func LoadServer() (Server, error) {
	cfg := Server{
		ListenAddr:      env("LANPARTYD_LISTEN_ADDR", "0.0.0.0:8090"),
		HeadscaleURL:    env("HEADSCALE_URL", ""),
		HeadscaleAPIKey: env("HEADSCALE_API_KEY", ""),
	}

	if cfg.HeadscaleURL == "" {
		return cfg, fmt.Errorf("HEADSCALE_URL is required")
	}
	if cfg.HeadscaleAPIKey == "" {
		return cfg, fmt.Errorf("HEADSCALE_API_KEY is required")
	}

	return cfg, nil
}

func LoadClient() Client {
	return Client{
		ServerURL: env("LANPARTY_SERVER_URL", ""),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
