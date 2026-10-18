package syncgo

import (
	"sync"
)

// CountingSemaphore implementa un semáforo contador atómico con canales y exclusión mutua.
type CountingSemaphore struct {
	permits int
	mu      sync.Mutex
	cond    *sync.Cond
}

// NewCountingSemaphore crea un nuevo semáforo contador con permisos iniciales.
func NewCountingSemaphore(initialPermits int) *CountingSemaphore {
	cs := &CountingSemaphore{
		permits: initialPermits,
	}
	cs.cond = sync.NewCond(&cs.mu)
	return cs
}

// Wait / P adquiere un permiso. Si no hay disponibles, bloquea hasta que otro hilo lo libere.
func (cs *CountingSemaphore) Wait() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	for cs.permits <= 0 {
		cs.cond.Wait()
	}
	cs.permits--
}

// P es alias estándar para Wait
func (cs *CountingSemaphore) P() {
	cs.Wait()
}

// Signal / V libera un permiso e incrementa el contador, despertando a un hilo en espera.
func (cs *CountingSemaphore) Signal() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.permits++
	cs.cond.Signal()
}

// V es alias estándar para Signal
func (cs *CountingSemaphore) V() {
	cs.Signal()
}

// Value retorna el valor actual del contador de forma segura.
func (cs *CountingSemaphore) Value() int {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.permits
}
