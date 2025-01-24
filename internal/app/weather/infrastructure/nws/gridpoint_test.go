package nws

// import (
// 	"encoding/json"
// 	"net/http"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestGridpointGetForecast(t *testing.T) {
// 	testCases := []struct {
// 		name         string
// 		officeID     string
// 		xCoordinate  int
// 		yCoordinate  int
// 		setupMock    func(w http.ResponseWriter, r *http.Request)
// 		wantErr      bool
// 		wantStatus   int
// 	}{
// 		{
// 			name:         "successfully gets forecast",
// 			officeID:     "OKX",
// 			xCoordinate:  41,
// 			yCoordinate:  -74,
// 			setupMock: func(w http.ResponseWriter, r *http.Request) {
// 				assert.Equal(t, "/gridpoints/OKX/41,-74/forecast", r.URL.Path)
// 				assert.Equal(t, "test-user-agent", r.Header.Get("User-Agent"))
// 				assert.Equal(t, "application/geo+json", r.Header.Get("Accept"))

// 				forecast := &Forecast{
// 					ID:   "https://api.weather.gov/gridpoints/OKX/41,-74/forecast",
// 					Type: "Feature",
// 					Properties: ForecastProperties{
// 						Geometry:          "Point",
// 						Units:             "us",
// 						ForecastGenerator: "BaselineForecastGenerator",
// 						GeneratedAt:       time.Now(),
// 						UpdateTime:        time.Now(),
// 						Elevation: QuantitativeValue{
// 							Value:    42.7,
// 							UnitCode: "unit:m",
// 						},
// 						Periods: []ForecastPeriod{
// 							{
// 								Number:           1,
// 								Name:            "Tonight",
// 								StartTime:       time.Now(),
// 								EndTime:         time.Now().Add(12 * time.Hour),
// 								IsDaytime:       false,
// 								TemperatureTrend: "falling",
// 								WindDirection:    "N",
// 								ShortForecast:   "Clear",
// 								DetailedForecast: "Clear overnight",
// 							},
// 						},
// 					},
// 				}

// 				w.Header().Set("Content-Type", "application/geo+json")
// 				json.NewEncoder(w).Encode(forecast)
// 			},
// 		},
// 		{
// 			name:         "handles server error",
// 			officeID:     "OKX",
// 			xCoordinate:  41,
// 			yCoordinate:  -74,
// 			setupMock: func(w http.ResponseWriter, r *http.Request) {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				w.Write([]byte("internal server error"))
// 			},
// 			wantErr:    true,
// 			wantStatus: http.StatusInternalServerError,
// 		},
// 		{
// 			name:         "handles invalid coordinates",
// 			officeID:     "OKX",
// 			xCoordinate:  999,
// 			yCoordinate:  999,
// 			setupMock: func(w http.ResponseWriter, r *http.Request) {
// 				w.WriteHeader(http.StatusNotFound)
// 				w.Write([]byte("coordinates not found"))
// 			},
// 			wantErr:    true,
// 			wantStatus: http.StatusNotFound,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			server, client := setupTestServer(tc.setupMock)
// 			defer server.Close()

// 			result, err := client.Gridpoint.GetForecast(tc.officeID, tc.xCoordinate, tc.yCoordinate)

// 			if tc.wantErr {
// 				require.Error(t, err)
// 				nwsErr, ok := err.(*NWSError)
// 				require.True(t, ok)
// 				assert.Equal(t, tc.wantStatus, nwsErr.StatusCode)
// 				return
// 			}

// 			require.NoError(t, err)
// 			require.NotNil(t, result)
// 			assert.Equal(t, "Feature", result.Type)
// 			assert.NotEmpty(t, result.Properties.Periods)
// 		})
// 	}
// }

// func TestGridpointGetHourlyForecast(t *testing.T) {
// 	testCases := []struct {
// 		name         string
// 		officeID     string
// 		xCoordinate  int
// 		yCoordinate  int
// 		setupMock    func(w http.ResponseWriter, r *http.Request)
// 		wantErr      bool
// 		wantStatus   int
// 	}{
// 		{
// 			name:         "successfully gets hourly forecast",
// 			officeID:     "OKX",
// 			xCoordinate:  41,
// 			yCoordinate:  -74,
// 			setupMock: func(w http.ResponseWriter, r *http.Request) {
// 				assert.Equal(t, "/gridpoints/OKX/41,-74/forecast/hourly", r.URL.Path)
// 				assert.Equal(t, "test-user-agent", r.Header.Get("User-Agent"))
// 				assert.Equal(t, "application/geo+json", r.Header.Get("Accept"))

// 				forecast := &Forecast{
// 					ID:   "https://api.weather.gov/gridpoints/OKX/41,-74/forecast/hourly",
// 					Type: "Feature",
// 					Properties: ForecastProperties{
// 						Geometry:          "Point",
// 						Units:             "us",
// 						ForecastGenerator: "HourlyForecastGenerator",
// 						GeneratedAt:       time.Now(),
// 						UpdateTime:        time.Now(),
// 						Elevation: QuantitativeValue{
// 							Value:    42.7,
// 							UnitCode: "unit:m",
// 						},
// 						Periods: []ForecastPeriod{
// 							{
// 								Number:           1,
// 								Name:            "Now",
// 								StartTime:       time.Now(),
// 								EndTime:         time.Now().Add(1 * time.Hour),
// 								IsDaytime:       false,
// 								TemperatureTrend: "steady",
// 								WindDirection:    "N",
// 								ShortForecast:   "Clear",
// 								DetailedForecast: "Clear conditions",
// 							},
// 						},
// 					},
// 				}

// 				w.Header().Set("Content-Type", "application/geo+json")
// 				json.NewEncoder(w).Encode(forecast)
// 			},
// 		},
// 		{
// 			name:         "handles server error",
// 			officeID:     "OKX",
// 			xCoordinate:  41,
// 			yCoordinate:  -74,
// 			setupMock: func(w http.ResponseWriter, r *http.Request) {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				w.Write([]byte("internal server error"))
// 			},
// 			wantErr:    true,
// 			wantStatus: http.StatusInternalServerError,
// 		},
// 		{
// 			name:         "handles invalid coordinates",
// 			officeID:     "OKX",
// 			xCoordinate:  999,
// 			yCoordinate:  999,
// 			setupMock: func(w http.ResponseWriter, r *http.Request) {
// 				w.WriteHeader(http.StatusNotFound)
// 				w.Write([]byte("coordinates not found"))
// 			},
// 			wantErr:    true,
// 			wantStatus: http.StatusNotFound,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			server, client := setupTestServer(tc.setupMock)
// 			defer server.Close()

// 			result, err := client.Gridpoint.GetHourlyForecast(tc.officeID, tc.xCoordinate, tc.yCoordinate)

// 			if tc.wantErr {
// 				require.Error(t, err)
// 				nwsErr, ok := err.(*NWSError)
// 				require.True(t, ok)
// 				assert.Equal(t, tc.wantStatus, nwsErr.StatusCode)
// 				return
// 			}

// 			require.NoError(t, err)
// 			require.NotNil(t, result)
// 			assert.Equal(t, "Feature", result.Type)
// 			assert.NotEmpty(t, result.Properties.Periods)
// 		})
// 	}
// }
