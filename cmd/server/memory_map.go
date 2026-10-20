package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// HandleMemoryMap reconstruye el estado de asignación en C a partir de los procesos
// del contrato y los PID liberados durante esta simulación.
func (h *CoreHandlers) HandleMemoryMap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}
	released := r.URL.Query().Get("released")
	if len(released) > 1024 {
		http.Error(w, `{"error":"Lista de PID demasiado larga"}`, http.StatusBadRequest)
		return
	}
	if released != "" {
		for _, value := range strings.Split(released, ",") {
			pid, err := strconv.Atoi(value)
			if err != nil || pid <= 0 {
				http.Error(w, `{"error":"released debe contener PID positivos"}`, http.StatusBadRequest)
				return
			}
		}
	}
	bin := h.findBinary("memory_cli")
	if bin == "" {
		http.Error(w, `{"error":"Compile memory_cli del motor C"}`, http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "--mode", "process-map", "--file", filepath.Join(h.baseDir, "data", "procesos_semana4.json"), "--released", released)
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, `{"error":"No se pudo generar el mapa de memoria C"}`, http.StatusBadRequest)
		return
	}
	var result any
	if err := json.Unmarshal(output, &result); err != nil {
		http.Error(w, `{"error":"Mapa de memoria C inválido"}`, http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(output)
}
