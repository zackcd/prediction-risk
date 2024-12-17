package main

import (
	"prediction-risk/internal/domain/services"
	"prediction-risk/internal/infrastructure/db"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"time"

	"github.com/go-chi/chi"
)

func main() {
	kalshiClient := kalshi.NewKalshiClient()
	stopLossRepo := db.NewStopLossRepoInMemory()
	stopLossService := services.NewStopLossService(stopLossRepo)
	stopLossMonitor := services.NewStopLossOrderMonitor(stopLossService, kalshiClient, 5*time.Second)

	stopLossMonitor.Start()
	defer stopLossMonitor.Stop()

	router := chi.NewRouter()
	handlers := api.NewStopLossHandlers(stopLossService)
}
