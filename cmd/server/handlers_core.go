package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"sync"
	syncgo "yakcheos/src/sync_go"
)

// CoreHandlers gestiona los endpoints de integración con los motores C y Go.
type CoreHandlers struct {
	baseDir      string
	mu           sync.RWMutex // Protege el reemplazo de los punteros de simulación
	taskMu       sync.Mutex
	task         *taskSimulation
	philosophers *syncgo.DiningPhilosophers
	buffer       *syncgo.BoundedBuffer
}

// NewCoreHandlers inicializa los manejadores del núcleo.
func NewCoreHandlers(baseDir string) *CoreHandlers {
	return &CoreHandlers{
		baseDir:      baseDir,
		philosophers: syncgo.NewDiningPhilosophers("deadlock"),
		buffer:       syncgo.NewBoundedBuffer(5),
	}
}

// findBinary busca la ruta de un ejecutable compilado.
func (h *CoreHandlers) findBinary(name string) string {
	candidates := []string{
		filepath.Join(h.baseDir, "bin", name),
		filepath.Join(h.baseDir, "bin", name+".exe"),
		filepath.Join(h.baseDir, "src", "core_c", name),
		filepath.Join(h.baseDir, "src", "core_c", name+".exe"),
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}
	// Buscar en PATH
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	return ""
}

// ScheduleRequest representa los parámetros de entrada para la planificación de CPU.
type ScheduleRequest struct {
	Algo      string `json:"algo"`
	File      string `json:"file"`
	Quantum   int    `json:"quantum"`
	Processes []struct {
		PID         int    `json:"pid"`
		Name        string `json:"name"`
		ArrivalTime int    `json:"arrival_time"`
		BurstTime   int    `json:"burst_time"`
		Priority    int    `json:"priority"`
		MemoryKB    int    `json:"memory_kb"`
	} `json:"processes,omitempty"`
}

// HandleCPUSchedule maneja /api/v1/cpu/schedule (T1.4).
func (h *CoreHandlers) HandleCPUSchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Proteger de panics
	defer func() {
		if rec := recover(); rec != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":   "Error interno en el servidor",
				"details": fmt.Sprintf("%v", rec),
			})
		}
	}()

	var req ScheduleRequest
	if r.Method == http.MethodPost {
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}
	}

	// Si no vino en el body, leer parámetros de URL
	if req.Algo == "" {
		req.Algo = r.URL.Query().Get("algo")
	}
	if req.File == "" {
		req.File = r.URL.Query().Get("file")
	}
	if req.Quantum <= 0 {
		qStr := r.URL.Query().Get("quantum")
		if q, err := strconv.Atoi(qStr); err == nil && q > 0 {
			req.Quantum = q
		} else {
			req.Quantum = 4
		}
	}

	if req.Algo == "" {
		req.Algo = "FCFS"
	}
	req.Algo = strings.ToUpper(req.Algo)

	// Validar algoritmo
	validAlgos := map[string]bool{"FCFS": true, "SJF": true, "RR": true, "ROUNDROBIN": true, "PRIORIDAD": true, "PRIORITY": true}
	if !validAlgos[req.Algo] {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":     fmt.Sprintf("Algoritmo no soportado: '%s'", req.Algo),
			"details":   "Algoritmos soportados: FCFS, SJF, RR, PRIORIDAD",
			"exit_code": 1,
		})
		return
	}

	// Archivo de procesos por defecto
	filePath := req.File
	if filePath == "" {
		filePath = filepath.Join(h.baseDir, "data", "procesos_semana4.json")
	} else if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(h.baseDir, filePath)
	}

	// Si vinieron procesos en el body, escribirlos a un archivo temporal para el binario C
	var tempFile string
	if len(req.Processes) > 0 {
		for i := range req.Processes {
			if req.Processes[i].Priority == 0 {
				req.Processes[i].Priority = 1
			}
		}
		tmp, err := os.CreateTemp("", "yakcheos_procs_*.json")
		if err == nil {
			tempFile = tmp.Name()
			_ = json.NewEncoder(tmp).Encode(req.Processes)
			_ = tmp.Close()
			filePath = tempFile
			defer os.Remove(tempFile)
		}
	}

	// Buscar binario C del planificador
	binPath := h.findBinary("scheduler")
	if binPath == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":     "Binario del planificador C no encontrado (ejecute 'make' en src/core_c)",
			"exit_code": 127,
		})
		return
	}

	// Invocar binario C vía os/exec con timeout defensivo
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, "--algo", req.Algo, "--file", filePath, "--quantum", strconv.Itoa(req.Quantum))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":      "Error al ejecutar el motor de C",
			"exit_code":  exitCode,
			"details":    strings.TrimSpace(stderr.String()),
			"latency_ms": latencyMs,
		})
		return
	}

	// Validar que la salida sea JSON válido
	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":      "La salida del binario C no es JSON válido",
			"raw_output": stdout.String(),
			"exit_code":  1,
		})
		return
	}

	result["latency_ms"] = latencyMs
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// MemoryTranslateRequest representa la petición de traducción de memoria (T1.4).
type MemoryTranslateRequest struct {
	Mode           string `json:"mode"`
	VirtualAddress int    `json:"virtual_address"`
	PageSize       int    `json:"page_size"`
	SegmentID      int    `json:"segment_id"`
	Offset         int    `json:"offset"`
}

