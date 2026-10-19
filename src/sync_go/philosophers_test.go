package syncgo

import (
	"testing"
	"time"
)

// TestPhilosophers_Deadlock verifica que la modalidad "deadlock" genera
// un bloqueo circular natural en pocos pasos sin forzar manualmente los estados.
func TestPhilosophers_Deadlock(t *testing.T) {
	dp := NewDiningPhilosophers("deadlock")

	deadlockReached := make(chan bool)

	go func() {
		// Simulamos ciclos donde todos los filósofos actúan secuencialmente.
		// En modo "deadlock", todos pasarán a Hungry, luego todos tomarán
		// su tenedor izquierdo, y al intentar el derecho, se interbloquearán.
		for cycle := 0; cycle < 100; cycle++ {
			for i := 0; i < dp.Count; i++ {
				dp.StepPhilosopher(i)
			}

			isDeadlock, _ := dp.CheckDeadlock()
			if isDeadlock {
				deadlockReached <- true
				return
			}
			time.Sleep(5 * time.Millisecond) // Simular un pequeño tiempo de procesamiento
		}
		deadlockReached <- false
	}()

	// Límite estricto de 5 segundos para demostrar el interbloqueo
	select {
	case reached := <-deadlockReached:
		if !reached {
			t.Fatal("Fallo: No se alcanzó el deadlock de forma natural tras los ciclos simulados.")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout: El deadlock no se formó en menos de 5 segundos.")
	}
}

// TestPhilosophers_PreventivaProgreso ejecuta 10 000 ciclos en modo "preventiva".
//
// Aclaración importante:
// Un ciclo de simulación se define como dar la oportunidad a cada filósofo (del 0 al 4)
// de ejecutar un StepPhilosopher(). Esta prueba aporta evidencia empírica clara
// de que los filósofos progresan y comen en esta implementación determinista.
// Sin embargo, NO constituye una demostración matemática formal de ausencia
// general de inanición (starvation-freedom) bajo cualquier entrelazado asíncrono posible.
func TestPhilosophers_PreventivaProgreso(t *testing.T) {
	dp := NewDiningPhilosophers("preventiva")
	const ciclosSimulacion = 10000

	for cycle := 0; cycle < ciclosSimulacion; cycle++ {
		for i := 0; i < dp.Count; i++ {
			dp.StepPhilosopher(i)
		}
	}

	// 1. Verificar que no se produjo un interbloqueo
	isDeadlock, blocked := dp.CheckDeadlock()
	if isDeadlock {
		t.Fatalf("Fallo crítico: Se produjo un deadlock inesperado en modo preventiva. Bloqueados: %v", blocked)
	}

	// 2. Registrar y verificar el progreso de cada filósofo
	for i := 0; i < dp.Count; i++ {
		if dp.EatCount[i] == 0 {
			t.Errorf("Inanición detectada: El filósofo %d nunca logró comer tras %d ciclos.", i, ciclosSimulacion)
		}
		// Opcional: Imprimir métricas de progreso si se ejecuta con verbosidad (go test -v)
		t.Logf("Filósofo %d comió %d veces", i, dp.EatCount[i])
	}
}

// TestPhilosophers_SetModeAndReset comprueba el cambio de modalidades
// y la limpieza de variables para garantizar simulaciones consistentes.
func TestPhilosophers_SetModeAndReset(t *testing.T) {
	dp := NewDiningPhilosophers("deadlock")

	// Ensuciamos el estado: hacemos que todos tomen su primer tenedor
	for i := 0; i < dp.Count; i++ {
		dp.StepPhilosopher(i) // Pasan a StateHungry
		dp.StepPhilosopher(i) // Toman el tenedor izquierdo
	}

	// Cambiar a "preventiva" debería limpiar todo y liberar recursos retenidos
	err := dp.SetMode("preventiva")
	if err != nil {
		t.Fatalf("Error inesperado al cambiar de modo: %v", err)
	}

	// Validar que el reinicio surtió efecto
	for i := 0; i < dp.Count; i++ {
		if dp.States[i] != StateThinking {
			t.Errorf("Tras SetMode(), el filósofo %d debería estar PENSANDO, pero está en estado: %s", i, dp.States[i])
		}
		if dp.ForksOwner[i] != -1 {
			t.Errorf("Tras SetMode(), el tenedor %d debería estar libre (-1), pero lo retiene el filósofo %d", i, dp.ForksOwner[i])
		}
	}

	// Ensuciar de nuevo para probar la función explícita Reset()
	dp.StepPhilosopher(0) // 0 pasa a Hungry
	dp.StepPhilosopher(0) // 0 toma un tenedor

	dp.Reset()

	if dp.States[0] != StateThinking || dp.ForksOwner[0] != -1 || dp.EatCount[0] != 0 {
		t.Fatal("Reset() no restauró correctamente el estado inicial de la simulación.")
	}
}
