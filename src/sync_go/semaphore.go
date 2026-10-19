package syncgo

import (
	"math"
	"sync/atomic"
)

// CountingSemaphore implementa un semáforo contador rápido y seguro.
// Utiliza un tipo atómico para el estado y un canal para la gestión de bloqueos.
type CountingSemaphore struct {
	permits atomic.Int64
	waiters chan struct{}
}

// NewCountingSemaphore crea un nuevo semáforo contador con permisos iniciales.
func NewCountingSemaphore(initialPermits int) *CountingSemaphore {
	if initialPermits < 0 {
		panic("syncgo: los permisos iniciales no pueden ser negativos")
	}

	cs := &CountingSemaphore{
		// Se usa un canal con capacidad máxima para evitar bloqueos en Signal().
		// Dado que el tipo es struct{}, Go no reserva memoria en el heap
		// para el buffer interno, haciéndolo extremadamente eficiente.
		waiters: make(chan struct{}, math.MaxInt32),
	}
	cs.permits.Store(int64(initialPermits))

	return cs
}

// Wait / P adquiere un permiso. Si no hay disponibles, bloquea hasta que otro hilo lo libere.
// Garantiza que la adquisición del permiso sea una operación atómica libre de condiciones de carrera.
func (cs *CountingSemaphore) Wait() {
	// Reducimos el contador atómicamente. Si el nuevo valor es menor a 0,
	// significa que los permisos se agotaron y el hilo debe encolarse.
	if cs.permits.Add(-1) < 0 {
		<-cs.waiters
	}
}

// WaitNotify adquiere un permiso. Si no hay disponibles, ejecuta el callback onBlock
// justo antes de que el hilo se bloquee, garantizando la coordinación atómica.
func (cs *CountingSemaphore) WaitNotify(onBlock func()) {
	// Restamos atómicamente. Si es menor a 0, sabemos con 100% de certeza que bloquearemos.
	if cs.permits.Add(-1) < 0 {
		if onBlock != nil {
			onBlock() // Registramos el timestamp/evento justo antes de dormir
		}
		<-cs.waiters // Suspensión real de la goroutine
	}
}

// P es un alias estándar para Wait.
func (cs *CountingSemaphore) P() {
	cs.Wait()
}

// Signal / V libera un permiso e incrementa el contador, despertando a un hilo en espera.
func (cs *CountingSemaphore) Signal() {
	// Incrementamos el contador atómicamente. Si el valor resultante es <= 0,
	// sabemos con certeza que hay un hilo esperando (o a punto de hacerlo) en el canal.
	if cs.permits.Add(1) <= 0 {
		cs.waiters <- struct{}{}
	}
}

// V es un alias estándar para Signal.
func (cs *CountingSemaphore) V() {
	cs.Signal()
}

// Value retorna el valor actual del contador de forma segura.
// Si el valor es negativo, su valor absoluto representa la cantidad de hilos en espera.
func (cs *CountingSemaphore) Value() int {
	return int(cs.permits.Load())
}
