package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func (h *CoreHandlers) HandleProcessDemo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}
	if runtime.GOOS != "linux" {
		http.Error(w, `{"error":"La demostración fork/exec requiere Linux o WSL Ubuntu"}`, http.StatusNotImplemented)
		return
	}
	bin := h.findBinary("process_demo")
	if bin == "" {
		http.Error(w, `{"error":"Compile process_demo con make -C src/core_c"}`, http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, bin).Output()
	if err != nil || !json.Valid(output) {
		http.Error(w, `{"error":"El demostrador de procesos no produjo una traza válida"}`, http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(output)
}

func (h *CoreHandlers) HandleThreadDemo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}
	python, err := exec.LookPath("python")
	if err != nil {
		python, err = exec.LookPath("python3")
		if err != nil {
			http.Error(w, `{"error":"Python 3 no está disponible"}`, http.StatusServiceUnavailable)
			return
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, python, filepath.Join(h.baseDir, "tests_python", "thread_demo.py")).Output()
	if err != nil || !json.Valid(output) {
		http.Error(w, `{"error":"El demostrador de hilos no produjo una traza válida"}`, http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(output)
}
