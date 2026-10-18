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

// BoundedBuffer modela el buffer acotado del problema Productor-Consumidor (Semana 5).
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

// NewBoundedBuffer inicializa el buffer con capacidad definida y sus 3 semáforos.
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

// Produce inserta un elemento respetando el orden canónico de semáforos: empty -> mutex -> insert -> signal mutex -> signal full.
func (bb *BoundedBuffer) Produce(producerID int, item int) {
	actor := fmt.Sprintf("Productor-%d", producerID)
	bb.recordEvent(actor, "WAIT_EMPTY", "Esperando slot libre en buffer", -1)
	bb.empty.Wait()

	bb.recordEvent(actor, "WAIT_MUTEX", "Esperando exclusión mutua (mutex)", -1)
	bb.mutex.Wait()

	// Sección crítica
	bb.bufMu.Lock()
	bb.buffer = append(bb.buffer, item)
	currentLen := len(bb.buffer)
	bb.bufMu.Unlock()

	bb.recordEvent(actor, "PRODUCED", fmt.Sprintf("Elemento %d insertado. Ocupación: %d/%d", item, currentLen, bb.capacity), currentLen)

	bb.mutex.Signal()
	bb.full.Signal()
}

// Consume extrae un elemento respetando: full -> mutex -> extract -> signal mutex -> signal empty.
func (bb *BoundedBuffer) Consume(consumerID int) int {
	actor := fmt.Sprintf("Consumidor-%d", consumerID)
	bb.recordEvent(actor, "WAIT_FULL", "Esperando elemento disponible en buffer", -1)
	bb.full.Wait()

	bb.recordEvent(actor, "WAIT_MUTEX", "Esperando exclusión mutua (mutex)", -1)
	bb.mutex.Wait()

	// Sección crítica
	bb.bufMu.Lock()
	item := bb.buffer[0]
	bb.buffer = bb.buffer[1:]
	currentLen := len(bb.buffer)
	bb.bufMu.Unlock()

	bb.recordEvent(actor, "CONSUMED", fmt.Sprintf("Elemento %d retirado. Ocupación: %d/%d", item, currentLen, bb.capacity), currentLen)

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
