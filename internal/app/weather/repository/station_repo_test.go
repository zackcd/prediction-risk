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

func createTestStation() *weather_domain.NWSStation {
	return &weather_domain.NWSStation{
		StationID: "KNYC",
		Name:      "New York Central Park",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func TestNWSStationRepository_GetStation(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Close(t)

	repo := NewStationRepo(testDB.DB())

	t.Run("get existing station", func(t *testing.T) {
		defer testDB.Cleanup(t)

		station := createTestStation()
		err := repo.Persist(station)
		require.NoError(t, err)

		found, err := repo.GetByID(station.StationID)
		require.NoError(t, err)
		assert.Equal(t, station.StationID, found.StationID)
		assert.Equal(t, station.Name, found.Name)
		assert.WithinDuration(t, station.CreatedAt, found.CreatedAt, time.Second)
		assert.WithinDuration(t, station.UpdatedAt, found.UpdatedAt, time.Second)
	})

	t.Run("station not found", func(t *testing.T) {
		defer testDB.Cleanup(t)

		_, err := repo.GetByID("NONEXISTENT")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "station not found")
	})
}

func TestNWSStationRepository_Persist(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Close(t)

	repo := NewStationRepo(testDB.DB())

	t.Run("creates new station", func(t *testing.T) {
		defer testDB.Cleanup(t)

		station := createTestStation()
		err := repo.Persist(station)
		require.NoError(t, err)

		saved, err := repo.GetByID(station.StationID)
		require.NoError(t, err)
		assert.Equal(t, station.StationID, saved.StationID)
		assert.Equal(t, station.Name, saved.Name)
		assert.WithinDuration(t, station.CreatedAt, saved.CreatedAt, time.Second)
	})

	t.Run("updates existing station", func(t *testing.T) {
		defer testDB.Cleanup(t)

		// Create initial station
		station := createTestStation()
		err := repo.Persist(station)
		require.NoError(t, err)

		// Update station
		updatedStation := &weather_domain.NWSStation{
			StationID: station.StationID,
			Name:      "Central Park Weather Station",
			CreatedAt: station.CreatedAt,
			UpdatedAt: time.Now().UTC().Add(time.Hour), // simulate time passing
		}
		err = repo.Persist(updatedStation)
		require.NoError(t, err)

		// Verify updates
		saved, err := repo.GetByID(station.StationID)
		require.NoError(t, err)
		assert.Equal(t, updatedStation.StationID, saved.StationID)
		assert.Equal(t, "Central Park Weather Station", saved.Name)
		assert.WithinDuration(t, station.CreatedAt, saved.CreatedAt, time.Second) // should not change
		assert.True(t, saved.UpdatedAt.After(station.UpdatedAt), "UpdatedAt should be updated")
	})
}
