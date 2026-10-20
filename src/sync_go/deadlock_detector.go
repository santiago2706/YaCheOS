package syncgo

import "sort"

// DeadlockReport contiene la estructura de alerta requerida por la tarea.
type DeadlockReport struct {
	Deadlock bool  `json:"deadlock"`
	Cycle    []int `json:"cycle"`
}

// DetectRAGCycle evalúa si existe un ciclo en el Grafo de Asignación de Recursos (RAG).
func DetectRAGCycle(dp *DiningPhilosophers) DeadlockReport {
	// 1. Manejo de entradas vacías o inválidas
	if dp == nil {
		return DeadlockReport{Cycle: make([]int, 0)}
	}

	dp.Mu.Lock()
	defer dp.Mu.Unlock()

	if dp.Count <= 0 || len(dp.ForksOwner) != dp.Count || len(dp.States) != dp.Count {
		return DeadlockReport{Cycle: make([]int, 0)}
	}

	// 2. Construcción del Grafo (RAG)
	// Nodos 0 a Count-1: Procesos (Filósofos)
	// Nodos Count a 2*Count-1: Recursos (Tenedores de 1 instancia)
	totalNodes := dp.Count * 2
	adj := make([][]int, totalNodes)

	// Aristas de Asignación: Recurso -> Proceso
	for r := 0; r < dp.Count; r++ {
		owner := dp.ForksOwner[r]
		if owner != -1 {
			adj[r+dp.Count] = append(adj[r+dp.Count], owner)
		}
	}

	// Aristas de Solicitud: Proceso -> Recurso
	for p := 0; p < dp.Count; p++ {
		if dp.States[p] == StateThinking || dp.States[p] == StateEating {
			continue // No están solicitando nada
		}

		leftFork := p
		rightFork := (p + 1) % dp.Count
		requestedFork := -1

		if dp.Mode == "deadlock" {
			if dp.ForksOwner[leftFork] != p {
				requestedFork = leftFork
			} else if dp.ForksOwner[rightFork] != p {
				requestedFork = rightFork
			}
		} else { // modalidad "preventiva"
			first, second := leftFork, rightFork
			if p%2 != 0 {
				first, second = rightFork, leftFork
			}
			if dp.ForksOwner[first] != p {
				requestedFork = first
			} else if dp.ForksOwner[second] != p {
				requestedFork = second
			}
		}

		if requestedFork != -1 {
			adj[p] = append(adj[p], requestedFork+dp.Count)
		}
	}

	// 3. Búsqueda de ciclos mediante DFS
	visited := make([]bool, totalNodes)
	recStack := make([]bool, totalNodes)
	path := make([]int, 0)

	var cycleStart int = -1
	var cycleNodes []int

	var dfs func(u int) bool
	dfs = func(u int) bool {
		visited[u] = true
		recStack[u] = true
		path = append(path, u)

		for _, v := range adj[u] {
			if !visited[v] {
				if dfs(v) {
					return true
				}
			} else if recStack[v] {
				// Ciclo detectado: encontramos un back-edge al stack actual
				cycleStart = v
				return true
			}
		}

		recStack[u] = false
		path = path[:len(path)-1]
		return false
	}

	// Recorrer todos los nodos (útil si hay subgrafos desconectados)
	for i := 0; i < totalNodes; i++ {
		if !visited[i] {
			if dfs(i) {
				// Extraer PIDs involucrados en el ciclo
				inCycle := false
				for _, node := range path {
					if node == cycleStart {
						inCycle = true
					}
					if inCycle && node < dp.Count { // Filtrar solo los procesos (Filósofos)
						cycleNodes = append(cycleNodes, node+1) // Formato Base-1 para UI
					}
				}
				// Ordenar para mantener consistencia estricta en el arreglo retornado (ej: [1, 2, 3, 4, 5])
				sort.Ints(cycleNodes)

				return DeadlockReport{
					Deadlock: true,
					Cycle:    cycleNodes,
				}
			}
		}
	}

	return DeadlockReport{
		Deadlock: false,
		Cycle:    make([]int, 0),
	}
}
