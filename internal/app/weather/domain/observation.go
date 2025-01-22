package weather_domain

import "time"


type RetrievalStats struct {
	TotalObservations     int
	MissingTemperature    int
	StoredObservations    int
	ObservationsWithError []time.Time
}