// HandleMemoryTranslate maneja /api/v1/memory/translate (T1.4).
func (h *CoreHandlers) HandleMemoryTranslate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer func() {
		if rec := recover(); rec != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":   "Error interno en el servidor",
				"details": fmt.Sprintf("%v", rec),
			})
		}
	}()

	var req MemoryTranslateRequest
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	// Parámetros de URL
	if req.Mode == "" {
		req.Mode = r.URL.Query().Get("mode")
	}
	if req.VirtualAddress == 0 {
		if vStr := r.URL.Query().Get("virtual"); vStr != "" {
			if v, err := strconv.Atoi(vStr); err == nil {
				req.VirtualAddress = v
			}
		}
	}
	if req.PageSize <= 0 {
		if pStr := r.URL.Query().Get("page_size"); pStr != "" {
			if p, err := strconv.Atoi(pStr); err == nil {
				req.PageSize = p
			}
		} else {
			req.PageSize = 4096
		}
	}
	if req.SegmentID == 0 && r.URL.Query().Get("segment") != "" {
		if s, err := strconv.Atoi(r.URL.Query().Get("segment")); err == nil {
			req.SegmentID = s
		}
	}
	if req.Offset == 0 && r.URL.Query().Get("offset") != "" {
		if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
			req.Offset = o
		}
	}

	binPath := h.findBinary("memory_translate")
	if binPath == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":     "Binario traductor de memoria C no encontrado (ejecute 'make' en src/core_c)",
			"exit_code": 127,
		})
		return
	}

	args := []string{}
	if req.Mode == "segmentation" || r.URL.Query().Get("segment") != "" {
		args = append(args, "--mode", "segmentation", "--segment", strconv.Itoa(req.SegmentID), "--offset", strconv.Itoa(req.Offset))
	} else {
		args = append(args, "--mode", "paging", "--virtual", strconv.Itoa(req.VirtualAddress), "--page-size", strconv.Itoa(req.PageSize))
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":      "Error en la traducción de memoria (Trap / Excepción)",
			"exit_code":  exitCode,
			"details":    strings.TrimSpace(stderr.String()),
			"latency_ms": latencyMs,
		})
		return
	}

	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":      "La salida del binario C no es JSON válido",
			"raw_output": stdout.String(),
			"exit_code":  1,
		})
		return
	}

	result["latency_ms"] = latencyMs
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// HandleSyncStatus maneja /api/v1/sync/status para consultar el estado del buffer y filósofos.
func (h *CoreHandlers) HandleSyncStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Obtenemos los punteros de simulación de forma segura
	h.mu.RLock()
	dp := h.philosophers
	bb := h.buffer
	h.mu.RUnlock()

	// 1. Snapshot de filósofos: Copia segura de arreglos y cálculo de RAG en la misma foto
	snap, rag := dp.GetSnapshotAndReport()

	// 2. Snapshot de buffer: GetState y GetEvents ya devuelven copias profundas (Deep Copy)
	count, items, mutexVal, emptyVal, fullVal := bb.GetState()

	res := map[string]any{
		"philosophers": map[string]any{
			"count":       snap.Count,
			"states":      snap.States,
			"forks_owner": snap.ForksOwner,
			"mode":        snap.Mode,
			"deadlock":    snap.Deadlock,
			"cycle":       snap.Cycle,
			"rag_report":  rag,
			"eat_count":   snap.EatCount,
		},
		"prod_cons": map[string]any{
			"capacity":   5,
			"buffer_len": count,
			"items":      items,
			"semaphores": map[string]int{
				"mutex":  mutexVal,
				"vacios": emptyVal,
				"llenos": fullVal,
			},
			"events": bb.GetEvents(),
		},
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}

// HandleSyncStep avanza 1 paso en la simulación de concurrencia.
func (h *CoreHandlers) HandleSyncStep(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	mode := r.URL.Query().Get("mode")
	actor := r.URL.Query().Get("actor")
	action := r.URL.Query().Get("action")

	// Usamos RLock para permitir concurrencia masiva en operaciones normales
	h.mu.RLock()
	dp := h.philosophers
	bb := h.buffer
	h.mu.RUnlock()

	if mode != "" {
		dp.SetMode(mode)
	}

	if actor == "philosopher" {
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err == nil && id >= 0 && id < dp.Count {
			dp.StepPhilosopher(id)
		} else {
			for i := 0; i < dp.Count; i++ {
				dp.StepPhilosopher(i)
			}
		}
	} else if action == "produce" {
		done := make(chan struct{})
		go func() {
			bb.Produce(1, int(time.Now().Unix()%1000))
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(25 * time.Millisecond):
		}
	} else if action == "consume" {
		done := make(chan struct{})
		go func() {
			bb.Consume(1)
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(25 * time.Millisecond):
		}
	} else if action == "reset" {
		// Aquí sí requerimos Lock completo (Write) para cambiar la simulación
		h.mu.Lock()
		h.buffer.StopPending() // Aborta las goroutines atrapadas del buffer anterior
		h.philosophers = syncgo.NewDiningPhilosophers(h.philosophers.Mode)
		h.buffer = syncgo.NewBoundedBuffer(5)
		h.mu.Unlock()
	}

	h.HandleSyncStatus(w, r)
}
