package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"ecoguardian-go/internal/database"
	appRouter "ecoguardian-go/internal/router"
	"ecoguardian-go/internal/service"

	"github.com/gin-gonic/gin"
)

type sensorResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Location   string `json:"location"`
	SensorType string `json:"sensor_type"`
}

type alertResponse struct {
	ID       uint   `json:"id"`
	SensorID uint   `json:"sensor_id"`
	Level    string `json:"level"`
	Message  string `json:"message"`
	Status   string `json:"status"`
}

type readingResponse struct {
	ID       uint `json:"id"`
	SensorID uint `json:"sensor_id"`
}

type createReadingResponse struct {
	Reading       readingResponse `json:"reading"`
	CreatedAlerts []alertResponse `json:"created_alerts"`
}

type dashboardSummaryResponse struct {
	SensorsCount      int `json:"sensors_count"`
	ReadingsCount     int `json:"readings_count"`
	ActiveAlertsCount int `json:"active_alerts_count"`
}

type simulatorStatusResponse struct {
	Running             bool      `json:"running"`
	IntervalSeconds     int       `json:"interval_seconds"`
	LastGeneratedAt     time.Time `json:"last_generated_at"`
	LastGeneratedCount  int       `json:"last_generated_count"`
	LastGeneratedAlerts int       `json:"last_generated_alerts"`
}

type simulatorGenerateOnceResponse struct {
	Message         string                  `json:"message"`
	Generated       int                     `json:"generated"`
	GeneratedAlerts int                     `json:"generated_alerts"`
	Status          simulatorStatusResponse `json:"status"`
}

