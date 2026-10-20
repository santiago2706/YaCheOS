package syncgo

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestSemaphore_ConcurrenciaMasiva verifica que de 15 goroutines,
// máximo 3 puedan entrar simultáneamente en la sección crítica.
func TestSemaphore_ConcurrenciaMasiva(t *testing.T) {
	const maxPermisos = 3
	const totalGoroutines = 15

	sem := NewCountingSemaphore(maxPermisos)

	var enSeccionCritica atomic.Int32
	var wg sync.WaitGroup

	// Canal para controlar el timeout total de la prueba
	completado := make(chan struct{})

	for i := 0; i < totalGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			sem.Wait() // Adquiere permiso

			// --- Inicio Sección Crítica ---
			actual := enSeccionCritica.Add(1)

			// Verificamos que no se exceda el límite de permisos concurrentes
			if actual > maxPermisos {
				t.Errorf("Fallo de exclusión: se detectaron %d hilos en la sección crítica, máximo permitido era %d", actual, maxPermisos)
			}

			// Simulamos trabajo para forzar superposición de hilos
			time.Sleep(20 * time.Millisecond)

			enSeccionCritica.Add(-1)
			// --- Fin Sección Crítica ---

			sem.Signal() // Libera permiso
		}()
	}

	// Goroutine auxiliar para esperar la finalización y cerrar el canal
	go func() {
		wg.Wait()
		close(completado)
	}()

	// Incorporamos límite de tiempo para detectar bloqueos indefinidos (deadlocks)
	select {
	case <-completado:
		// La prueba finalizó exitosamente antes del timeout
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout: Las goroutines sufrieron un bloqueo permanente (deadlock) y la prueba no terminó")
	}
}

// TestSemaphore_BloqueoYDesbloqueo verifica que una adquisición adicional se bloquee
// y que un Signal() despierte a una goroutine en espera.
func TestSemaphore_BloqueoYDesbloqueo(t *testing.T) {
	sem := NewCountingSemaphore(1)

	// Consumimos el único permiso disponible
	sem.Wait()

	adquirido := make(chan struct{})

	go func() {
		sem.Wait() // Debería bloquearse aquí porque ya no hay permisos
		close(adquirido)
		sem.Signal()
	}()

	// 1. Comprobar que la adquisición extra se bloquee hasta liberar un permiso
	select {
	case <-adquirido:
		t.Fatal("Fallo: La goroutine adquirió un permiso inexistente sin bloquearse")
	case <-time.After(50 * time.Millisecond):
		// Éxito: el canal no se cerró, lo que significa que la goroutine está bloqueada en Wait()
	}

	// 2. Comprobar que Signal() desbloquea a la goroutine
	sem.Signal()

	select {
	case <-adquirido:
		// Éxito: al hacer Signal(), la goroutine pudo avanzar y cerró el canal
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout: Signal() no logró despertar a la goroutine en espera")
	}
}

// TestSemaphore_AliasPV prueba específicamente los alias P() y V() para asegurar
// que se comporten exactamente igual que Wait() y Signal()
func TestSemaphore_AliasPV(t *testing.T) {
	sem := NewCountingSemaphore(1)

	// Consumimos con P()
	sem.P()

	adquirido := make(chan struct{})
	go func() {
		sem.P() // Espera por permiso
		close(adquirido)
		sem.V() // Libera
	}()

	// Verificar bloqueo con P()
	select {
	case <-adquirido:
		t.Fatal("Fallo: P() no bloqueó la ejecución cuando no había permisos")
	case <-time.After(50 * time.Millisecond):
	}

	// Liberar con V()
	sem.V()

	// Verificar desbloqueo con V()
	select {
	case <-adquirido:
		// Éxito
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout: V() no desbloqueó a la goroutine en espera")
	}
}
