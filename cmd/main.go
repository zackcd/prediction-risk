package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"prediction-risk/internal/config"
	"prediction-risk/internal/domain/services"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"prediction-risk/internal/infrastructure/repositories/inmemory"
	"prediction-risk/internal/interfaces/api"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	kalshiPrivateKey, err := parsePrivateKey(config.Kalshi.PrivateKey)
	if err != nil {
		log.Fatalf("error parsing Kalshi private key: %v", err)
	}

	// Setup external dependencies
	kalshiClient := kalshi.NewKalshiClient(
		config.Kalshi.BaseURL,
		config.Kalshi.APIKeyID,
		kalshiPrivateKey,
	)
	stopLossRepo := inmemory.NewStopLossRepoInMemory()

	// Setup internal services
	exchangeService := services.NewExchangeService(kalshiClient.Market, kalshiClient.Portfolio)
	stopLossService := services.NewStopLossService(stopLossRepo, exchangeService)
	stopLossMonitor := services.NewStopLossMonitor(stopLossService, exchangeService, 5*time.Second)

	// Start background processes monitoring
	stopLossMonitor.Start(config.IsDryRun)
	defer stopLossMonitor.Stop()

	// Setup router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Mount routes
	stopLossRoutes := api.NewStopLossRoutes(stopLossService)
	stopLossRoutes.Register(router)

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
		Handler: router,
	}

	log.Printf("Server listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error starting server: %v", err)
	}
}

func parsePrivateKey(pemEncodedKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemEncodedKey))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %v", err)
	}

	return privateKey, nil
}
