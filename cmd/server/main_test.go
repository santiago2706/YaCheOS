package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRootEndpointReturns200(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	if _, err := os.Stat(filepath.Join(webDir, "index.html")); err != nil {
		t.Fatalf("web/index.html no fue encontrado en %s: %v", webDir, err)
	}

	router := setupRouter(webDir, dataDir, baseDir)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Se esperaba codigo HTTP 200 para '/', pero se obtuvo %d", rr.Code)
	}
}

func TestHealthEndpointReturns200AndBetaVersion(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	router := setupRouter(webDir, dataDir, baseDir)
	req, err := http.NewRequest(http.MethodGet, "/api/health", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Se esperaba codigo HTTP 200 para '/api/health', pero se obtuvo %d", rr.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("Error al decodificar JSON de health: %v", err)
	}
	if res["version"] != "v0.5.0-beta" {
		t.Errorf("Se esperaba version 'v0.5.0-beta', pero se obtuvo '%v'", res["version"])
	}
	if res["sprint"] != float64(2) {
		t.Errorf("Se esperaba sprint 2, pero se obtuvo '%v'", res["sprint"])
	}
}

// TestCPUScheduleEndpointFCFS verifica T1.4 (invocación de C y resultados FCFS).
func TestCPUScheduleEndpointFCFS(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	router := setupRouter(webDir, dataDir, baseDir)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/cpu/schedule?algo=FCFS", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Se esperaba codigo HTTP 200 para schedule FCFS, pero se obtuvo %d: %s", rr.Code, rr.Body.String())
	}

	var data map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatalf("Error decodificando respuesta JSON: %v", err)
	}

	if data["algorithm"] != "FCFS" {
		t.Errorf("Se esperaba algorithm FCFS, se obtuvo %v", data["algorithm"])
	}

	// Verificar promedio de espera de 5.75 min según contrato Semana 4
	avgWait, ok := data["average_waiting_time"].(float64)
	if !ok || avgWait != 5.75 {
		t.Errorf("Se esperaba average_waiting_time 5.75, pero se obtuvo %v", data["average_waiting_time"])
	}
}

// TestCPUScheduleCompactProcesses verifica los procesos enviados juntos en el body.
func TestCPUScheduleCompactProcesses(t *testing.T) {
	h := NewCoreHandlers(resolveBaseDir())
	body := `{"algo":"FCFS","processes":[{"pid":1,"name":"A","arrival_time":0,"burst_time":3},{"pid":2,"name":"B","arrival_time":0,"burst_time":2}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cpu/schedule", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.HandleCPUSchedule(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("schedule failed: %d %s", rr.Code, rr.Body.String())
	}
	var result struct {
		Processes []struct {
			PID int `json:"pid"`
		} `json:"processes"`
		AverageWaitingTime float64 `json:"average_waiting_time"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Processes) != 2 || result.Processes[1].PID != 2 || result.AverageWaitingTime != 1.5 {
		t.Fatalf("compact JSON must schedule both processes: %+v", result)
	}
}

// TestCPUScheduleLatencyUnder200ms verifica el DoD de T1.4 (< 200 ms).
func TestCPUScheduleLatencyUnder200ms(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	router := setupRouter(webDir, dataDir, baseDir)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/cpu/schedule?algo=SJF", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	start := time.Now()
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	elapsed := time.Since(start)

	if rr.Code != http.StatusOK {
		t.Fatalf("Se esperaba HTTP 200, se obtuvo %d", rr.Code)
	}

	if elapsed > 200*time.Millisecond {
		t.Errorf("Latencia %v superó el límite DoD de 200ms", elapsed)
	}
}

// TestCPUScheduleInvalidAlgoReturns400NoPanic verifica el manejo de errores sin panic (DoD T1.4).
func TestCPUScheduleInvalidAlgoReturns400NoPanic(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	router := setupRouter(webDir, dataDir, baseDir)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/cpu/schedule?algo=INVENTADO", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Se esperaba codigo HTTP 400 para algoritmo invalido, pero se obtuvo %d", rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "error") {
		t.Errorf("La respuesta deberia contener campo de error JSON: %s", rr.Body.String())
	}
}

// TestMemoryTranslatePaging10000 verifica el cálculo exacto de paginación (Semana 6: 10000 -> 30480).
func TestMemoryTranslatePaging10000(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	router := setupRouter(webDir, dataDir, baseDir)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/memory/translate?virtual=10000&page_size=4096", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Se esperaba HTTP 200 para translate, se obtuvo %d: %s", rr.Code, rr.Body.String())
	}

	var data map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatalf("Error al parsear JSON: %v", err)
	}

	if data["physical_address"] != float64(30480) {
		t.Errorf("Se esperaba direccion fisica 30480, pero se obtuvo %v", data["physical_address"])
	}
	if data["page_number"] != float64(2) {
		t.Errorf("Se esperaba pagina 2, pero se obtuvo %v", data["page_number"])
	}
	if data["frame_number"] != float64(7) {
		t.Errorf("Se esperaba marco 7, pero se obtuvo %v", data["frame_number"])
	}
	if data["offset"] != float64(1808) {
		t.Errorf("Se esperaba offset 1808, pero se obtuvo %v", data["offset"])
	}
}

// TestMemoryTranslateSegmentationTrapReturns400NoPanic verifica la captura de segment trap (T1.4 & T2.6).
func TestMemoryTranslateSegmentationTrapReturns400NoPanic(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	router := setupRouter(webDir, dataDir, baseDir)
	// Acceso inválido: segmento 1 (límite 800), offset 900
	req, err := http.NewRequest(http.MethodGet, "/api/v1/memory/translate?mode=segmentation&segment=1&offset=900", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Se esperaba HTTP 400 por violación de segmento (Trap), pero se obtuvo %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "error") {
		t.Errorf("Se esperaba mensaje de error detallado: %s", rr.Body.String())
	}
}

