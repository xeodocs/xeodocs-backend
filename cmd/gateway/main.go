package main

import (
	"log"
	"net/http"

	"github.com/xeodocs/xeodocs-backend/internal/gateway"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/", gateway.AuthProxyHandler(cfg))
	mux.HandleFunc("/v1/users", gateway.AuthProxyHandler(cfg))
	mux.HandleFunc("/v1/users/", gateway.AuthProxyHandler(cfg))
	mux.HandleFunc("/v1/roles", gateway.AuthProxyHandler(cfg))
	mux.HandleFunc("/v1/roles/", gateway.AuthProxyHandler(cfg))

	log.Printf("Starting Gateway Service on port %s", cfg.GatewayPort)
	log.Fatal(http.ListenAndServe(":"+cfg.GatewayPort, mux))
}
