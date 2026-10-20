package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var beladyReferences = []int{1, 2, 3, 4, 1, 2, 5, 1, 2, 3, 4, 5}

type replacementRequest struct {
	References []int `json:"references"`
	Frames     int   `json:"frames"`
}

type replacementResult struct {
	Algorithm  string `json:"algorithm"`
	FrameCount int    `json:"frame_count"`
	Hits       int    `json:"hits"`
	Faults     int    `json:"faults"`
	Trace      []struct {
		Index   int   `json:"index"`
		Page    int   `json:"page"`
		Frames  []int `json:"frames"`
		Hit     bool  `json:"hit"`
		Fault   bool  `json:"fault"`
		Evicted *int  `json:"evicted"`
	} `json:"trace"`
}

func (h *CoreHandlers) runReplacement(ctx context.Context, request replacementRequest) (map[string]replacementResult, error) {
	python, err := exec.LookPath("python")
	if err != nil {
		python, err = exec.LookPath("python3")
		if err != nil {
			return nil, errors.New("Python 3 no está disponible en PATH")
		}
	}
	input, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, python, filepath.Join(h.baseDir, "tests_python", "page_replacement.py"))
	cmd.Stdin = bytes.NewReader(input)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("falló el reemplazo en Python: %w", err)
	}
	var result map[string]replacementResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("respuesta Python inválida: %w", err)
	}
	for _, name := range []string{"FIFO", "LRU", "OPTIMAL"} {
		if item, ok := result[name]; !ok || item.Hits+item.Faults != len(request.References) || len(item.Trace) != len(request.References) {
			return nil, fmt.Errorf("traza %s incompleta", name)
		}
	}
	return result, nil
}

// HandlePageReplacementIntegrated publica exactamente las trazas calculadas por el módulo Python.
func (h *CoreHandlers) HandlePageReplacementIntegrated(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, `{"error":"Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}
	req := replacementRequest{References: append([]int(nil), beladyReferences...), Frames: 3}
	if r.Method == http.MethodPost {
		decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, `{"error":"JSON de reemplazo inválido"}`, http.StatusBadRequest)
			return
		}
		if decoder.Decode(new(any)) != io.EOF {
			http.Error(w, `{"error":"Solo se permite un objeto JSON"}`, http.StatusBadRequest)
			return
		}
	} else {
		if raw := r.URL.Query().Get("frames"); raw != "" {
			frames, err := strconv.Atoi(raw)
			if err != nil {
				http.Error(w, `{"error":"frames debe ser un entero"}`, http.StatusBadRequest)
				return
			}
			req.Frames = frames
		}
		if raw := r.URL.Query().Get("references"); raw != "" {
			req.References = nil
			for _, value := range strings.Split(raw, ",") {
				page, err := strconv.Atoi(strings.TrimSpace(value))
				if err != nil {
					http.Error(w, `{"error":"references debe ser una lista de enteros"}`, http.StatusBadRequest)
					return
				}
				req.References = append(req.References, page)
			}
		}
	}
	if req.Frames < 1 || req.Frames > 256 || len(req.References) < 1 || len(req.References) > 10000 {
		http.Error(w, `{"error":"Se requieren 1-256 marcos y 1-10000 referencias"}`, http.StatusBadRequest)
		return
	}
	for _, page := range req.References {
		if page < 0 {
			http.Error(w, `{"error":"Las páginas deben ser no negativas"}`, http.StatusBadRequest)
			return
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	result, err := h.runReplacement(ctx, req)
	if err != nil {
		http.Error(w, `{"error":"No se pudo ejecutar el reemplazo en Python"}`, http.StatusServiceUnavailable)
		return
	}
	response := map[string]any{
		"references": req.References,
		"frames":     req.Frames,
		"fifo":       result["FIFO"],
		"lru":        result["LRU"],
		"optimal":    result["OPTIMAL"],
	}
	if len(req.References) == len(beladyReferences) {
		match := true
		for i, page := range req.References {
			if page != beladyReferences[i] {
				match = false
				break
			}
		}
		if match {
			three := result["FIFO"].Faults
			four := result["FIFO"].Faults
			if req.Frames != 3 {
				other, err := h.runReplacement(ctx, replacementRequest{References: beladyReferences, Frames: 3})
				if err != nil {
					http.Error(w, `{"error":"Falló la comparación de Belady"}`, http.StatusServiceUnavailable)
					return
				}
				three = other["FIFO"].Faults
			}
			if req.Frames != 4 {
				other, err := h.runReplacement(ctx, replacementRequest{References: beladyReferences, Frames: 4})
				if err != nil {
					http.Error(w, `{"error":"Falló la comparación de Belady"}`, http.StatusServiceUnavailable)
					return
				}
				four = other["FIFO"].Faults
			}
			response["belady_anomaly"] = map[string]any{"fifo_3_frames_faults": three, "fifo_4_frames_faults": four, "anomaly_confirmed": four > three}
		}
	}
	_ = json.NewEncoder(w).Encode(response)
}
