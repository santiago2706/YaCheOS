package syncgo

import (
	"sync"
	"testing"
)

func TestCountingSemaphoreRace(t *testing.T) {
	sem := NewCountingSemaphore(3)
	var active int
	var maxActive int
	var mu sync.Mutex
	var wg sync.WaitGroup

	numGoroutines := 15
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			sem.Wait()

			mu.Lock()
			active++
			if active > maxActive {
				maxActive = active
			}
			mu.Unlock()

			// Sección crítica simulada
			mu.Lock()
			active--
			mu.Unlock()

			sem.Signal()
		}()
	}

	wg.Wait()

	if maxActive > 3 {
		t.Fatalf("El número máximo de goroutines activas (%d) superó el límite permitido (3)", maxActive)
	}
}

func TestProdConsSimulation(t *testing.T) {
	bb := NewBoundedBuffer(5)
	var wg sync.WaitGroup

	// 2 productores producen 10 elementos cada uno = 20
	wg.Add(2)
	for p := 1; p <= 2; p++ {
		pID := p
		go func() {
			defer wg.Done()
			for i := 1; i <= 10; i++ {
				bb.Produce(pID, pID*100+i)
			}
		}()
	}

	// 2 consumidores consumen 10 elementos cada uno = 20
	var consumedItems []int
	var consMu sync.Mutex
	wg.Add(2)
	for c := 1; c <= 2; c++ {
		cID := c
		go func() {
			defer wg.Done()
			for i := 1; i <= 10; i++ {
				item := bb.Consume(cID)
				consMu.Lock()
				consumedItems = append(consumedItems, item)
				consMu.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(consumedItems) != 20 {
		t.Fatalf("Se esperaban 20 elementos consumidos, pero se obtuvieron %d", len(consumedItems))
	}
}