func TestCreateSensorAndListSensors(t *testing.T) {
	engine := setupTestServer(t)

	rec := performJSONRequest(t, engine, http.MethodPost, "/api/sensors", map[string]any{
		"name":        "Air Sensor A1",
		"location":    "Saint Petersburg Center",
		"sensor_type": "air",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var created sensorResponse
	decodeJSON(t, rec, &created)

	if created.ID == 0 {
		t.Fatal("expected created sensor id to be non-zero")
	}
	if created.Name != "Air Sensor A1" {
		t.Fatalf("expected sensor name Air Sensor A1, got %s", created.Name)
	}

	rec = performJSONRequest(t, engine, http.MethodGet, "/api/sensors", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var sensors []sensorResponse
	decodeJSON(t, rec, &sensors)

	if len(sensors) != 1 {
		t.Fatalf("expected 1 sensor, got %d", len(sensors))
	}
	if sensors[0].Name != "Air Sensor A1" {
		t.Fatalf("expected listed sensor name Air Sensor A1, got %s", sensors[0].Name)
	}
}

func TestCreateReadingCreatesAlertsAndUpdatesSummary(t *testing.T) {
	engine := setupTestServer(t)

	sensorID := createSensor(t, engine, "Air Sensor A1", "Center", "air")

	rec := performJSONRequest(t, engine, http.MethodPost, "/api/readings", map[string]any{
		"sensor_id":   sensorID,
		"pm25":        60.0,
		"co2":         1600.0,
		"noise":       90.0,
		"temperature": 23.5,
		"humidity":    54.0,
		"water_ph":    9.1,
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var created createReadingResponse
	decodeJSON(t, rec, &created)

	if created.Reading.ID == 0 {
		t.Fatal("expected reading id to be non-zero")
	}
	if len(created.CreatedAlerts) != 4 {
		t.Fatalf("expected 4 created alerts, got %d", len(created.CreatedAlerts))
	}

	rec = performJSONRequest(t, engine, http.MethodGet, "/api/dashboard/summary", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var summary dashboardSummaryResponse
	decodeJSON(t, rec, &summary)

	if summary.SensorsCount != 1 {
		t.Fatalf("expected sensors_count = 1, got %d", summary.SensorsCount)
	}
	if summary.ReadingsCount != 1 {
		t.Fatalf("expected readings_count = 1, got %d", summary.ReadingsCount)
	}
	if summary.ActiveAlertsCount != 4 {
		t.Fatalf("expected active_alerts_count = 4, got %d", summary.ActiveAlertsCount)
	}

	rec = performJSONRequest(t, engine, http.MethodGet, "/api/alerts", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var alerts []alertResponse
	decodeJSON(t, rec, &alerts)

	if len(alerts) != 4 {
		t.Fatalf("expected 4 alerts in list, got %d", len(alerts))
	}
}

func TestResolveAlert(t *testing.T) {
	engine := setupTestServer(t)

	sensorID := createSensor(t, engine, "Noise Sensor N1", "Nevsky", "noise")

	rec := performJSONRequest(t, engine, http.MethodPost, "/api/readings", map[string]any{
		"sensor_id":   sensorID,
		"noise":       80.0,
		"temperature": 20.0,
		"humidity":    45.0,
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var created createReadingResponse
	decodeJSON(t, rec, &created)

	if len(created.CreatedAlerts) == 0 {
		t.Fatal("expected at least one created alert")
	}

	alertID := created.CreatedAlerts[0].ID

	rec = performJSONRequest(t, engine, http.MethodPost, "/api/alerts/"+strconv.FormatUint(uint64(alertID), 10)+"/resolve", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var resolved alertResponse
	decodeJSON(t, rec, &resolved)

	if resolved.Status != "resolved" {
		t.Fatalf("expected resolved status, got %s", resolved.Status)
	}
}

func TestSimulatorGenerateOnce(t *testing.T) {
	engine := setupTestServer(t)

	createSensor(t, engine, "Air Sensor A1", "Center", "air")
	createSensor(t, engine, "Water Sensor W1", "River", "water")

	rec := performJSONRequest(t, engine, http.MethodPost, "/api/simulator/generate-once", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var result simulatorGenerateOnceResponse
	decodeJSON(t, rec, &result)

	if result.Generated != 2 {
		t.Fatalf("expected generated = 2, got %d", result.Generated)
	}

	if result.Status.LastGeneratedCount != 2 {
		t.Fatalf("expected last_generated_count = 2, got %d", result.Status.LastGeneratedCount)
	}

	rec = performJSONRequest(t, engine, http.MethodGet, "/api/dashboard/summary", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var summary dashboardSummaryResponse
	decodeJSON(t, rec, &summary)

	if summary.ReadingsCount != 2 {
		t.Fatalf("expected readings_count = 2, got %d", summary.ReadingsCount)
	}
}

func setupTestServer(t *testing.T) *gin.Engine {
	t.Helper()

	changeToProjectRoot(t)

	gin.SetMode(gin.TestMode)

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Init(dbPath)
	if err != nil {
		t.Fatalf("failed to init test database: %v", err)
	}

	simulator := service.NewSimulatorService(db, time.Second)
	t.Cleanup(func() {
		_ = simulator.Stop()
	})

	return appRouter.SetupRouter(db, simulator)
}

func changeToProjectRoot(t *testing.T) {
	t.Helper()

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	root := originalWD
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}

		parent := filepath.Dir(root)
		if parent == root {
			t.Fatal("failed to find project root with go.mod")
		}
		root = parent
	}

	if err := os.Chdir(root); err != nil {
		t.Fatalf("failed to change directory to project root: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
}

func createSensor(t *testing.T, engine *gin.Engine, name, location, sensorType string) uint {
	t.Helper()

	rec := performJSONRequest(t, engine, http.MethodPost, "/api/sensors", map[string]any{
		"name":        name,
		"location":    location,
		"sensor_type": sensorType,
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var created sensorResponse
	decodeJSON(t, rec, &created)

	return created.ID
}

func performJSONRequest(t *testing.T, engine *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&requestBody).Encode(body); err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &requestBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	return rec
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.Unmarshal(rec.Body.Bytes(), target); err != nil {
		t.Fatalf("failed to decode response json: %v; body: %s", err, rec.Body.String())
	}
}

func itoa(value uint) string {
	return json.Number(stringifyUint(value)).String()
}

func stringifyUint(value uint) string {
	if value == 0 {
		return "0"
	}

	var digits []byte
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}
