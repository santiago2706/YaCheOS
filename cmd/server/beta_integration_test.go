package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func betaRouter() http.Handler {
	base := resolveBaseDir()
	return setupRouter(filepath.Join(base, "web"), filepath.Join(base, "data"), base)
}

func TestReplacementUsesPythonTraceForCustomInput(t *testing.T) {
	router := betaRouter()
	body := `{"references":[1,2,1,3],"frames":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/memory/replace", strings.NewReader(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	var response struct {
		References []int `json:"references"`
		FIFO       struct {
			Hits   int `json:"hits"`
			Faults int `json:"faults"`
			Trace  []struct {
				Page int  `json:"page"`
				Hit  bool `json:"hit"`
			} `json:"trace"`
		} `json:"fifo"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.References) != 4 || response.FIFO.Hits != 1 || response.FIFO.Faults != 3 || len(response.FIFO.Trace) != 4 || !response.FIFO.Trace[2].Hit {
		t.Fatalf("traza Python incorrecta: %+v", response)
	}
}

func TestReplacementRejectsInvalidInput(t *testing.T) {
	for _, body := range []string{`{"references":[-1],"frames":3}`, `{"references":[1],"frames":0}`, `{"references":[1],"frames":2} {}`} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/memory/replace", strings.NewReader(body))
		w := httptest.NewRecorder()
		betaRouter().ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("input %q: status %d", body, w.Code)
		}
	}
}

func TestMemoryMapUsesCAndReleasesPID(t *testing.T) {
	if _, err := os.Stat(filepath.Join(resolveBaseDir(), "bin", "memory_cli.exe")); err != nil {
		if _, err = os.Stat(filepath.Join(resolveBaseDir(), "bin", "memory_cli")); err != nil {
			t.Skip("compile memory_cli first")
		}
	}
	router := betaRouter()
	request := func(path string) map[string]any {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status %d: %s", w.Code, w.Body.String())
		}
		var result map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	initial := request("/api/v1/memory/map")
	released := request("/api/v1/memory/map?released=2")
	if initial["total_free_kb"] != float64(448) || released["total_free_kb"] != float64(512) {
		t.Fatalf("C no liberó 64 KB: before=%v after=%v", initial["total_free_kb"], released["total_free_kb"])
	}
}

func TestThreadDemoReturnsRealTerminatedThreads(t *testing.T) {
	w := httptest.NewRecorder()
	betaRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/thread/demo", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["same_process"] != true || result["distinct_thread_ids"] != true || result["all_terminated"] != true {
		t.Fatalf("demostrador de hilos incorrecto: %v", result)
	}
}

func TestGanttSVGServed(t *testing.T) {
	w := httptest.NewRecorder()
	betaRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/docs/gantt/gantt_fcfs.svg", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "<svg") {
		t.Fatalf("Gantt no disponible: %d", w.Code)
	}
}

func TestSimulationStateStepKillAndReset(t *testing.T) {
	if _, err := os.Stat(filepath.Join(resolveBaseDir(), "bin", "memory_cli.exe")); err != nil {
		if _, err = os.Stat(filepath.Join(resolveBaseDir(), "bin", "memory_cli")); err != nil {
			t.Skip("compile memory_cli first")
		}
	}
	router := betaRouter()
	request := func(action string) map[string]any {
		method, body := http.MethodGet, ""
		if action != "" {
			method, body = http.MethodPost, action
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(method, "/api/v1/simulation", strings.NewReader(body)))
		if w.Code != http.StatusOK {
			t.Fatalf("action %s: %d %s", action, w.Code, w.Body.String())
		}
		var result map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	initial := request("")
	if initial["tick"] != float64(0) || initial["algorithm"] != "FCFS" {
		t.Fatalf("estado inicial: %v", initial)
	}
	stepped := request(`{"action":"step"}`)
	if stepped["tick"] != float64(1) {
		t.Fatalf("el reloj no avanzó: %v", stepped["tick"])
	}
	killed := request(`{"action":"kill","pid":2}`)
	if killed["memory"].(map[string]any)["total_free_kb"] != float64(512) || killed["modified"] != true {
		t.Fatalf("kill no actualizó C: %v", killed)
	}
	reset := request(`{"action":"reset","algorithm":"SJF"}`)
	if reset["tick"] != float64(0) || reset["algorithm"] != "SJF" || reset["memory"].(map[string]any)["total_free_kb"] != float64(448) {
		t.Fatalf("reset no restauró el estado: %v", reset)
	}
}