// TestPageReplacementBeladyAnomaly verifica el endpoint de reemplazo y la anomalía de Belady.
func TestPageReplacementBeladyAnomaly(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	router := setupRouter(webDir, dataDir, baseDir)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/memory/replace", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Se esperaba HTTP 200, se obtuvo %d", rr.Code)
	}

	var data map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatalf("Error parseando JSON: %v", err)
	}

	belady, ok := data["belady_anomaly"].(map[string]any)
	if !ok {
		t.Fatalf("Se esperaba campo belady_anomaly en respuesta")
	}

	if belady["fifo_3_frames_faults"] != float64(9) {
		t.Errorf("Se esperaban 9 fallos con 3 marcos, se obtuvo %v", belady["fifo_3_frames_faults"])
	}
	if belady["fifo_4_frames_faults"] != float64(10) {
		t.Errorf("Se esperaban 10 fallos con 4 marcos, se obtuvo %v", belady["fifo_4_frames_faults"])
	}
}

// TestSyncStatusEndpoint verifica la integración con el módulo de sincronización (T1.6).
func TestSyncStatusEndpoint(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	router := setupRouter(webDir, dataDir, baseDir)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/sync/status", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Se esperaba HTTP 200 para sync status, se obtuvo %d", rr.Code)
	}

	var data map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatalf("Error parseando JSON: %v", err)
	}

	if _, ok := data["philosophers"]; !ok {
		t.Errorf("Falta campo 'philosophers' en status de sincronización")
	}
	if _, ok := data["prod_cons"]; !ok {
		t.Errorf("Falta campo 'prod_cons' en status de sincronización")
	}
}

// TestCPUScheduleRoundRobinQuantum verifica el cálculo exacto de Round Robin con q=4 (Semana 4: W_prom=9.25, R_prom=4.0).
func TestCPUScheduleRoundRobinQuantum(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	router := setupRouter(webDir, dataDir, baseDir)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/cpu/schedule?algo=RR&quantum=4", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Se esperaba HTTP 200 para schedule RR, se obtuvo %d: %s", rr.Code, rr.Body.String())
	}

	var data map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatalf("Error decodificando JSON: %v", err)
	}

	if data["average_waiting_time"] != 9.25 {
		t.Errorf("Se esperaba average_waiting_time 9.25 para RR, se obtuvo %v", data["average_waiting_time"])
	}
	if data["average_response_time"] != 4.0 {
		t.Errorf("Se esperaba average_response_time 4.0 para RR, se obtuvo %v", data["average_response_time"])
	}
}

// TestMemoryTranslateSegmentationValid5500 verifica el cálculo exacto de segmentación válida (Base 5200 + 300 = 5500).
func TestMemoryTranslateSegmentationValid5500(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	router := setupRouter(webDir, dataDir, baseDir)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/memory/translate?mode=segmentation&segment=1&offset=300", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Se esperaba HTTP 200, se obtuvo %d: %s", rr.Code, rr.Body.String())
	}

	var data map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatalf("Error decodificando JSON: %v", err)
	}

	if data["physical_address"] != float64(5500) {
		t.Errorf("Se esperaba direccion fisica 5500, se obtuvo %v", data["physical_address"])
	}
}

// TestSyncStepProduceAndConsume verifica el avance paso a paso del buffer concurrente vía endpoint.
func TestSyncStepProduceAndConsume(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	router := setupRouter(webDir, dataDir, baseDir)

	// Reset
	reqReset, _ := http.NewRequest(http.MethodGet, "/api/v1/sync/step?action=reset", nil)
	rrReset := httptest.NewRecorder()
	router.ServeHTTP(rrReset, reqReset)
	if rrReset.Code != http.StatusOK {
		t.Fatalf("Error al reiniciar sincronizacion: %d", rrReset.Code)
	}

	// Produce
	reqProd, _ := http.NewRequest(http.MethodGet, "/api/v1/sync/step?action=produce", nil)
	rrProd := httptest.NewRecorder()
	router.ServeHTTP(rrProd, reqProd)
	if rrProd.Code != http.StatusOK {
		t.Fatalf("Error al producir: %d", rrProd.Code)
	}

	var dataProd map[string]any
	_ = json.Unmarshal(rrProd.Body.Bytes(), &dataProd)
	pcProd := dataProd["prod_cons"].(map[string]any)
	if pcProd["buffer_len"] != float64(1) {
		t.Errorf("Se esperaba buffer_len 1 despues de producir, se obtuvo %v", pcProd["buffer_len"])
	}

	// Consume
	reqCons, _ := http.NewRequest(http.MethodGet, "/api/v1/sync/step?action=consume", nil)
	rrCons := httptest.NewRecorder()
	router.ServeHTTP(rrCons, reqCons)
	if rrCons.Code != http.StatusOK {
		t.Fatalf("Error al consumir: %d", rrCons.Code)
	}

	var dataCons map[string]any
	_ = json.Unmarshal(rrCons.Body.Bytes(), &dataCons)
	pcCons := dataCons["prod_cons"].(map[string]any)
	if pcCons["buffer_len"] != float64(0) {
		t.Errorf("Se esperaba buffer_len 0 despues de consumir, se obtuvo %v", pcCons["buffer_len"])
	}
}
