package weather_repository

import (
	"prediction-risk/internal/app/testutil"
	weather_domain "prediction-risk/internal/app/weather/domain"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestObservation(stationID string) *weather_domain.TemperatureObservation {
	return weather_domain.NewTemperatureObservation(
		stationID,
		weather_domain.Temperature{
			Value:           23.5,
			TemperatureUnit: weather_domain.Celsius,
		},
		time.Now().UTC(),
	)
}

func TestTemperatureObservationRepo_Persist(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Close(t)

	repo := NewTemperatureObservationRepo(testDB.DB())

	t.Run("creates new observation", func(t *testing.T) {
		testDB.InsertTestData(t)
		defer testDB.Cleanup(t)

		observation := createTestObservation("KNYC")
		err := repo.Persist(observation)
		require.NoError(t, err)

		// Verify by retrieving
		filter := &weather_domain.TemperatureObservationFilter{
			StationID: &observation.StationID,
		}

		saved, err := repo.Get(filter)
		require.NoError(t, err)
		require.Len(t, saved, 1)

		assert.Equal(t, observation.ObservationID, saved[0].ObservationID)
		assert.Equal(t, observation.StationID, saved[0].StationID)
		assert.Equal(t, observation.Temperature.Value, saved[0].Temperature.Value)
		assert.Equal(t, observation.Temperature.TemperatureUnit, saved[0].Temperature.TemperatureUnit)
		assert.WithinDuration(t, observation.Timestamp, saved[0].Timestamp, time.Second)
	})

	t.Run("updates existing observation", func(t *testing.T) {
		testDB.InsertTestData(t)
		defer testDB.Cleanup(t)

		// Create initial observation
		observation := createTestObservation("KNYC")
		err := repo.Persist(observation)
		require.NoError(t, err)

		// Update observation
		observation.Temperature.Value = 24.5
		observation.UpdatedAt = time.Now().UTC().Add(time.Hour)
		err = repo.Persist(observation)
		require.NoError(t, err)

		// Verify update
		filter := &weather_domain.TemperatureObservationFilter{
			StationID: &observation.StationID,
		}

		saved, err := repo.Get(filter)
		require.NoError(t, err)
		require.Len(t, saved, 1)

		assert.Equal(t, 24.5, saved[0].Temperature.Value)
		assert.True(t, saved[0].UpdatedAt.After(saved[0].CreatedAt))
		assert.Equal(t, observation.ObservationID, saved[0].ObservationID)
	})

	t.Run("persists multiple observations for same station", func(t *testing.T) {
		testDB.InsertTestData(t)
		defer testDB.Cleanup(t)

		stationID := "KNYC"

		// Create first observation
		obs1 := weather_domain.NewTemperatureObservation(
			stationID,
			weather_domain.Temperature{Value: 23.5, TemperatureUnit: weather_domain.Celsius},
			time.Now().UTC(),
		)
		err := repo.Persist(obs1)
		require.NoError(t, err)

		// Create second observation
		obs2 := weather_domain.NewTemperatureObservation(
			stationID,
			weather_domain.Temperature{Value: 24.5, TemperatureUnit: weather_domain.Celsius},
			time.Now().UTC().Add(time.Hour),
		)
		err = repo.Persist(obs2)
		require.NoError(t, err)

		filter := &weather_domain.TemperatureObservationFilter{
			StationID: &stationID,
		}

		saved, err := repo.Get(filter)
		require.NoError(t, err)
		require.Len(t, saved, 2)

		// Should be ordered by timestamp DESC
		assert.Equal(t, 24.5, saved[0].Temperature.Value)
		assert.Equal(t, 23.5, saved[1].Temperature.Value)
	})
}

func TestTemperatureObservationRepo_Get(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Close(t)

	repo := NewTemperatureObservationRepo(testDB.DB())

	t.Run("get with station filter", func(t *testing.T) {
		testDB.InsertTestData(t)
		defer testDB.Cleanup(t)

		// Create observations for different stations
		obs1 := createTestObservation("KNYC")
		obs2 := createTestObservation("047740")

		require.NoError(t, repo.Persist(obs1))
		require.NoError(t, repo.Persist(obs2))

		// Filter by first station
		filter := &weather_domain.TemperatureObservationFilter{
			StationID: &obs1.StationID,
		}

		results, err := repo.Get(filter)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "KNYC", results[0].StationID)
	})

	t.Run("get all observations", func(t *testing.T) {
		testDB.InsertTestData(t)
		defer testDB.Cleanup(t)

		// Create observations for different stations
		obs1 := createTestObservation("KNYC")
		obs2 := createTestObservation("047740")

		require.NoError(t, repo.Persist(obs1))
		require.NoError(t, repo.Persist(obs2))

		// Get without filter
		results, err := repo.Get(nil)
		require.NoError(t, err)
		require.Len(t, results, 2)
	})

	t.Run("get with non-existent station", func(t *testing.T) {
		testDB.InsertTestData(t)
		defer testDB.Cleanup(t)

		nonExistentStation := "NONEXISTENT"
		filter := &weather_domain.TemperatureObservationFilter{
			StationID: &nonExistentStation,
		}

		results, err := repo.Get(filter)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("returned observations are ordered by timestamp DESC", func(t *testing.T) {
		testDB.InsertTestData(t)
		defer testDB.Cleanup(t)

		stationID := "KNYC"
		now := time.Now().UTC()

		// Create three observations with different timestamps
		times := []time.Time{
			now.Add(-2 * time.Hour),
			now.Add(-1 * time.Hour),
			now,
		}

		for i, timestamp := range times {
			obs := weather_domain.NewTemperatureObservation(
				stationID,
				weather_domain.Temperature{
					Value:           20.0 + float64(i),
					TemperatureUnit: weather_domain.Celsius,
				},
				timestamp,
			)
			require.NoError(t, repo.Persist(obs))
		}

		filter := &weather_domain.TemperatureObservationFilter{
			StationID: &stationID,
		}

		results, err := repo.Get(filter)
		require.NoError(t, err)
		require.Len(t, results, 3)

		// Verify timestamps are in descending order
		for i := 1; i < len(results); i++ {
			assert.True(t, results[i-1].Timestamp.After(results[i].Timestamp))
		}

		// Verify temperatures match expected order
		assert.Equal(t, 22.0, results[0].Temperature.Value)
		assert.Equal(t, 21.0, results[1].Temperature.Value)
		assert.Equal(t, 20.0, results[2].Temperature.Value)
	})
}
