package syncgo

import "sync"

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

// SetMode cambia entre "deadlock" y "preventiva".
func (dp *DiningPhilosophers) SetMode(mode string) {
	dp.Mu.Lock()
	defer dp.Mu.Unlock()

	dp.Mode = mode
}

// StepPhilosopher avanza un paso en el estado del filósofo.
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
			if dp.ForksOwner[leftFork] == -1 ||
				dp.ForksOwner[leftFork] == id {
				dp.ForksOwner[leftFork] = id

				if dp.ForksOwner[rightFork] == -1 {
					dp.ForksOwner[rightFork] = id
					dp.States[id] = StateEating
					dp.EatCount[id]++
				} else {
					dp.States[id] = StateDeadlocked
				}
			}
		} else {
			first := leftFork
			second := rightFork

			if id%2 != 0 {
				first = rightFork
				second = leftFork
			}

			if (dp.ForksOwner[first] == -1 ||
				dp.ForksOwner[first] == id) &&
				dp.ForksOwner[second] == -1 {
				dp.ForksOwner[first] = id
				dp.ForksOwner[second] = id
				dp.States[id] = StateEating
				dp.EatCount[id]++
			}
		}

	case StateEating:
		if dp.ForksOwner[leftFork] == id {
			dp.ForksOwner[leftFork] = -1
		}
		if dp.ForksOwner[rightFork] == id {
			dp.ForksOwner[rightFork] = -1
		}

		dp.States[id] = StateThinking

	case StateDeadlocked:
		if dp.Mode == "preventiva" {
			if dp.ForksOwner[leftFork] == id {
				dp.ForksOwner[leftFork] = -1
			}

			dp.States[id] = StateThinking
		}
	}
}

// CheckDeadlock comprueba si todos los filósofos están bloqueados.
func (dp *DiningPhilosophers) CheckDeadlock() (bool, []int) {
	dp.Mu.Lock()
	defer dp.Mu.Unlock()

	blocked := make([]int, 0)

	for i := 0; i < dp.Count; i++ {
		if dp.States[i] == StateDeadlocked {
			blocked = append(blocked, i+1)
		}
	}

	if len(blocked) == dp.Count {
		return true, blocked
	}

	return false, blocked
}
