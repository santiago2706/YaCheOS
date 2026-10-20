# YaKcheOS — primera mitad funcional

Simulador web de los módulos M1 (procesos, hilos y CPU), M2 (sincronización) y M3 (memoria). El alcance y los casos numéricos se definen en [el plan de sprints](plan_de_sprints_yakcheos_6_modulos.md). El servidor Go conserva el estado de la simulación; C calcula planificación, asignación y traducción de memoria; Python calcula y audita el reemplazo de páginas.

## Requisitos

Go, Python 3 con `pytest` y GCC. Para la demostración real de `fork/exec/waitpid` se necesita Linux o Ubuntu en WSL. Los procesos y PID de C son reales únicamente en ese demostrador; los del Task Manager son simulados.

## Compilar los motores C

En Linux o WSL, desde la raíz:

```bash
make -C src/core_c
```

En Windows con GCC, desde PowerShell en la raíz:

```powershell
New-Item -ItemType Directory -Path bin -Force | Out-Null
gcc -std=c11 -Wall -Wextra -Werror -O2 -o bin/scheduler.exe src/core_c/queue.c src/core_c/scheduler_nonpreemptive.c src/core_c/scheduler_preemptive.c src/core_c/scheduler.c
gcc -std=c11 -Wall -Wextra -Werror -O2 -o bin/memory_translate.exe src/core_c/memory_translate.c
gcc -std=c11 -Wall -Wextra -Werror -O2 -DCORE_ENGINE_UNIFIED -o bin/memory_cli.exe src/core_c/queue.c src/core_c/scheduler_nonpreemptive.c src/core_c/scheduler_preemptive.c src/core_c/scheduler.c src/core_c/memory_contigous.c src/core_c/memory_paging.c src/core_c/memory_segmentation.c src/core_c/memory.c
```

Los binarios de `bin/` son locales y están ignorados por Git. Recompílalos después de cambiar las fuentes C. En Linux, `make` también genera `process_demo`.

## Ejecutar

```bash
go run ./cmd/server
```

Abrir [http://localhost:8080](http://localhost:8080). No basta con `go run` en una copia recién clonada: primero se compilan los motores C anteriores.

## Pruebas

```bash
python -m pytest -q tests_python
go test -count=1 ./...
make -C src/core_c test
```

En Windows sin `make`, el test C se puede compilar así:

```powershell
gcc -std=c11 -Wall -Wextra -Werror -O2 -o bin/core_test.exe src/core_c/main_test.c src/core_c/queue.c src/core_c/memory_contigous.c
bin/core_test.exe
```

Para comprobar carreras de Go: `go test -race ./src/sync_go ./cmd/server`. En Windows se necesita un GCC de 64 bits compatible con Go `amd64`; un GCC `mingw32` de 32 bits no sirve para esa prueba.

## Recorrido de demostración

1. En Task Manager, cambiar FCFS/SJF/RR/Prioridad y observar Gantt, bloques de espera, tabla y promedios. Avanzar ticks, terminar P2 y comprobar que el planificador C recalcula la línea de tiempo y que el asignador C libera 64 KB.
2. Ejecutar la demostración de dos hilos. En Linux/WSL, ejecutar también `fork/exec` y observar PID/PPID, recolección con `waitpid` y la diferencia frente al PCB simulado.
3. En Sincronización, ejecutar productor-consumidor y los dos modos de filósofos; comprobar la alerta RAG y el reinicio.
4. En MMU, traducir 10000 a 30480 y verificar segmentación válida y trap. En Reemplazo, comparar FIFO/LRU/Óptimo con el mismo patrón y marcos, abrir la traza por referencia y observar la anomalía de Belady (FIFO: 9 fallos con 3 marcos, 10 con 4).

Las entradas de referencia están en `data/` y `tests_python/fixtures/`. Las trazas C, SVG e informe comparativo están en [docs/informe_tecnico_beta.md](docs/informe_tecnico_beta.md).

## API principal

- `GET /api/v1/simulation`, `POST /api/v1/simulation`: estado, paso, reinicio y terminación de PID.
- `GET/POST /api/v1/cpu/schedule`: ejecución del planificador C.
- `GET /api/v1/memory/map` y `GET/POST /api/v1/memory/translate`: asignador y traductor C.
- `GET/POST /api/v1/memory/replace`: Python con entrada de referencias y marcos, tabla y trazas.
- `GET /api/v1/sync/status`, `GET/POST /api/v1/sync/step`: sincronización Go.
- `GET /api/v1/thread/demo`, `GET /api/v1/process/demo`: demostradores controlados (el segundo requiere Linux/WSL).
