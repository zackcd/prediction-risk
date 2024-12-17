package main

import (
	"prediction-risk/internal/domain/services"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"time"
)

func main() {
	kalshiClient := kalshi.NewKalshiClient()
	stopLossRepo := repositories.NewStopLossOrderRepo()
	stopLossService := services.NewStopLossService(stopLossRepo)
	stopLossMonitor := services.NewStopLossOrderMonitor(stopLossService, kalshiClient, 5*time.Second)

	stopLossMonitor.Start()
	defer stopLossMonitor.Stop()

	router := chi.NewRouter()
	handlers := api.NewStopLossHandlers(stopLossService)
}
