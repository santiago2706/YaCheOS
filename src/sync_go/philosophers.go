package syncgo

import (
	"fmt"
	"sync"
)

type PhilosopherState string

const (
	StateThinking   PhilosopherState = "PENSANDO"
	StateHungry     PhilosopherState = "CON_HAMBRE"
	StateEating     PhilosopherState = "COMIENDO"
	StateDeadlocked PhilosopherState = "INTERBLOQUEADO"
)

// DiningPhilosophers representa la mesa de filósofos.
type DiningPhilosophers struct {
	Count      int
	States     []PhilosopherState
	Forks      []*sync.Mutex
	ForksOwner []int // -1 indica que el tenedor está libre.
	Mode       string
	Mu         sync.Mutex
	EatCount   []int
}

// NewDiningPhilosophers crea cinco filósofos y cinco tenedores.
func NewDiningPhilosophers(mode string) *DiningPhilosophers {
	if mode != "deadlock" && mode != "preventiva" {
		mode = "preventiva" // Fallback seguro
	}

	count := 5
	dp := &DiningPhilosophers{
		Count:      count,
		States:     make([]PhilosopherState, count),
		Forks:      make([]*sync.Mutex, count),
		ForksOwner: make([]int, count),
		Mode:       mode,
		EatCount:   make([]int, count),
	}

	for i := 0; i < count; i++ {
		dp.States[i] = StateThinking
		dp.Forks[i] = &sync.Mutex{}
		dp.ForksOwner[i] = -1
	}

	return dp
}

// SetMode cambia entre "deadlock" y "preventiva" limpiando recursos para evitar inconsistencias.
func (dp *DiningPhilosophers) SetMode(mode string) error {
	dp.Mu.Lock()
	defer dp.Mu.Unlock()

	if mode != "deadlock" && mode != "preventiva" {
		return fmt.Errorf("modalidad no admitida: %s", mode)
	}

	// Al cambiar de modalidad, reseteamos para no dejar recursos retenidos
	if dp.Mode != mode {
		dp.Mode = mode
		dp.resetInternal()
	}
	return nil
}

// Reset reinicia toda la simulación desde cero de forma consistente.
func (dp *DiningPhilosophers) Reset() {
	dp.Mu.Lock()
	defer dp.Mu.Unlock()
	dp.resetInternal()
}

// resetInternal libera tenedores, restablece estados y contadores (debe usarse dentro de un Lock).
func (dp *DiningPhilosophers) resetInternal() {
	for i := 0; i < dp.Count; i++ {
		dp.States[i] = StateThinking
		dp.ForksOwner[i] = -1
		dp.EatCount[i] = 0
	}
}

// StepPhilosopher avanza un paso en el estado del filósofo indicado.
func (dp *DiningPhilosophers) StepPhilosopher(id int) {
	dp.Mu.Lock()
	defer dp.Mu.Unlock()

	leftFork := id
	rightFork := (id + 1) % dp.Count

	switch dp.States[id] {
	case StateThinking:
		dp.States[id] = StateHungry

	case StateHungry:
		if dp.Mode == "deadlock" {
			// MODALIDAD DEADLOCK: Todos toman el izquierdo primero
			if dp.ForksOwner[leftFork] != id {
				// 1. Intenta tomar el izquierdo
				if dp.ForksOwner[leftFork] == -1 {
					dp.ForksOwner[leftFork] = id
				}
			} else {
				// 2. Ya tiene el izquierdo, intenta el derecho
				if dp.ForksOwner[rightFork] == -1 {
					dp.ForksOwner[rightFork] = id
					dp.States[id] = StateEating
					dp.EatCount[id]++
				} else {
					// 3. El derecho está ocupado por otro, ¡Interbloqueo!
					dp.States[id] = StateDeadlocked
				}
			}
		} else {
			// MODALIDAD PREVENTIVA: Asimetría para evitar espera circular
			first := leftFork
			second := rightFork

			// Los impares toman primero el derecho
			if id%2 != 0 {
				first = rightFork
				second = leftFork
			}

			if dp.ForksOwner[first] != id {
				// 1. Intenta tomar su primer tenedor asignado
				if dp.ForksOwner[first] == -1 {
					dp.ForksOwner[first] = id
				}
			} else {
				// 2. Ya tiene el primer tenedor, intenta el segundo
				if dp.ForksOwner[second] == -1 {
					dp.ForksOwner[second] = id
					dp.States[id] = StateEating
					dp.EatCount[id]++
				}
				// NOTA: Si el segundo está ocupado, simplemente se mantiene en StateHungry.
				// Por la asimetría de adquisición, se garantiza matemáticamente que
				// alguien comerá y soltará los tenedores, evitando el interbloqueo.
			}
		}

	case StateEating:
		// Libera ambos recursos
		if dp.ForksOwner[leftFork] == id {
			dp.ForksOwner[leftFork] = -1
		}
		if dp.ForksOwner[rightFork] == id {
			dp.ForksOwner[rightFork] = -1
		}

		dp.States[id] = StateThinking

	case StateDeadlocked:
		// Fallback de recuperación si se forzó un cambio de estado externo
		if dp.Mode == "preventiva" {
			if dp.ForksOwner[leftFork] == id {
				dp.ForksOwner[leftFork] = -1
			}
			if dp.ForksOwner[rightFork] == id {
				dp.ForksOwner[rightFork] = -1
			}
			dp.States[id] = StateThinking
		}
	}
}

// CheckDeadlock comprueba si existe un interbloqueo utilizando un Grafo
// de Asignación de Recursos (RAG). Se unifica para evitar resultados contradictorios.
func (dp *DiningPhilosophers) CheckDeadlock() (bool, []int) {
	// Ya no bloqueamos el Mutex aquí.
	// DetectRAGCycle se encarga de la exclusión mutua localmente.
	report := DetectRAGCycle(dp)
	return report.Deadlock, report.Cycle
}

// PhilosopherSnapshot representa una instantánea segura del estado en un instante t.
type PhilosopherSnapshot struct {
	Count      int
	States     []PhilosopherState
	ForksOwner []int
	Mode       string
	EatCount   []int
	Deadlock   bool
	Cycle      []int
}

// GetSnapshotAndReport obtiene una copia de los arreglos bajo mutex y luego
// construye el reporte RAG analizando la copia para garantizar consistencia absoluta.
func (dp *DiningPhilosophers) GetSnapshotAndReport() (PhilosopherSnapshot, DeadlockReport) {
	dp.Mu.Lock()

	// 1. Copiamos el estado para evitar bloqueos y race-conditions en el JSON Encoder
	dpCopy := &DiningPhilosophers{
		Count:      dp.Count,
		Mode:       dp.Mode,
		States:     make([]PhilosopherState, dp.Count),
		ForksOwner: make([]int, dp.Count),
		EatCount:   make([]int, dp.Count),
	}
	copy(dpCopy.States, dp.States)
	copy(dpCopy.ForksOwner, dp.ForksOwner)
	copy(dpCopy.EatCount, dp.EatCount)

	dp.Mu.Unlock() // Liberamos INMEDIATAMENTE para que la simulación continúe

	// 2. Ejecutamos el motor de detección de ciclos RAG sobre la copia congelada
	report := DetectRAGCycle(dpCopy)

	snap := PhilosopherSnapshot{
		Count:      dpCopy.Count,
		States:     dpCopy.States,
		ForksOwner: dpCopy.ForksOwner,
		Mode:       dpCopy.Mode,
		EatCount:   dpCopy.EatCount,
		Deadlock:   report.Deadlock,
		Cycle:      report.Cycle,
	}

	return snap, report
}
