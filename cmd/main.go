package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"prediction-risk/internal/app/exchange/infrastructure/kalshi"
	exchange_service "prediction-risk/internal/app/exchange/service"
	trigger_repository "prediction-risk/internal/app/risk/trigger/repository"
	trigger_service "prediction-risk/internal/app/risk/trigger/service"
	"prediction-risk/internal/config"
	"prediction-risk/internal/interfaces/api"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"postgres",
		config.Databases.Port,
		config.Databases.User,
		config.Databases.Password,
		config.Databases.Name,
	)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	triggerRepo := trigger_repository.NewTriggerRepository(db)
	exchangeService := exchange_service.NewExchangeService(kalshiClient)
	triggerService := trigger_service.NewTriggerService(triggerRepo)
	triggerExecutor := trigger_service.NewTriggerExecutor(triggerService, exchangeService)
	triggerMonitor := trigger_service.NewTriggerMonitor(triggerService, triggerExecutor, exchangeService, 5*time.Second, config.IsDryRun)

	// Run monitors
	monitors := []Monitor{triggerMonitor}
	for _, m := range monitors {
		RunMonitor(m)
	}

	// Setup router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Mount routes
	stopTriggerRoutes := api.NewStopTriggerRoutes(triggerService)
	stopTriggerRoutes.Register(router)

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
