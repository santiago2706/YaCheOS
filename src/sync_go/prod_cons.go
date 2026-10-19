package syncgo

import (
	"fmt"
	"sync"
	"time"
)

// SyncEvent representa un evento registrado durante la simulación de sincronización.
type SyncEvent struct {
	Timestamp int64  `json:"timestamp"`
	Actor     string `json:"actor"`
	Action    string `json:"action"`
	BufferLen int    `json:"buffer_len"`
	Details   string `json:"details"`
}

// BoundedBuffer modela el buffer acotado del problema Productor-Consumidor.
type BoundedBuffer struct {
	capacity int
	buffer   []int
	bufMu    sync.Mutex
	mutex    *CountingSemaphore
	empty    *CountingSemaphore
	full     *CountingSemaphore
	events   []SyncEvent
	eventsMu sync.Mutex
	active   bool // Bandera para gestionar la cancelación de operaciones pendientes
}

// NewBoundedBuffer inicializa el buffer con capacidad definida, sus 3 semáforos y estado activo.
func NewBoundedBuffer(capacity int) *BoundedBuffer {
	return &BoundedBuffer{
		capacity: capacity,
		buffer:   make([]int, 0, capacity),
		mutex:    NewCountingSemaphore(1),
		empty:    NewCountingSemaphore(capacity),
		full:     NewCountingSemaphore(0),
		events:   make([]SyncEvent, 0),
		active:   true,
	}
}

// isActive verifica si el buffer sigue en ejecución de forma thread-safe.
func (bb *BoundedBuffer) isActive() bool {
	bb.bufMu.Lock()
	defer bb.bufMu.Unlock()
	return bb.active
}

// StopPending desactiva el buffer y libera a las goroutines en espera
// para evitar que queden colgadas en memoria tras un reset.
func (bb *BoundedBuffer) StopPending() {
	bb.bufMu.Lock()
	bb.active = false
	bb.bufMu.Unlock()

	// Enviar señales suficientes para desbloquear cualquier goroutine atrapada
	for i := 0; i < bb.capacity+10; i++ {
		bb.empty.Signal()
		bb.full.Signal()
		bb.mutex.Signal()
	}
}

func (bb *BoundedBuffer) recordEvent(actor, action, details string, bufLen int) {
	bb.eventsMu.Lock()
	defer bb.eventsMu.Unlock()
	bb.events = append(bb.events, SyncEvent{
		Timestamp: time.Now().UnixMilli(),
		Actor:     actor,
		Action:    action,
		BufferLen: bufLen,
		Details:   details,
	})
}

// lenSafe obtiene la longitud del buffer con seguridad (Thread-Safe).
func (bb *BoundedBuffer) lenSafe() int {
	bb.bufMu.Lock()
	defer bb.bufMu.Unlock()
	return len(bb.buffer)
}

// Produce inserta un elemento respetando el orden canónico de semáforos: empty -> mutex -> insert -> signal mutex -> signal full.
func (bb *BoundedBuffer) Produce(producerID int, item int) {
	if !bb.isActive() {
		return
	}
	actor := fmt.Sprintf("Productor-%d", producerID)

	// 1. Coordinación de Bloqueo por falta de espacio
	blockedEmpty := false
	bb.empty.WaitNotify(func() {
		blockedEmpty = true
		bb.recordEvent(actor, "BLOCKED", "Buffer lleno. Productor bloqueado esperando espacio.", bb.lenSafe())
	})

	if !bb.isActive() {
		bb.empty.Signal()
		return
	}

	if blockedEmpty {
		bb.recordEvent(actor, "RESUMED", "Productor reanudado tras liberarse espacio.", bb.lenSafe())
	}

	// 2. Coordinación de Bloqueo por exclusión mutua
	blockedMutex := false
	bb.mutex.WaitNotify(func() {
		blockedMutex = true
		bb.recordEvent(actor, "BLOCKED_MUTEX", "Esperando exclusión mutua (sección crítica ocupada).", bb.lenSafe())
	})

	if !bb.isActive() {
		bb.mutex.Signal()
		bb.empty.Signal()
		return
	}

	if blockedMutex {
		bb.recordEvent(actor, "RESUMED_MUTEX", "Exclusión mutua adquirida.", bb.lenSafe())
	}

	// Sección crítica
	bb.bufMu.Lock()
	bb.buffer = append(bb.buffer, item)
	currentLen := len(bb.buffer)
	bb.bufMu.Unlock()

	bb.recordEvent(actor, "PRODUCED", fmt.Sprintf("Elemento %d insertado. Ocupación: %d/%d", item, currentLen, bb.capacity), currentLen)

	// Liberar siempre mutex antes de signal full
	bb.mutex.Signal()
	bb.full.Signal()
}

// Consume extrae un elemento respetando: full -> mutex -> extract -> signal mutex -> signal empty.
func (bb *BoundedBuffer) Consume(consumerID int) int {
	if !bb.isActive() {
		return -1
	}
	actor := fmt.Sprintf("Consumidor-%d", consumerID)

	// 1. Coordinación de Bloqueo por falta de datos
	blockedFull := false
	bb.full.WaitNotify(func() {
		blockedFull = true
		bb.recordEvent(actor, "BLOCKED", "Buffer vacío. Consumidor bloqueado esperando datos.", bb.lenSafe())
	})

	if !bb.isActive() {
		bb.full.Signal()
		return -1
	}

	if blockedFull {
		bb.recordEvent(actor, "RESUMED", "Consumidor reanudado tras recibir datos nuevos.", bb.lenSafe())
	}

	// 2. Coordinación de Bloqueo por exclusión mutua
	blockedMutex := false
	bb.mutex.WaitNotify(func() {
		blockedMutex = true
		bb.recordEvent(actor, "BLOCKED_MUTEX", "Esperando exclusión mutua (sección crítica ocupada).", bb.lenSafe())
	})

	if !bb.isActive() {
		bb.mutex.Signal()
		bb.full.Signal()
		return -1
	}

	if blockedMutex {
		bb.recordEvent(actor, "RESUMED_MUTEX", "Exclusión mutua adquirida.", bb.lenSafe())
	}

	// Sección crítica
	bb.bufMu.Lock()
	item := bb.buffer[0]
	bb.buffer = bb.buffer[1:]
	currentLen := len(bb.buffer)
	bb.bufMu.Unlock()

	bb.recordEvent(actor, "CONSUMED", fmt.Sprintf("Elemento %d retirado. Ocupación: %d/%d", item, currentLen, bb.capacity), currentLen)

	// Liberar siempre mutex antes de signal empty
	bb.mutex.Signal()
	bb.empty.Signal()

	return item
}

// GetState devuelve una instantánea segura del estado actual del buffer.
func (bb *BoundedBuffer) GetState() (int, []int, int, int, int) {
	bb.bufMu.Lock()
	defer bb.bufMu.Unlock()
	bufCopy := make([]int, len(bb.buffer))
	copy(bufCopy, bb.buffer)
	return len(bufCopy), bufCopy, bb.mutex.Value(), bb.empty.Value(), bb.full.Value()
}

// GetEvents retorna los eventos registrados.
func (bb *BoundedBuffer) GetEvents() []SyncEvent {
	bb.eventsMu.Lock()
	defer bb.eventsMu.Unlock()
	res := make([]SyncEvent, len(bb.events))
	copy(res, bb.events)
	return res
}
