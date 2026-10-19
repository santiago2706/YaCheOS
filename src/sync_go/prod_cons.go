package syncgo

import (
	"fmt"
	"sync"
	"time"
)

type SyncEvent struct {
	Timestamp int64  `json:"timestamp"`
	Actor     string `json:"actor"`
	Action    string `json:"action"`
	BufferLen int    `json:"buffer_len"`
	Details   string `json:"details"`
}

type BoundedBuffer struct {
	capacity int
	buffer   []int
	bufMu    sync.Mutex
	mutex    *CountingSemaphore
	empty    *CountingSemaphore
	full     *CountingSemaphore
	events   []SyncEvent
	eventsMu sync.Mutex
}

func NewBoundedBuffer(capacity int) *BoundedBuffer {
	return &BoundedBuffer{
		capacity: capacity,
		buffer:   make([]int, 0, capacity),
		mutex:    NewCountingSemaphore(1),
		empty:    NewCountingSemaphore(capacity),
		full:     NewCountingSemaphore(0),
		events:   make([]SyncEvent, 0),
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

// lenSafe obtiene la longitud del buffer con seguridad (Thread-Safe)
func (bb *BoundedBuffer) lenSafe() int {
	bb.bufMu.Lock()
	defer bb.bufMu.Unlock()
	return len(bb.buffer)
}

// Produce inserta un elemento respetando el orden canónico de semáforos.
func (bb *BoundedBuffer) Produce(producerID int, item int) {
	actor := fmt.Sprintf("Productor-%d", producerID)

	// 1. Coordinación de Bloqueo por falta de espacio
	blockedEmpty := false
	bb.empty.WaitNotify(func() {
		blockedEmpty = true
		bb.recordEvent(actor, "BLOCKED", "Buffer lleno. Productor bloqueado esperando espacio.", bb.lenSafe())
	})
	if blockedEmpty {
		bb.recordEvent(actor, "RESUMED", "Productor reanudado tras liberarse espacio.", bb.lenSafe())
	}

	// 2. Coordinación de Bloqueo por exclusión mutua
	blockedMutex := false
	bb.mutex.WaitNotify(func() {
		blockedMutex = true
		bb.recordEvent(actor, "BLOCKED_MUTEX", "Esperando exclusión mutua (sección crítica ocupada).", bb.lenSafe())
	})
	if blockedMutex {
		bb.recordEvent(actor, "RESUMED_MUTEX", "Exclusión mutua adquirida.", bb.lenSafe())
	}

	// Sección crítica
	bb.bufMu.Lock()
	bb.buffer = append(bb.buffer, item)
	currentLen := len(bb.buffer)
	bb.bufMu.Unlock()

	bb.recordEvent(actor, "PRODUCED", fmt.Sprintf("Elemento %d insertado. Ocupación: %d/%d", item, currentLen, bb.capacity), currentLen)

	// Regla clave: Liberar SIEMPRE el mutex antes de enviar la señal al consumidor
	bb.mutex.Signal()
	bb.full.Signal()
}

// Consume extrae un elemento respetando el orden canónico de semáforos.
func (bb *BoundedBuffer) Consume(consumerID int) int {
	actor := fmt.Sprintf("Consumidor-%d", consumerID)

	// 1. Coordinación de Bloqueo por falta de datos
	blockedFull := false
	bb.full.WaitNotify(func() {
		blockedFull = true
		bb.recordEvent(actor, "BLOCKED", "Buffer vacío. Consumidor bloqueado esperando datos.", bb.lenSafe())
	})
	if blockedFull {
		bb.recordEvent(actor, "RESUMED", "Consumidor reanudado tras recibir datos nuevos.", bb.lenSafe())
	}

	// 2. Coordinación de Bloqueo por exclusión mutua
	blockedMutex := false
	bb.mutex.WaitNotify(func() {
		blockedMutex = true
		bb.recordEvent(actor, "BLOCKED_MUTEX", "Esperando exclusión mutua (sección crítica ocupada).", bb.lenSafe())
	})
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

	// Regla clave: Liberar SIEMPRE el mutex antes de enviar la señal al productor
	bb.mutex.Signal()
	bb.empty.Signal()

	return item
}

func (bb *BoundedBuffer) GetState() (int, []int, int, int, int) {
	bb.bufMu.Lock()
	defer bb.bufMu.Unlock()
	bufCopy := make([]int, len(bb.buffer))
	copy(bufCopy, bb.buffer)
	return len(bufCopy), bufCopy, bb.mutex.Value(), bb.empty.Value(), bb.full.Value()
}

func (bb *BoundedBuffer) GetEvents() []SyncEvent {
	bb.eventsMu.Lock()
	defer bb.eventsMu.Unlock()
	res := make([]SyncEvent, len(bb.events))
	copy(res, bb.events)
	return res
}
