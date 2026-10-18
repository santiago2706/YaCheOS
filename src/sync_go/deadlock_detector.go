package syncgo

// DeadlockReport contiene la estructura de alerta requerida por la tarea T3.4.
type DeadlockReport struct {
	Deadlock bool  `json:"deadlock"`
	Cycle    []int `json:"cycle"`
}

// DetectRAGCycle evalúa si existe un ciclo en el Grafo de Asignación de Recursos para tenedores/procesos.
func DetectRAGCycle(dp *DiningPhilosophers) DeadlockReport {
	dp.Mu.Lock()
	defer dp.Mu.Unlock()

	// En el caso clásico de 5 filósofos, si cada filósofo i retiene el recurso i y espera el (i+1)%5, hay ciclo [1,2,3,4,5]
	cycle := make([]int, 0)
	allHoldingLeft := true

	for i := 0; i < dp.Count; i++ {
		if dp.ForksOwner[i] == i && dp.States[i] == StateDeadlocked {
			cycle = append(cycle, i+1)
		} else {
			allHoldingLeft = false
		}
	}

	if allHoldingLeft && len(cycle) == dp.Count {
		return DeadlockReport{
			Deadlock: true,
			Cycle:    cycle,
		}
	}

	return DeadlockReport{
		Deadlock: false,
		Cycle:    []int{},
	}
}
