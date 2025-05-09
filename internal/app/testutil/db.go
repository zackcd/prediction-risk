package testutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDB struct {
	container testcontainers.Container
	db        *sqlx.DB
}

func (tdb *TestDB) DB() *sqlx.DB {
	return tdb.db
}

func SetupTestDB(t *testing.T) *TestDB {
	ctx := context.Background()

	// Container request
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		WaitingFor: wait.ForAll(
			wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
				return fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", host, port.Port())
			}),
		).WithDeadline(30 * time.Second),
	}

	// Start container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "Failed to start container")

	// Get the mapped port
	mappedPort, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err, "Failed to get mapped port")

	hostIP, err := container.Host(ctx)
	require.NoError(t, err, "Failed to get host IP")

	t.Logf("Container Host IP: %s", hostIP)
	t.Logf("Mapped Port: %s", mappedPort.Port())

	// Connection string with mapped port
	connStr := fmt.Sprintf(
		"postgres://postgres:postgres@%s:%s/testdb?sslmode=disable",
		hostIP,
		mappedPort.Port(),
	)

	// Connect to the database
	db, err := sqlx.Open("postgres", connStr)
	require.NoError(t, err, "Failed to create database instance")

	err = db.Ping()
	require.NoError(t, err, "Failed to connect to database after retries")

	// Get project root directory
	projectRoot, err := findProjectRoot()
	require.NoError(t, err)

	// Run dbmate migration with explicit migrations path
	cmd := exec.Command("dbmate",
		"--url", connStr,
		"--migrations-dir", filepath.Join(projectRoot, "db/migrations"),
		"up",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("dbmate output: %s", string(output))
		require.NoError(t, err, "Failed to run migrations")
	}

	return &TestDB{
		container: container,
		db:        db,
	}
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

func (tdb *TestDB) Close(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, tdb.db.Close())
	require.NoError(t, tdb.container.Terminate(ctx))
}

func (tdb *TestDB) InsertTestData(t *testing.T) {
	queries := []string{
		`INSERT INTO weather.nws_station (station_id, name)
		VALUES ('KNYC', 'New York City, Central Park')
		ON CONFLICT (station_ID) DO NOTHING;`,
		`INSERT INTO weather.nws_station (station_id, name)
		VALUES ('047740', 'San Diego Lindbe, CA')
		ON CONFLICT (station_ID) DO NOTHING;`,
	}

	for _, query := range queries {
		_, err := tdb.db.Exec(query)
		require.NoError(t, err)
	}
}

// Cleanup truncates all project-related tables in the database
func (tdb *TestDB) Cleanup(t *testing.T) {
	query := `
		SELECT table_schema || '.' || table_name
		FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
		AND table_type = 'BASE TABLE';
	`

	var tables []string
	err := tdb.db.Select(&tables, query)
	require.NoError(t, err)

	if len(tables) > 0 {
		// Build TRUNCATE statement for all tables
		truncateQuery := fmt.Sprintf("TRUNCATE TABLE %s CASCADE;",
			strings.Join(tables, ", "))
		_, err = tdb.db.Exec(truncateQuery)
		require.NoError(t, err)
	}
}
