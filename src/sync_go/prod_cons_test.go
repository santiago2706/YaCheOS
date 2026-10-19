package syncgo

import (
	"sync"
	"testing"
	"time"
)

// TestBoundedBuffer_Integrity verifica la concurrencia masiva:
// 2 productores y 2 consumidores procesando 20 elementos exactos.
func TestBoundedBuffer_Integrity(t *testing.T) {
	const capacity = 5
	const totalItems = 20
	const itemsPerActor = totalItems / 2

	bb := NewBoundedBuffer(capacity)

	var wgProducers sync.WaitGroup
	var wgConsumers sync.WaitGroup

	// Usamos un mapa concurrente para registrar los elementos consumidos
	// y detectar duplicados o pérdidas.
	var mu sync.Mutex
	consumedItems := make(map[int]int)

	// Iniciar 2 Productores
	for p := 1; p <= 2; p++ {
		wgProducers.Add(1)
		go func(pID int) {
			defer wgProducers.Done()
			for i := 1; i <= itemsPerActor; i++ {
				// Genera IDs únicos: Productor 1 (1-10), Productor 2 (11-20)
				item := (pID-1)*itemsPerActor + i
				bb.Produce(pID, item)
				time.Sleep(2 * time.Millisecond) // Forzar cambios de contexto
			}
		}(p)
	}

	// Iniciar 2 Consumidores
	for c := 1; c <= 2; c++ {
		wgConsumers.Add(1)
		go func(cID int) {
			defer wgConsumers.Done()
			for i := 1; i <= itemsPerActor; i++ {
				item := bb.Consume(cID)

				mu.Lock()
				consumedItems[item]++
				mu.Unlock()

				time.Sleep(3 * time.Millisecond) // Forzar cambios de contexto
			}
		}(c)
	}

	// Asegurar que todos terminen
	wgProducers.Wait()
	wgConsumers.Wait()

	// 1. Comprobar pérdidas ni duplicados
	if len(consumedItems) != totalItems {
		t.Errorf("Se esperaban %d elementos consumidos, se obtuvieron %d", totalItems, len(consumedItems))
	}
	for i := 1; i <= totalItems; i++ {
		count := consumedItems[i]
		if count == 0 {
			t.Errorf("Pérdida de datos: el elemento %d nunca fue consumido", i)
		} else if count > 1 {
			t.Errorf("Duplicado de datos: el elemento %d fue consumido %d veces", i, count)
		}
	}

	// 2. Verificar que la ocupación permaneció estrictamente entre 0 y 5
	events := bb.GetEvents()
	for _, ev := range events {
		if ev.BufferLen > capacity {
			t.Errorf("Límite excedido: el buffer registró una longitud de %d (máximo %d)", ev.BufferLen, capacity)
		}
		// Descartamos los eventos con BufferLen = -1 que son pre-bloqueos en WaitNotify
		if ev.BufferLen < -1 {
			t.Errorf("Límite inferior excedido: longitud de buffer inválida %d", ev.BufferLen)
		}
	}
}

// TestBoundedBuffer_Bloqueos verifica explícitamente que los hilos se bloqueen
// cuando el buffer está lleno o vacío, y que se emitan los eventos correspondientes.
func TestBoundedBuffer_Bloqueos(t *testing.T) {
	const capacity = 2
	bb := NewBoundedBuffer(capacity)

	var wg sync.WaitGroup

	// --- 1. Probar bloqueo por BUFFER LLENO ---

	// Llenamos el buffer
	bb.Produce(1, 100)
	bb.Produce(1, 101)

	wg.Add(1)
	go func() {
		defer wg.Done()
		bb.Produce(2, 102) // Debería bloquearse aquí
	}()

	// Damos tiempo para asegurar el bloqueo
	time.Sleep(50 * time.Millisecond)

	// Extraemos un elemento para despertar al Productor 2
	bb.Consume(1)
	wg.Wait() // Esperamos a que el Productor 2 termine su inserción

	// Verificamos eventos de bloqueo y reanudación del Productor 2
	verificarEvento(t, bb.GetEvents(), "Productor-2", "BLOCKED")
	verificarEvento(t, bb.GetEvents(), "Productor-2", "RESUMED")

	// --- 2. Probar bloqueo por BUFFER VACÍO ---

	// Vaciamos el buffer completamente (quedaban 101 y 102)
	bb.Consume(1)
	bb.Consume(1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		bb.Consume(2) // Debería bloquearse aquí
	}()

	time.Sleep(50 * time.Millisecond)

	// Producimos para despertar al Consumidor 2
	bb.Produce(1, 103)
	wg.Wait()

	// Verificamos eventos de bloqueo y reanudación del Consumidor 2
	verificarEvento(t, bb.GetEvents(), "Consumidor-2", "BLOCKED")
	verificarEvento(t, bb.GetEvents(), "Consumidor-2", "RESUMED")
}

// verificarEvento es un helper para buscar la existencia de una acción específica de un actor en el log
func verificarEvento(t *testing.T, events []SyncEvent, actor string, action string) {
	t.Helper()
	for _, ev := range events {
		if ev.Actor == actor && ev.Action == action {
			return // Evento encontrado exitosamente
		}
	}
	t.Errorf("No se encontró el evento '%s' para el actor '%s'", action, actor)
}
