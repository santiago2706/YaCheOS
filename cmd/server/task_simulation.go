package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type taskProcess struct {
	PID       int    `json:"pid"`
	Name      string `json:"name"`
	Arrival   int    `json:"arrival_time"`
	Burst     int    `json:"burst_time"`
	Priority  int    `json:"priority"`
	MemoryKB  int    `json:"memory_kb"`
	Remaining int    `json:"remaining"`
	State     string `json:"state"`
}

type taskSegment struct {
	PID      int    `json:"pid"`
	Name     string `json:"name"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Duration int    `json:"duration"`
}

type taskSchedule struct {
	Timeline          []taskSegment `json:"timeline"`
	AverageWaiting    float64       `json:"average_waiting_time"`
	AverageTurnaround float64       `json:"average_turnaround_time"`
	AverageResponse   float64       `json:"average_response_time"`
}

type taskMemory struct {
	TotalSize int `json:"total_size_kb"`
	TotalFree int `json:"total_free_kb"`
	Blocks    []struct {
		Start int  `json:"start"`
		Size  int  `json:"size"`
		PID   int  `json:"pid"`
		Free  bool `json:"free"`
	} `json:"blocks"`
}

type taskSimulation struct {
	Algorithm string        `json:"algorithm"`
	Tick      int           `json:"tick"`
	Modified  bool          `json:"modified"`
	Processes []taskProcess `json:"processes"`
	Timeline  []taskSegment `json:"timeline"`
	Metrics   taskSchedule  `json:"metrics"`
	Memory    taskMemory    `json:"memory"`
}

func (h *CoreHandlers) computeTaskSchedule(ctx context.Context, algorithm string, processes []taskProcess) (taskSchedule, error) {
	bin := h.findBinary("scheduler")
	if bin == "" {
		return taskSchedule{}, errors.New("compile bin/scheduler del motor C")
	}
	file, err := os.CreateTemp("", "yakcheos_task_*.json")
	if err != nil {
		return taskSchedule{}, err
	}
	defer os.Remove(file.Name())
	if err := json.NewEncoder(file).Encode(processes); err != nil {
		file.Close()
		return taskSchedule{}, err
	}
	if err := file.Close(); err != nil {
		return taskSchedule{}, err
	}
	cmd := exec.CommandContext(ctx, bin, "--algo", algorithm, "--file", file.Name(), "--quantum", "4")
	output, err := cmd.Output()
	if err != nil {
		return taskSchedule{}, fmt.Errorf("planificador C: %w", err)
	}
	var result taskSchedule
	if err := json.Unmarshal(output, &result); err != nil {
		return taskSchedule{}, err
	}
	if result.Timeline == nil {
		return taskSchedule{}, errors.New("planificador C no devolvió timeline")
	}
	return result, nil
}

func (h *CoreHandlers) computeTaskMemory(ctx context.Context, processes []taskProcess) (taskMemory, error) {
	bin := h.findBinary("memory_cli")
	if bin == "" {
		return taskMemory{}, errors.New("compile bin/memory_cli del motor C")
	}
	var released []string
	for _, p := range processes {
		if p.State == "TERMINATED" {
			released = append(released, strconv.Itoa(p.PID))
		}
	}
	cmd := exec.CommandContext(ctx, bin, "--mode", "process-map", "--file", filepath.Join(h.baseDir, "data", "procesos_semana4.json"), "--released", strings.Join(released, ","))
	output, err := cmd.Output()
	if err != nil {
		return taskMemory{}, fmt.Errorf("asignador C: %w", err)
	}
	var result taskMemory
	if err := json.Unmarshal(output, &result); err != nil {
		return taskMemory{}, err
	}
	if result.TotalSize != 1024 || len(result.Blocks) == 0 {
		return taskMemory{}, errors.New("mapa C incompleto")
	}
	return result, nil
}

func (h *CoreHandlers) resetTask(ctx context.Context, algorithm string) error {
	if algorithm != "FCFS" && algorithm != "SJF" && algorithm != "RR" && algorithm != "PRIORIDAD" {
		return errors.New("algoritmo de CPU inválido")
	}
	data, err := os.ReadFile(filepath.Join(h.baseDir, "data", "procesos_semana4.json"))
	if err != nil {
		return err
	}
	var base []taskProcess
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}
	if len(base) == 0 {
		return errors.New("no hay procesos en el contrato")
	}
	for i := range base {
		base[i].Remaining = base[i].Burst
		base[i].State = "READY"
		if base[i].Arrival > 0 {
			base[i].State = "WAITING"
		}
	}
	schedule, err := h.computeTaskSchedule(ctx, algorithm, base)
	if err != nil {
		return err
	}
	memory, err := h.computeTaskMemory(ctx, base)
	if err != nil {
		return err
	}
	h.task = &taskSimulation{Algorithm: algorithm, Processes: base, Timeline: schedule.Timeline, Metrics: schedule, Memory: memory}
	return nil
}

func (h *CoreHandlers) stepTask(ctx context.Context) error {
	current := h.task
	state := *current
	state.Processes = append([]taskProcess(nil), current.Processes...)
	remaining := false
	for _, p := range state.Processes {
		if p.State != "TERMINATED" {
			remaining = true
			break
		}
	}
	if !remaining {
		return nil
	}
	activePID := 0
	for _, seg := range state.Timeline {
		if seg.Start <= state.Tick && state.Tick < seg.End {
			activePID = seg.PID
			break
		}
	}
	state.Tick++
	memoryChanged := false
	for i := range state.Processes {
		p := &state.Processes[i]
		if p.State == "TERMINATED" {
			continue
		}
		if p.Arrival > state.Tick {
			p.State = "WAITING"
			continue
		}
		if p.PID == activePID {
			p.Remaining--
			if p.Remaining <= 0 {
				p.Remaining = 0
				p.State = "TERMINATED"
				memoryChanged = true
			} else {
				p.State = "RUNNING"
			}
		} else {
			p.State = "READY"
		}
	}
	if memoryChanged {
		memory, err := h.computeTaskMemory(ctx, state.Processes)
		if err != nil {
			return err
		}
		state.Memory = memory
	}
	h.task = &state
	return nil
}

func (h *CoreHandlers) killTask(ctx context.Context, pid int) error {
	current := h.task
	state := *current
	state.Processes = append([]taskProcess(nil), current.Processes...)
	index := -1
	for i := range state.Processes {
		if state.Processes[i].PID == pid {
			index = i
			break
		}
	}
	if index < 0 || state.Processes[index].State == "TERMINATED" {
		return errors.New("PID ausente o ya terminado")
	}
	state.Processes[index].State = "TERMINATED"
	state.Processes[index].Remaining = 0
	state.Modified = true
	var survivors []taskProcess
	for _, p := range state.Processes {
		if p.State != "TERMINATED" {
			p.Burst = p.Remaining
			if p.Arrival < state.Tick {
				p.Arrival = state.Tick
			}
			survivors = append(survivors, p)
		}
	}
	var future taskSchedule
	var err error
	if len(survivors) > 0 {
		future, err = h.computeTaskSchedule(ctx, state.Algorithm, survivors)
		if err != nil {
			return err
		}
	}
	var past []taskSegment
	for _, seg := range state.Timeline {
		if seg.Start >= state.Tick {
			break
		}
		if seg.End > state.Tick {
			seg.End = state.Tick
			seg.Duration = seg.End - seg.Start
		}
		past = append(past, seg)
	}
	state.Timeline = append(past, future.Timeline...)
	state.Metrics = future
	memory, err := h.computeTaskMemory(ctx, state.Processes)
	if err != nil {
		return err
	}
	state.Memory = memory
	h.task = &state
	return nil
}

func (h *CoreHandlers) HandleTaskSimulation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, `{"error":"Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	h.taskMu.Lock()
	defer h.taskMu.Unlock()
	if h.task == nil {
		if err := h.resetTask(ctx, "FCFS"); err != nil {
			http.Error(w, `{"error":"No se pudo iniciar la simulación C"}`, http.StatusServiceUnavailable)
			return
		}
	}
	if r.Method == http.MethodPost {
		var action struct {
			Action    string `json:"action"`
			Algorithm string `json:"algorithm"`
			PID       int    `json:"pid"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&action); err != nil {
			http.Error(w, `{"error":"Acción JSON inválida"}`, http.StatusBadRequest)
			return
		}
		var err error
		switch action.Action {
		case "step":
			err = h.stepTask(ctx)
		case "reset":
			if action.Algorithm == "" {
				action.Algorithm = h.task.Algorithm
			}
			err = h.resetTask(ctx, action.Algorithm)
		case "kill":
			err = h.killTask(ctx, action.PID)
		default:
			err = errors.New("acción desconocida")
		}
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusBadRequest)
			return
		}
	}
	sort.Slice(h.task.Processes, func(i, j int) bool { return h.task.Processes[i].PID < h.task.Processes[j].PID })
	_ = json.NewEncoder(w).Encode(h.task)
}
