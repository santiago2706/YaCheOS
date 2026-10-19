# YaKcheOS: Plan Maestro de Sprints y Guía de Tareas por Integrante
**Universidad Nacional de Ingeniería (UNI)**  
**Facultad de Ingeniería Industrial y de Sistemas (FIIS) · Escuela Profesional de Ingeniería de Software**  
**Curso:** SW407 - Sistemas Operativos (2026-II)  
**Docente:** Mg. Ing. José Carlos García La Riva  
**Proyecto:** YaKcheOS (Simulador Web de Sistema Operativo)  
**Equipo:** 4 Integrantes  
**Stack de Desarrollo:** C (Core y Memoria), Go (Orquestación Web y Sincronización), Python (Reemplazo de Páginas, QA, Oráculo Teórico y Analítica)  

**Descripción del proyecto:** YaKcheOS es un simulador educativo de sistema operativo que se utiliza desde el navegador. Permite cargar procesos, ejecutar algoritmos de planificación, observar la memoria, analizar la sincronización y, en los siguientes sprints, trabajar con archivos virtuales, dispositivos de E/S, seguridad y comunicación entre nodos. El escritorio y la terminal son interfaces para explorar estos módulos; los resultados se representan mediante estados, trazas, mapas de memoria y métricas verificables.

**Documento de referencia:** `Propuesta_Proyecto_SW407 (2).pdf`, secciones IV (seis módulos), V (entregas) y VI (rúbrica).

**Base de fechas:** Inicio operativo propuesto: **12/10/2026**. Los 10 días son **días calendario consecutivos**, del 12 al 21 de octubre, ambos incluidos. Si el inicio real cambia, desplazar todas las fechas manteniendo las duraciones. Estas fechas organizan el desarrollo interno; no sustituyen las semanas de evaluación del curso.

**Distribución Temporal del Ciclo:**

* **Fase 1 — Primera mitad funcional (12–21 de octubre; días 1–10):** Módulos 1, 2 y 3 completos e integrados: procesos/hilos/CPU, sincronización y memoria **incluido reemplazo FIFO, LRU y Óptimo**. Interfaz web, pruebas e informe del hito acompañan la entrega.

* **Fase 2 — Segunda mitad y cierre (22 de octubre–9 de diciembre):** Módulos 4, 5 y 6: E/S y archivos, seguridad y sistemas distribuidos; después, estrés, documentación y sustentación.

* **Duración interna propuesta:** 59 días calendario: 10 días para la primera mitad y las 7 semanas restantes del plan anterior para completar e integrar la segunda mitad. Se conserva el equipo de 4 integrantes y la organización en 5 sprints; se comprimen los dos primeros a 5 días cada uno.

**Qué significa “mitad terminada”:** 3 de los 6 módulos de la propuesta deben poder ejecutarse, verificarse y demostrarse al día 10. Es el 50% de los módulos, no una afirmación de que representen el 50% del esfuerzo. Un módulo con algoritmos pendientes, datos ficticios en la interfaz o sin integración no se considera terminado.

---

## 📌 Guía Rápida para el Integrante
> **Si eres miembro del equipo:** Busca tu sección (Integrante 1, 2, 3 o 4). Allí tienes tus herramientas, los archivos exactos que te toca programar en el repositorio, tus comandos de prueba local y tu **Definition of Done (DoD)** antes de abrir un Pull Request en GitHub.
>
> **Alineación académica:** La propuesta establece seis módulos y una rúbrica con 30% de correctitud y 15% de análisis comparativo. Se mantienen esos criterios. Los ejemplos numéricos del plan sirven como casos de prueba; las referencias a láminas específicas deben comprobarse con el material de clase.
>
> **Cómo leer los tiempos:** Sprint 1: días 1–5; Sprint 2: días 6–10; Sprint 3: 22/10–04/11; Sprint 4: 05/11–18/11; Sprint 5: 19/11–09/12. Las semanas del curso son referencias académicas independientes de estas fechas.
>
> **Cómo se completa una tarea:** Implementar la funcionalidad, comprobar el caso indicado, integrar su salida con el módulo que la utiliza y solicitar la revisión cruzada. El DoD es el criterio verificable para darla por terminada; no basta con que la interfaz aparezca en pantalla.

---

## 1. Arquitectura de Integración y Reglas de Coordinación

### 1.1 Principios del Simulador
1. **Sin manipulación del kernel anfitrión:** Es un simulador que corre como una máquina de estados determinista en memoria; no modifica el kernel, la memoria física ni los dispositivos reales del equipo.
2. **Interfaz ligera con Go, HTML5/CSS y JavaScript mínimo:** Go levanta un servidor HTTP local (`net/http`) que sirve la interfaz y expone endpoints JSON (`/api/...`). Los scripts de `web/js/` gestionan peticiones, ventanas, terminal y actualización visual; no se introduce un framework de JavaScript ni TypeScript.
3. **Contratos de datos desacoplados:** Los módulos en C, Go y Python leen parámetros mediante entradas estándar o archivos JSON y producen salidas estructuradas (`cpu_timeline.json`, `memory_map.json`, `events.json`). Nadie se bloquea esperando código ajeno.
4. **Ejecución y estado de sesión:** Go conserva el estado de la simulación y del VFS durante la sesión activa. El motor en C recibe entradas y devuelve resultados; al ejecutarlo como subproceso, su memoria no persiste entre llamadas. Los cambios del VFS deben devolver el estado actualizado para que Go lo conserve. La terminal interpreta únicamente comandos del simulador y nunca ejecuta una shell del equipo anfitrión.
5. **Trazas reproducibles:** CPU y memoria se ejecutan con entradas y reglas de desempate definidas. En Go, el orden real de las goroutines puede variar: las pruebas verifican invariantes y usan coordinación explícita cuando se necesita reproducir un interbloqueo.
6. **Regla de revisión cruzada obligatoria:** Ningún cambio se fusiona en la rama principal (`main`) sin la revisión y aprobación de otro integrante del equipo.

### 1.2 Correspondencia con los Seis Módulos de la Propuesta

| Módulo de la propuesta | Cobertura encontrada en el plan anterior | Ajuste incorporado | Cierre previsto |
|---|---|---|---|
| **M1. Procesos, hilos y planificación** | PCB, cola, FCFS, SJF, RR y prioridad; faltaba una actividad explícita de `fork/exec` y hilos. | T2.0: creación/terminación con `fork/exec` en C y demostración de hilos con `threading` en Python. | **21/10**, Sprint 2. |
| **M2. Sincronización** | Semáforos, mutex, productor-consumidor, filósofos y detector de interbloqueos. | Mantener el contenido y comprobar invariantes, terminación y progreso; distinguir la demostración intencional de bloqueo del modo correcto. | **21/10**, Sprint 2. |
| **M3. Gestión de memoria** | Asignación contigua, fragmentación, paginación y segmentación; reemplazo estaba después del primer hito. | Adelantar S3.1 (FIFO/LRU/Óptimo) al Sprint 2, con implementación paralela en el Sprint 1 y comparación en el informe Beta. | **21/10**, Sprint 2. |
| **M4. E/S y sistemas de archivos** | VFS, interrupciones, buffering implícito, spooler y permisos. | Añadir sondeo, DMA y buffering explícito; comparar asignación contigua, enlazada e indexada. | **18/11**, Sprint 4; VFS base el 04/11. |
| **M5. Seguridad** | Usuarios y permisos `rwx`; sin autenticación ni cifrado explícitos. | Añadir autenticación, control de acceso, demostraciones simétrica/asimétrica y comparación conceptual VM/contenedor. | **18/11**, Sprint 4. |
| **M6. Sistemas distribuidos** | Ausente. | Incorporar paso de mensajes entre dos procesos que representan nodos, con trazas verificables. | **25/11**, primera semana del Sprint 5. |

**Se conservan los módulos y funcionalidades previos.** Segmentación, filósofos, detección de interbloqueos, escritorio, terminal, editor y spooler siguen en el plan. Los agregados completan la propuesta; no introducen un séptimo módulo.

**Ficha técnica:** Mantener C para el motor, Go para orquestación/sincronización/comunicación y Python para validación y análisis. La propuesta sugiere C/Python; justificar el uso adicional de Go en la ficha técnica y contrastar su aceptación con el docente. `fork/exec` se demuestra en un programa didáctico controlado de C en Linux/WSL, separado de los PIDs simulados. Esta actividad usa llamadas reales para cumplir M1, sin modificar el kernel ni conceder ejecución libre desde la terminal web.

### 1.3 Estructura del Repositorio Compartido
```text
yakcheos/
├── cmd/
│   └── server/             <- [Rol 1: Go] main.go del servidor web y handlers API
├── src/
│   ├── core_c/             <- [Rol 2: C] Algoritmos de CPU, Memoria y VFS
│   │   ├── pcb.h
│   │   ├── scheduler.c
│   │   ├── memory.c
│   │   └── Makefile
│   └── sync_go/            <- [Rol 3: Go] Concurrencia y Sincronización
│       ├── semaphore.go
│       ├── prod_cons.go
│       └── philosophers.go
├── web/                    <- [Rol 1: Go] Archivos estáticos de interfaz
│   ├── index.html
│   ├── css/
│   │   ├── desktop.css
│   │   └── terminal.css
│   └── js/app.js           <- Script mínimo para llamadas AJAX/Fetch
├── tests_python/           <- [Rol 4: Python] Suite de validación y métricas
│   ├── ground_truth_semana4.py
│   ├── test_scheduler.py
│   └── generate_gantt.py
├── docs/                   <- Informes Beta/final y comparación de virtualización
└── data/                   <- Contratos JSON y casos base de prueba
    ├── procesos_semana4.json
    ├── sincronizacion_semana5.json
    ├── memoria_semana6.json
    ├── page_references.json
    └── distributed_case.json
```

Extensiones de la misma arquitectura: `cmd/node/` y `src/distributed_go/` para M6; `src/core_c/process_demo.c` para M1; scripts `thread_demo.py`, `page_replacement.py` y pruebas de seguridad/distribuidos en `tests_python/`. Los handlers de autenticación y cifrado se agregan al servidor Go.

---

## 2. Contratos de Interfaz (Casos Base de Validación)

### 2.1 Caso Base de CPU - Planificación de Procesos (Semana 4: `data/procesos_semana4.json`)
Se conservan las siguientes entradas numéricas del plan original para validar CPU. Las referencias académicas a láminas deben verificarse con el material del curso:

* **P1:** Llegada $t = 0\text{ min}$, Duración = $5\text{ min}$, Prioridad = $3$

* **P2:** Llegada $t = 1\text{ min}$, Duración = $3\text{ min}$, Prioridad = $4$ (más baja)

* **P3:** Llegada $t = 2\text{ min}$, Duración = $8\text{ min}$, Prioridad = $1$ (más alta)

* **P4:** Llegada $t = 3\text{ min}$, Duración = $6\text{ min}$, Prioridad = $2$

```json
[
  {"pid": 1, "name": "P1", "arrival_time": 0, "burst_time": 5, "priority": 3, "memory_kb": 128},
  {"pid": 2, "name": "P2", "arrival_time": 1, "burst_time": 3, "priority": 4, "memory_kb": 64},
  {"pid": 3, "name": "P3", "arrival_time": 2, "burst_time": 8, "priority": 1, "memory_kb": 256},
  {"pid": 4, "name": "P4", "arrival_time": 3, "burst_time": 6, "priority": 2, "memory_kb": 128}
]
```

#### Resultados Teóricos que el Simulador DEBE Reproducir:
Las duraciones se mantienen en minutos de tiempo **simulado**; no representan cuánto tarda el programa en ejecutarse. Una unidad equivale a un tick en estos casos. Se usa `espera = finalización − llegada − ráfaga`, `retorno = finalización − llegada` y `respuesta = primer inicio − llegada` para procesos con una única ráfaga de CPU. Ante empates, se compara llegada y luego PID. En RR, las llegadas hasta el final del quantum se encolan antes de reinsertar el proceso desalojado.

* **FCFS:** Gantt: $P_1[0-5] \to P_2[5-8] \to P_3[8-16] \to P_4[16-22]$. Esperas: $P_1: 0, P_2: 4, P_3: 6, P_4: 13$. **Espera Promedio: 5.75 min**.

* **SJF (No apropiativo):** Gantt: $P_1[0-5] \to P_2[5-8] \to P_4[8-14] \to P_3[14-22]$. Esperas: $P_1: 0, P_2: 4, P_4: 5, P_3: 12$. **Espera Promedio: 5.25 min** *(menor espera promedio entre los algoritmos comparados para este caso)*.

* **Round Robin (Quantum = 4 min):** Gantt: $P_1[0-4] \to P_2[4-7] \to P_3[7-11] \to P_4[11-15] \to P_1[15-16] \to P_3[16-20] \to P_4[20-22]$. Esperas: $P_1: 11, P_2: 3, P_3: 10, P_4: 13$. **Espera Promedio: 9.25 min**, **Respuesta Promedio: 4.0 min**.

* **Prioridad (No apropiativo):** Gantt: $P_1[0-5] \to P_3[5-13] \to P_4[13-19] \to P_2[19-22]$. Esperas: $P_1: 0, P_3: 3, P_4: 10, P_2: 18$. **Espera Promedio: 7.75 min**.

### 2.2 Caso Base de Sincronización (Semana 5: `data/sincronizacion_semana5.json`)

* **Productor-Consumidor (buffer acotado):** Capacidad = 5 elementos, 2 productores (10 elementos cada uno), 2 consumidores (10 elementos cada uno), total = 20 elementos producidos y 20 consumidos. Semáforos: `mutex=1`, `vacios=5`, `llenos=0`.

* **Filósofos comensales:** 5 procesos y 5 recursos de exclusión mutua en disposición circular. Cada proceso necesita sus dos recursos adyacentes. Modos conmutables: interbloqueo por adquisición simultánea del recurso izquierdo y prevención mediante orden asimétrico.

### 2.3 Caso Base de Memoria (Semana 6: `data/memoria_semana6.json`)

* **Paginación con traducción de direcciones:** Tamaño página/marco = $4096\text{ Bytes}$. Tabla del proceso #100: P0 $\to$ M5, P1 $\to$ M2, P2 $\to$ M7, P3 $\to$ M0.
  * **Cálculo obligatorio:** Dirección virtual $10\,000 \implies \text{Página } 2 \text{ (Marco } 7\text{), Desplazamiento } 1808 \implies \text{Dirección Física } = 7 \times 4096 + 1808 = \mathbf{30\,480}$.

* **Segmentación:** Segmento 0 (Código): Base 1000, Límite 400. Segmento 1 (Datos): Base 5200, Límite 800. Segmento 2 (Pila): Base 9000, Límite 300.
  * Comprobación válida: $(1, 300) \to 5200 + 300 = 5500$.
  * Comprobación de fallo (*Trap*): $(1, 900) \to 900 > 800 \to \text{Violación de segmento (Excepción)}$.

---

## 3. Fase 1: Primera Mitad Terminada en 10 Días (Módulos 1, 2 y 3)

**Sprint 1 — Bases y motores (12–16 de octubre; días 1–5):** El Integrante 1 prepara servidor, escritorio, terminal y contratos. El Integrante 2 desarrolla procesos y CPU, y establece el modelo de memoria. El Integrante 3 construye sincronización. El Integrante 4 prepara el oráculo, la demostración de hilos y el reemplazo de páginas en Python, para reducir la dependencia del motor C. El cierre exige motores ejecutables por separado y primeras pruebas; todavía no es la entrega de la mitad.

**Sprint 2 — Completar e integrar M1–M3 (17–21 de octubre; días 6–10):** Completar memoria, conectar reemplazo de páginas, finalizar sincronización, integrar endpoints y mostrar resultados en web. Los días 9 y 10 se reservan para validación, correcciones e informe, utilizando entradas reales del simulador. La versión `v0.5.0-beta` identifica los tres módulos terminados.

**Orden de integración:** Contratos → motores independientes → endpoints → vistas → pruebas conjuntas → informe y demo. Para el reemplazo, Go invoca un script Python con entrada/salida JSON; CPU y traducción/asignación de memoria mantienen el motor C. Esta distribución conserva C/Go/Python y deja cada algoritmo dentro de su módulo académico.

### Plan Diario del Hito de 10 Días

| Día / fecha | Integrante 1: web e integración | Integrante 2: C, procesos y memoria | Integrante 3: sincronización Go | Integrante 4: Python, QA y analítica |
|---|---|---|---|---|
| 1 · 12/10 | Repositorio, servidor, contratos JSON. | PCB, cola y demostrador `fork/exec`. | Semáforo y mutex. | Oráculo y demostrador `threading`. |
| 2 · 13/10 | Escritorio y ventanas básicas. | FCFS y SJF. | Buffer productor-consumidor. | Validar CPU y entradas de reemplazo. |
| 3 · 14/10 | Terminal y comandos CPU con datos de prueba de contrato. | RR y prioridad no apropiativa. | Pruebas del buffer y trazas. | Implementar FIFO y LRU. |
| 4 · 15/10 | Endpoint CPU y consumo de trazas. | Prioridad apropiativa/aging; bloques contiguos. | Filósofos: dos modalidades. | Implementar Óptimo y contar fallos. |
| 5 · 16/10 | CPU visible; primera integración. | First/Best Fit, liberación y fragmentación. | Detector de interbloqueos y eventos. | Comparar CPU y reemplazo; revisar M1. |
| 6 · 17/10 | Endpoints de memoria y reemplazo. | Paginación y traducción. | Integrar endpoints de sincronización. | Gantt, Belady y tablas comparativas. |
| 7 · 18/10 | Task Manager y mapa de memoria. | Segmentación y casos inválidos. | Vista de buffer y recursos con apoyo del Rol 1. | Pruebas de memoria y regresión CPU. |
| 8 · 19/10 | Flujo conjunto y control de procesos simulados. | Correcciones de CPU/memoria. | Cancelación/reinicio y liberación de recursos. | Informe Beta y pruebas conjuntas. |
| 9 · 20/10 | Demo candidata y errores de interfaz/API. | Corregir hallazgos; revisar evidencias. | Corregir carreras/bloqueos involuntarios. | Ejecutar checklist M1–M3 y revisión cruzada. |
| 10 · 21/10 | Entrega ejecutable y etiqueta Beta. | Demo CPU y memoria. | Demo concurrente correcta. | Informe, resultados y demo de reemplazo. |

**Condición de viabilidad:** Es un plan acelerado, no una estimación de horas garantizada. Se requiere trabajo paralelo de los cuatro integrantes y revisión diaria. Se mantiene la interfaz ligera; refinamientos de animación, estética y ordenación avanzada pueden terminarse en el Sprint 5. Los algoritmos y pruebas de M1–M3 no se posponen. El plan no presupone una jornada de 2 horas por persona, porque no se ha fijado esa disponibilidad para este proyecto.

### Criterios de Aceptación de la Primera Mitad — 21/10

- **M1:** demostrar creación y terminación con `fork/exec`, ejecución de hilos y FCFS/SJF/RR/Prioridad; calcular espera, retorno y respuesta y comparar con el oráculo. Distinguir procesos del sistema anfitrión de los procesos simulados.

- **M2:** productor-consumidor sin pérdidas, duplicados, overflow ni underflow; semáforos/mutex sin carreras. El modo preventivo de filósofos termina con progreso de todos; el modo de bloqueo intencional se muestra identificado y puede cancelarse/reiniciarse.

- **M3:** First Fit/Best Fit, liberación, fragmentación, paginación y traducción; FIFO/LRU/Óptimo sobre el mismo patrón y número de marcos, con fallos y tabla comparativa. Conservar también segmentación y validación de límites.

- **Integración:** ejecutar los tres módulos desde la web; ver trazas, Gantt, métricas y mapa de memoria. No cuentan como terminados resultados escritos directamente en la interfaz.

- **Evidencias:** repositorio con participación de los cuatro, instrucciones de arranque, casos reproducibles, informe Beta y demo de aproximadamente 5 minutos.

---

### 👤 INTEGRANTE 1: Líder Técnico & Orquestador Web

* **Tecnología:** Go (`net/http`, HTML5/CSS, `os/exec`)

* **Objetivo:** Orquestar el simulador para que se abra en el navegador con un solo comando (`go run ./cmd/server`), ofreciendo escritorio, terminal Linux y monitor de recursos.

#### Sprint 1 (12–16/10 · Días 1–5): Servidor, Ventanas y Terminal tipo Linux (Semana 1 y 2 del Curso)

* **Tarea T1.1: Inicialización de Repositorio y Servidor Base en Go**
  * *Qué harás:* Configurar el repositorio Git, archivo `.gitignore`, ramas protegidas y un servidor HTTP nativo en Go que escuche en el puerto `8080` y sirva archivos estáticos desde `/web`.
  * *Archivos a crear:* `cmd/server/main.go`, `go.mod`.
  * *Cómo probarlo:* Ejecutar `go run ./cmd/server` y abrir en el navegador `http://localhost:8080`.
  * *Definition of Done (DoD):* 
    * El servidor compilado inicia en menos de 1 segundo en el entorno de prueba documentado; el tiempo de compilación de `go run` se mide por separado.
    * Devuelve HTTP 200 con la página de bienvenida.
    * Código formateado con `gofmt` y sin dependencias externas pesadas.

* **Tarea T1.2: Sistema de Ventanas Flotantes de Escritorio**
  * *Qué harás:* Maquetar la interfaz visual: fondo de escritorio, barra de tareas inferior con menú de inicio, reloj y un gestor de ventanas ligero que permita moverlas arrastrando la barra superior, minimizarlas, maximizarlas y cerrarlas.
  * *Archivos a crear:* `web/index.html`, `web/css/desktop.css`.
  * *Cómo probarlo:* Abrir la web y abrir 3 ventanas simultáneas (Terminal, Task Manager, Sincronización).
  * *Definition of Done (DoD):*
    * Se pueden abrir, arrastrar, minimizar y cerrar 3 ventanas sin que se bloquee el navegador.
    * Comportamiento responsivo y visualmente limpio; el objetivo de fluidez de 60 FPS se revisa en el entorno de prueba. Su refinamiento puede cerrarse en Sprint 5 sin retrasar los módulos M1–M3.

* **Tarea T1.3: Emulador de Terminal CLI en Web (Taller Semana 2)**
  * *Qué harás:* Crear la ventana de Terminal con pantalla negra y prompt `yakcheos:~$`. Debe recibir comandos por teclado y despachar acciones del laboratorio de la Semana 2:
    * `help`: Lista los comandos disponibles.
    * `clear`: Limpia la pantalla.
    * `ps`: Muestra la lista de procesos activos.
    * `run <algo> <archivo.json>`: Solicita simulación de CPU.
    * `ls`, `mkdir`, `cat`: Comandos del sistema de archivos simulado. En la Beta se reconoce su sintaxis y se informa que su ejecución se habilitará con el VFS del Sprint 3.
  * *Archivos a crear:* `web/css/terminal.css`, `web/js/terminal.js`.
  * *Definition of Done (DoD):*
    * Soporta historial de comandos con flechas de arriba/abajo.
    * Muestra mensajes de error claros ante comandos inexistentes.

#### Sprint 2 (17–21/10 · Días 6–10): Endpoints de Integración y Task Manager

* **Tarea T1.4: Endpoints API en Go e Invocación del Motor en C (Semanas 4 y 6)**
  * *Qué harás:* Crear los handlers HTTP en Go (`/api/v1/cpu/schedule`, `/api/v1/memory/translate`) que reciban la lista de procesos desde la web, ejecuten el binario de C generado por el Integrante 2 vía `os/exec.Command`, capturen la salida JSON y la devuelvan al navegador.
  * *Archivos a modificar:* `cmd/server/handlers_core.go`.
  * *Cómo probarlo:* Enviar una petición POST vía `curl` o desde la terminal web y comprobar que retorne el JSON de C.
  * *Definition of Done (DoD):*
    * Latencia de respuesta < 200 ms para lotes de 15 procesos.
    * Manejo de código de salida de C: si C falla, Go devuelve error HTTP 400 detallado sin caerse el servidor (*no panic*).

* **Tarea T1.5: Ventana del Administrador de Tareas (Task Manager)**
  * *Qué harás:* Construir la interfaz de la ventana "Task Manager" que muestra:
    1. Tabla en vivo de procesos (PID, Nombre, Estado, CPU restante, RAM).
    2. Mapa gráfico de bloques de memoria (bloques ocupados y libres).
    3. Botón para terminar procesos (`kill <PID>`) de la simulación y actualizar su estado y memoria asociada.
  * *Definition of Done (DoD):*
    * La tabla se actualiza dinámicamente con cada tick de simulación.
    * Permite ordenar columnas al hacer clic en el encabezado; esta mejora puede cerrarse en Sprint 5 si las métricas y estados ya se visualizan correctamente en el día 10.

* **Tarea T1.6: Integración Final de la Versión Beta**
  * *Qué harás:* Integrar en el escritorio la ventana del módulo de Sincronización (desarrollado por el Integrante 3) y verificar la navegación completa.
  * *Definition of Done (DoD):*
    * Ejecución fluida de una demo completa de 5 minutos abriendo terminal, ejecutando procesos, viendo el Task Manager y los filósofos adquiriendo y liberando recursos.
    * Etiquetado de versión `v0.5.0-beta` en Git.

---

### 👤 INTEGRANTE 2: Ingeniero de CPU y Memoria (Core)

* **Tecnología:** C puro (C99/C11, `gcc`, `make`, Valgrind)

* **Objetivo:** Construir el motor determinista de algoritmos de planificación de procesos (Semana 4) y el asignador/traductor de memoria RAM (Semana 6).

#### Sprint 1 (12–16/10 · Días 1–5): Módulo 1 - Planificación de Procesos (Semana 4 del Curso)

* **Tarea T2.0: Creación, Ejecución y Terminación de Procesos e Hilos (M1)**
  * *Qué harás:* El Rol 2 crea un demostrador C que use `fork()`, ejecute un programa controlado con `exec()` y recoja su finalización con `waitpid()`. El Rol 4 complementa con un ejemplo Python de dos hilos, `start()`/`join()` y registro de sus identificadores. El Rol 1 habilita una ejecución controlada desde la web o muestra la traza producida por estos demostradores.
  * *Archivos:* `src/core_c/process_demo.c`, `tests_python/thread_demo.py`.
  * *DoD:* Se observa PID/PPID y salida del hijo; no quedan procesos zombis; ambos hilos terminan; se explica la diferencia entre proceso, hilo y PCB simulado. Validado antes del día 10.

* **Tarea T2.1: Modelado de PCB y Bucle de Simulación por Ticks**
  * *Qué harás:* Diseñar el archivo de cabecera `pcb.h` con `struct Process` conteniendo `pid`, `name`, `arrival_time`, `burst_time`, `remaining_time`, `priority`, `state` (`READY`, `RUNNING`, `TERMINATED`). Programar la cola de listos (`struct Queue`).
  * *Archivos a crear:* `src/core_c/pcb.h`, `src/core_c/queue.c`.
  * *Cómo probarlo:* Crear un `main_test.c` que inserte 5 procesos en la cola, avance el reloj y verifique la extracción secuencial.
  * *Definition of Done (DoD):*
    * Compilación con `gcc -Wall -Wextra -Werror` sin ninguna advertencia.
    * Cero fugas de memoria verificado con `valgrind --leak-check=full`.

* **Tarea T2.2: Algoritmos No Apropiativos: FCFS, SJF y Validación de Prioridad (Casos Base)**
  * *Qué harás:* Implementar en C:
    1. **FCFS (First-Come, First-Served):** El primero en llegar se ejecuta hasta terminar.
    2. **SJF (Shortest Job First no apropiativo):** Entre los procesos que ya llegaron, el de menor ráfaga toma la CPU.
    * Mantener también el modo de prioridad no apropiativa descrito en el contrato: elegir el menor valor de prioridad entre los procesos listos y ejecutarlo hasta terminar.
    * Exportar la traza paso a paso al formato `cpu_timeline.json`.
  * *Archivos a crear:* `src/core_c/scheduler_nonpreemptive.c`.
  * *Definition of Done (DoD):*
    * Las salidas de tiempos de espera y retorno coinciden al 100% con los ejercicios resueltos a mano del curso (FCFS: 5.75 min, SJF: 5.25 min).
    * El ejecutable admite banderas: `./scheduler --algo FCFS --file data/procesos_semana4.json`.

* **Tarea T2.3: Algoritmos Apropiativos: Round Robin y Prioridades con Envejecimiento**
  * *Qué harás:* Implementar en C:
    1. **Round Robin (RR):** Si un proceso alcanza el quantum especificado ($quantum=4\text{ min}$), se desaloja y pasa al final de la cola de listos.
    2. **Prioridades Apropiativas:** Si llega un proceso con mayor prioridad (menor número entero), interrumpe al proceso en ejecución. Incorporar envejecimiento (*aging*) para que $P_2$ no sufra inanición.
  * *Archivos a crear:* `src/core_c/scheduler_preemptive.c`.
  * *Definition of Done (DoD):*
    * Conmutación de contexto exacta cuando expira el quantum o llega un proceso con mayor prioridad efectiva. Definir y documentar el intervalo de envejecimiento y cómo mejora la prioridad de un proceso en espera, conservando la convención de que un número menor representa mayor prioridad.
    * RR produce exactamente 9.25 min de espera promedio y 4.0 min de respuesta. El valor 7.75 min corresponde a prioridad **no apropiativa** y se valida en esa modalidad. La prioridad apropiativa y el envejecimiento se validan por separado contra el oráculo de Python, usando la misma política configurada. Validación aprobada por el Integrante 4.

#### Sprint 2 (17–21/10 · Días 6–10): Módulo 3 - Gestión de Memoria (Semana 6 del Curso)

Las tareas de asignación contigua empiezan los días 4 y 5 para adelantar trabajo; paginación, segmentación y la integración se cierran durante este sprint. El reemplazo lo desarrolla el Rol 4 en paralelo.

* **Tarea T2.4: Asignación Contigua de Memoria (Bloques y Huecos Libres)**
  * *Qué harás:* Modelar la memoria como un arreglo de bloques, cada uno con dirección de inicio, tamaño y estado libre u ocupado. Implementar:
    * `alloc_first_fit(pid, size_kb)`: Asigna en el primer bloque libre donde quepa.
    * `alloc_best_fit(pid, size_kb)`: Asigna en el hueco más ajustado para minimizar desperdicio.
    * `free_memory(pid)`: Libera los bloques y fusiona huecos contiguos.
  * *Archivos a crear:* `src/core_c/memory_contiguous.c`.
  * *Definition of Done (DoD):*
    * Ante una solicitud de 3 unidades contiguas con 6 unidades libres distribuidas en huecos menores de 3, el sistema rechaza la asignación y emite una alerta de *Fragmentación Externa*. Registrar el mapa inicial, la solicitud y el mapa resultante para reproducir el caso.
    * Emite un mapa de bloques en JSON: `[{"start": 0, "size": 128, "pid": 1, "free": false}, ...]`.

* **Tarea T2.5: Paginación y Traductor de Direcciones MMU (Cálculo Lámina 6)**
  * *Qué harás:* Dividir la memoria en marcos de 4096 Bytes. Crear la tabla de páginas por proceso y la función de traducción `translate_virtual_to_physical()`.
  * *Archivos a crear:* `src/core_c/memory_paging.c`.
  * *Definition of Done (DoD):*
    * Para la dirección virtual `10000` con marco `7`, la función calcula de forma determinista **`30480`**.
    * Generación del ejecutable final consolidado `build/core_engine`.

* **Tarea T2.6: Segmentación con Base y Límite (Detección de Trap)**
  * *Qué harás:* Modelar los segmentos de Código (base 1000, límite 400), Datos (base 5200, límite 800) y Pila (base 9000, límite 300) según lámina 7. Interpretar el límite como longitud del segmento: aceptar únicamente desplazamientos `0 <= offset < limite`. Un desplazamiento igual o mayor al límite genera una excepción simulada (*trap*).
  * *Definition of Done (DoD):*
    * Acceso `(1, 300)` retorna `5500`. Acceso `(1, 900)` lanza `EXCEPTION_SEGMENTATION_FAULT`.

---

### 👤 INTEGRANTE 3: Ingeniero de Concurrencia y Sincronización

* **Tecnología:** Go (`goroutines`, `channels`, paquetes `sync` y `sync/atomic`)

* **Objetivo:** Modelar los mecanismos de sincronización de la Semana 5 (buffer productor-consumidor y filósofos comensales), emitiendo trazas de eventos paso a paso.

#### Sprint 1 (12–16/10 · Días 1–5): Primitivas y Productor-Consumidor (Semana 5 del Curso)

* **Tarea T3.1: Abstracción de Semáforo Contador y Mutex en Go**
  * *Qué harás:* Crear una estructura `CountingSemaphore` utilizando canales y tipos atómicos de Go, con métodos:
    * `Wait()` / `P()`: Si hay un permiso disponible, lo consume y continúa; si no hay ninguno, bloquea hasta obtener uno. La espera y la adquisición deben ser atómicas.
    * `Signal()` / `V()`: Incrementa el contador y despierta a una goroutine en espera.
  * *Archivos a crear:* `src/sync_go/semaphore.go`.
  * *Cómo probarlo:* Crear un test concurrente con 15 goroutines compitiendo por un recurso limitado a 3 accesos concurrentes.
  * *Definition of Done (DoD):*
    * Pasa el detector de carreras nativo de Go: `go test -race -v ./src/sync_go`.
    * En ningún momento el número de goroutines activas en la sección crítica supera el límite del semáforo.

* **Tarea T3.2: Simulación de Productor-Consumidor con Buffer Acotado**
  * *Qué harás:* Construir el modelo con un buffer compartido de 5 elementos, 2 productores y 2 consumidores (20 elementos producidos y consumidos en total).
    * Controlado por 3 semáforos: `mutex=1`, `vacios=5` y `llenos=0`.
    * Respetar el orden de adquisición: productores esperan `vacios` antes de `mutex`; consumidores esperan `llenos` antes de `mutex`. Tras modificar el buffer, liberar `mutex` y señalizar `llenos` o `vacios`, respectivamente.
    * Emitir eventos con marca de tiempo, identificador de proceso y ocupación del buffer; registrar `BLOCKED` cuando un productor espera espacio o un consumidor espera datos.
  * *Archivos a crear:* `src/sync_go/prod_cons.go`.
  * *Definition of Done (DoD):*
    * Verificación de que no existe sobreescritura (overflow) ni extracción en vacío (underflow).
    * Si el buffer se llena, los productores pasan a estado `BLOCKED` comprobable.

#### Sprint 2 (17–21/10 · Días 6–10): Filósofos Comensales y Detector de Deadlock

* **Tarea T3.3: Problema de los Filósofos Comensales**
  * *Qué harás:* Modelar 5 filósofos como procesos que alternan entre `PENSANDO`, `CON_HAMBRE` y `COMIENDO`, compitiendo por 5 recursos circulares. Estos son los estados convencionales del problema de sincronización; la implementación representa adquisición, espera y liberación de recursos. Implementar dos modalidades conmutables:
    1. *Modalidad con Deadlock:* Cada filósofo toma primero el recurso izquierdo. Una barrera coordina la adquisición inicial para reproducir la espera circular: cada proceso queda esperando un recurso retenido por otro.
    2. *Modalidad Preventiva:* Romper la simetría (filósofos impares toman primero el recurso derecho) para eliminar la espera circular.
  * *Archivos a crear:* `src/sync_go/philosophers.go`.
  * *Definition of Done (DoD):*
    * En la modalidad con deadlock, se demuestra el bloqueo total en menos de 5 segundos.
    * En la modalidad preventiva, la simulación completa 10,000 ciclos sin interbloqueos; se registra cuántos avances realiza cada filósofo para comprobar que ninguno quedó sin progresar durante la prueba. Esta prueba aporta evidencia de progreso, pero no demuestra por sí sola ausencia de inanición para toda ejecución posible.

* **Tarea T3.4: Motor de Detección de Deadlock (Ciclos en Grafos RAG)**
  * *Qué harás:* Implementar un algoritmo que revise el Grafo de Asignación de Recursos (RAG). Para recursos con una única instancia, un ciclo identifica un interbloqueo y genera una alerta JSON con los PIDs involucrados. El detector de este sprint se aplica a esos recursos; un ciclo no basta para concluir interbloqueo si un tipo de recurso tiene múltiples instancias.
  * *Archivos a crear:* `src/sync_go/deadlock_detector.go`.
  * *Definition of Done (DoD):*
    * Identifica y lista con precisión los PIDs que forman el ciclo: `{"deadlock": true, "cycle": [1, 2, 3, 4, 5]}`.
    * Integrado como endpoint en Go para alimentar la interfaz visual del Integrante 1.

---

### 👤 INTEGRANTE 4: Ingeniero de QA, Oráculo Teórico y Analítica

* **Tecnología:** Python 3 (`pytest`, `pandas`, `matplotlib`, reemplazo de páginas y scripts de validación)

* **Objetivo:** Respaldar el 30% de la nota de resultados correctos y el 15% de análisis comparativo. Construir el oráculo teórico y auditar los motores de C y Go frente a las prácticas del profesor.

#### Sprint 1 (12–16/10 · Días 1–5): Oráculo Teórico de Planificación de CPU

* **Tarea T4.1: Oráculo Teórico de CPU en Python (Ground Truth Semana 4)**
  * *Qué harás:* Escribir un script en Python puro que tome los procesos $P_1$ a $P_4$ y calcule con precisión matemática los tiempos de inicio, finalización, espera ($W_i$), retorno ($T_i$) y sus promedios para FCFS, SJF, Round Robin y Prioridades. Diferenciar prioridad no apropiativa, apropiativa y apropiativa con envejecimiento, utilizando la misma configuración que el motor C.
  * *Archivos a crear:* `tests_python/ground_truth_semana4.py`.
  * *Cómo probarlo:* Comparar los resultados del script contra la lámina 10 de la Semana 4.
  * *Definition of Done (DoD):*
    * Los cálculos de $W_{prom}$ y $T_{prom}$ coinciden con exactitud de 3 decimales frente a los ejercicios del profesor (FCFS: 5.75, SJF: 5.25, RR: 9.25, Prioridad: 7.75).
    * Funciona para casos con tiempos de llegada iguales o escalonados.

* **Tarea T4.2: Suite de Pruebas Automatizadas de Caja Negra para C**
  * *Qué harás:* Crear una suite en `pytest` que:
    1. Genere 10 archivos de procesos de prueba (incluyendo `procesos_semana4.json` y casos límite).
    2. Ejecute el binario de C del Integrante 2 con cada caso.
    3. Lea la salida `cpu_timeline.json` y valide que cada número sea idéntico al del oráculo de Python.
  * *Archivos a crear:* `tests_python/test_scheduler.py`.
  * *Definition of Done (DoD):*
    * El comando `pytest tests_python/test_scheduler.py` ejecuta todas las pruebas y reporta `10 passed` en verde.
    * Si el motor de C tiene un desfase de 1 tick, el test falla y muestra claramente la diferencia.

#### Sprint 2 (17–21/10 · Días 6–10): Reemplazo de Páginas, Gantt e Informe de Memoria

* **Tarea T4.3: Generador de Diagramas de Gantt Automatizados**
  * *Qué harás:* Escribir un script en Python que lea el JSON de salida del planificador y dibuje automáticamente el diagrama de Gantt exacto de las láminas 6 a 9 en formato de imagen (PNG/SVG) o tabla HTML para que el Integrante 1 lo muestre en la web.
  * *Archivos a crear:* `tests_python/generate_gantt.py`.
  * *Definition of Done (DoD):*
    * El gráfico incluye colores diferenciados por proceso, eje de tiempo rotulado, bloques de espera y cuadro de métricas promedio.

* **Tarea T4.4: Auditoría del Taller de Memoria e Informe Comparativo**
  * *Qué harás:* Validar los cálculos de paginación del Integrante 2 ($10\,000 \to 30\,480$) y redactar el análisis comparativo del informe Beta: asignación contigua, paginación, segmentación y fallos de FIFO/LRU/Óptimo. Añadir trazas y parámetros suficientes para repetir cada prueba.
  * *Archivos a crear:* `docs/informe_tecnico_beta.md`.
  * *Definition of Done (DoD):*
    * Informe riguroso con tablas comparativas, gráficos de Gantt incrustados y fundamentación teórica según los libros guía (Silberschatz / Tanenbaum).

* **Tarea S3.1 (Reprogramada al Hito de 10 Días · Rol 4 - Python; revisión Rol 2): Reemplazo de Páginas FIFO, LRU y Óptimo (M3)**
  * *Qué harás:* Implementar los tres algoritmos en `tests_python/page_replacement.py`, iniciarlos en Sprint 1 y cerrar su integración en Sprint 2. Recibir patrón de referencias y cantidad de marcos; devolver traza de marcos, aciertos y fallos. Óptimo utiliza el conocimiento del patrón futuro únicamente como referencia teórica.
  * *Integración:* Rol 1 expone `/api/v1/memory/replace`; Rol 2 comprueba la coherencia con páginas y marcos del módulo de memoria. Rol 4 compara los tres sobre entradas idénticas y documenta las diferencias.
  * *DoD:* Pruebas reproducibles, tabla y traza visibles en web. Usar `1,2,3,4,1,2,5,1,2,3,4,5` para demostrar FIFO con 9 fallos en 3 marcos y 10 en 4 (Belady). Óptimo no supera los fallos de los otros algoritmos sobre la misma entrada; puede empatar. Cerrado el **21/10**.

---

## 4. Fase 2: Módulos 4, 5 y 6 y Cierre (22/10–09/12)

La segunda fase conserva las funcionalidades restantes del plan y completa los contenidos faltantes de la propuesta.

| Fechas de 2026 | Sprint | Módulos y resultado de cierre |
|---|---|---|
| 22/10–04/11 · 14 días | Sprint 3: VFS y asignación de archivos | Base de M4; archivos operativos, tres métodos de asignación, terminal/editor. Reemplazo ya entregado; aquí solo regresiones. |
| 05/11–18/11 · 14 días | Sprint 4: E/S y seguridad | M4 y M5 completos: sondeo/interrupciones/DMA, buffering, permisos, autenticación y cifrado; comparación conceptual de virtualización. |
| 19/11–09/12 · 21 días | Sprint 5: Distribuidos y entrega final | M6 completo al 25/11; luego estrés, informe, empaquetado y sustentación de los seis módulos. |

### 4.1 Sprint 3 (22/10–04/11): VFS y Asignación de Archivos (M4)

**Propósito:** Construir el sistema de archivos virtual y sus métodos de asignación, manteniendo pruebas de regresión del reemplazo entregado en la Beta.

**Desarrollo:** 22–28/10: definir inodos, directorios y disco virtual; implementar operaciones y asignación de bloques. 29/10–04/11: conectar terminal/editor y comparar los tres métodos de asignación. El Rol 4 revisa resultados y regresiones M1–M3.

**Cierre:** Un archivo guardado en el editor se consulta desde terminal; los métodos contiguo, enlazado e indexado se ejecutan y comparan. La persistencia del VFS se limita a la sesión activa. El reemplazo de páginas permanece operativo.

* **Tarea S3.2 (Rol 2 - C): Sistema de Archivos Simulado (Inodos y Bloques)**
  * *Detalle:* Estructura de inodos en memoria, mapa de bits de bloques y soporte para crear, leer y borrar archivos (`vfs_write`, `vfs_read`).
  * *DoD:* Creación y lectura de archivos de hasta 64 KB sin fugas de bloques de disco virtual.

* **Tarea S3.3 (Rol 1 - Go): Comandos VFS en la Terminal Web**
  * *Detalle:* Habilitar comandos de gestión de archivos en la consola: `ls -l`, `mkdir`, `cd`, `cat`, `rm`.
  * *DoD:* Navegación en directorios anidados con persistencia durante la sesión activa.

* **Tarea S3.4 (Rol 3 - Go): Editor de Texto (Bloc de Notas) Conectado al VFS**
  * *Detalle:* Ventana gráfica de Bloc de Notas que permite escribir y guardar archivos de texto dentro del sistema de archivos simulado.
  * *DoD:* Archivo guardado desde el Bloc de Notas es visible y legible desde la terminal con `cat`.

* **Tarea S3.5 (Rol 2 - C; validación Rol 4): Asignación Contigua, Enlazada e Indexada (M4)**
  * *Detalle:* Usar el mismo disco virtual y operaciones de archivo con una política seleccionable. Contigua reserva una secuencia de bloques; enlazada mantiene el siguiente bloque; indexada conserva una tabla de referencias. El modelo de inodos se mantiene como metadatos y no sustituye la comparación de estos métodos.
  * *DoD:* Crear, leer y borrar el mismo archivo con los tres métodos; comprobar contenido y recuperación de bloques. Documentar fragmentación, metadatos y acceso secuencial/directo. Sin implementar un sistema de archivos real.

### 4.2 Sprint 4 (05/11–18/11): E/S, Spooling y Seguridad (M4 y M5)

**Propósito:** Completar las modalidades de E/S y el control de acceso, autenticación y cifrado de archivos.

**Desarrollo:** 05–11/11: ráfagas de E/S, sondeo/interrupciones/DMA, buffering y spooler; en paralelo, autenticación y permisos. 12–18/11: cifrado, comparación VM/contenedor, benchmarks e integración. Los permisos se comprueban en servidor y VFS, incluyendo editor y terminal.

**Cierre:** Se comparan sondeo/interrupciones/DMA, los trabajos no se mezclan, los accesos no autorizados se rechazan y un archivo se cifra/descifra de forma verificable. Se entrega la comparación conceptual de virtualización.

* **Tarea S4.1 (Rol 2 - C): Ráfagas de E/S e Interrupciones**
  * *Detalle:* Procesos con ráfagas compuestas: cuando un proceso requiere E/S, pasa al estado `BLOCKED` y cede la CPU; una interrupción lo devuelve a `READY`.
  * *DoD:* El planificador no deja la CPU ociosa mientras haya procesos listos compitiendo con procesos esperando E/S.

* **Tarea S4.2 (Rol 3 - Go): Simulación de Dispositivos y Cola de Impresión (Spooler)**
  * *Detalle:* Simulación de una impresora virtual donde múltiples procesos envían trabajos de impresión concurrentemente a una cola (*spooler*).
  * *DoD:* Las líneas de texto de diferentes trabajos se imprimen ordenadas sin mezclarse entre sí.

* **Tarea S4.3 (Rol 1 - Go): Sistema de Usuarios y Permisos UNIX (`rwx`)**
  * *Detalle:* Usuarios `root` y `alumno`. Permisos octales (`755`, `644`) sobre inodos. Comandos `su`, `chmod`.
  * *DoD:* Un usuario estándar recibe `Permission denied` al intentar modificar archivos de `root`.

* **Tarea S4.4 (Rol 4 - Python): Benchmarking de Rendimiento con E/S**
  * *Detalle:* Evaluar cómo varía el Throughput y el tiempo de respuesta según el tamaño del quantum en presencia de ráfagas frecuentes de E/S.
  * *DoD:* Curvas de rendimiento generadas en Python incluidas en el informe final.

* **Tarea S4.5 (Rol 3 - Go; conexión Rol 2, medición Rol 4): Sondeo, Interrupciones, DMA y Buffering (M4)**
  * *Detalle:* Simular el mismo dispositivo y transferencia con tres modos. Sondeo registra consultas periódicas de CPU; interrupciones generan un evento al completar E/S; DMA modela una transferencia por controlador con intervención inicial/final de CPU. Usar tiempos y contadores simulados y un buffer de capacidad definida.
  * *DoD:* Traza y comparación de intervenciones de CPU, tiempo simulado y ocupación de buffer para la misma entrada. No hay desbordamiento ni datos perdidos. DMA es simulado, no acceso a hardware real.

* **Tarea S4.6 (Rol 1 - Go; pruebas Rol 4): Autenticación y Control de Acceso (M5)**
  * *Detalle:* Incorporar inicio/cierre de sesión de los usuarios didácticos `root` y `alumno`, verificar contraseña mediante hash de una biblioteca y comprobar identidad y permisos en cada operación del VFS. `su` requiere autenticación del usuario destino; no basta cambiar un campo en la interfaz.
  * *DoD:* Rechazar credenciales incorrectas y operaciones sin sesión/autorización; conservar aislamiento de permisos entre terminal y editor. No almacenar contraseñas en texto plano. Pruebas positivas y negativas reproducibles.

* **Tarea S4.7 (Rol 4 - Python; integración Rol 1): Cifrado/Descifrado Simétrico y Asimétrico (M5)**
  * *Detalle:* Utilizar una biblioteca existente, por ejemplo `cryptography`, para una demostración simétrica autenticada y otra asimétrica sobre un archivo pequeño de prueba. En el caso asimétrico, explicar los límites de tamaño y usar un archivo que quepa en la operación elegida. Go conecta las acciones al archivo virtual mediante entrada/salida controlada. No diseñar algoritmos criptográficos propios.
  * *DoD:* Recuperar exactamente los bytes originales; mostrar que el contenido cifrado difiere y que una clave incorrecta falla. Guardar en el informe la secuencia reproducible y la diferencia entre clave compartida y par pública/privada.

* **Tarea S4.8 (Roles 3 y 4; revisión del equipo): Nociones de Virtualización (M5)**
  * *Detalle:* Añadir una tabla VM vs. contenedor sobre kernel, aislamiento y recursos, y explicar dónde se ejecuta YaKcheOS. El simulador representa mecanismos de SO; no constituye por sí mismo una VM ni un hipervisor.
  * *DoD:* Comparación conceptual y diagrama sencillo incorporados al informe. La propuesta exige nociones, no construir un hipervisor ni un sistema de contenedores.

### 4.3 Sprint 5 (19/11–09/12): Sistemas Distribuidos, Estrés y Sustentación (M6 y Cierre)

**Propósito:** Completar el intercambio de mensajes de M6 y consolidar la entrega integrada de los seis módulos.

**Desarrollo:** 19–25/11: implementar e integrar dos nodos con paso de mensajes; mientras tanto, otros roles ejecutan estrés y redactan el informe. 26/11–02/12: corregir hallazgos, cerrar documentación y empaquetado. 03–09/12: regresión de los seis módulos, demo y ensayo de sustentación de **10–15 minutos**, conforme a la propuesta.

**Cierre:** Seis módulos demostrables, arranque reproducible, pruebas documentadas y dominio de la integración por los cuatro integrantes.

* **Tarea S5.1 (Rol 1 - Go): Pantalla de Arranque (Boot) y Despliegue Portable**
  * *Detalle:* Secuencia de inicio estética simulada y script de arranque en un solo paso (`go run ./cmd/server`).
  * *DoD:* Carga en navegador en $<2\text{ segundos}$ en el entorno de prueba documentado, con el artefacto previamente compilado. Entregar instrucciones de compilación para C/Go o binarios compatibles con las plataformas soportadas; `go run ./cmd/server` requiere Go instalado y el motor C disponible.

* **Tarea S5.2 (Roles 2 y 3 - C/Go): Batería de Pruebas de Estrés**
  * *Detalle:* Ejecución con 50 procesos simultáneos, memoria al 99% y concurrencia masiva.
  * *DoD:* Cero fallos de segmentación (*segfaults*) o caídas de servidor (*panics*).

* **Tarea S5.3 (Rol 4 - Python): Compilación del Dossier Final y Guía de Demostración**
  * *Detalle:* Documento consolidado (15-20 págs) con la fundamentación teórica de cada módulo, tablas comparativas y paso a paso para el jurado.
  * *DoD:* Memoria técnica en PDF formateada según los estándares de la facultad.

* **Tarea S5.4 (Equipo Completo): Ensayo General de Sustentación Oral**
  * *Detalle:* Simulación de defensa de 10–15 minutos con preguntas teóricas del profesor (explicación de algoritmos, fórmulas de espera, prevención de interbloqueos y traducción de memoria).
  * *DoD:* Los 4 integrantes son capaces de sustentar con soltura tanto su módulo como la integración general.

* **Tarea S5.0 (Rol 3 - Go; integración Rol 1; validación Rol 4): Dos Nodos y Paso de Mensajes (M6)**
  * *Detalle:* Ejecutar dos procesos Go independientes en localhost, cada uno con identificador y puerto distintos. Enviar un mensaje de A a B y una respuesta de B a A mediante TCP/HTTP con JSON. Registrar origen, destino, identificador de mensaje y contenido; fijar un timeout para evitar espera infinita.
  * *Archivos:* `cmd/node/main.go`, `src/distributed_go/messages.go`, `data/distributed_case.json`, `tests_python/test_distributed.py`.
  * *DoD:* Dos procesos activos intercambian mensajes verificables; la interfaz muestra las trazas. Si B se detiene, A informa el error/timeout sin bloquear el servidor. No se requieren máquinas externas, consenso ni replicación. **Entrega M6: 25/11**.

---

## 5. Cuadro de Control y Puntos de Verificación

| Fecha / hito | Módulos | Evidencia de cierre |
|---|---|---|
| **16/10 · Día 5** | Base M1–M3 | CPU y sincronización ejecutables; reemplazo implementado; primera integración web. |
| **21/10 · Día 10** | **M1, M2 y M3 completos — primera mitad** | Procesos/hilos; métricas CPU; concurrencia correcta; asignación/traducción/fragmentación; FIFO/LRU/Óptimo y comparación; informe y demo Beta. |
| **04/11** | VFS de M4 | Terminal/editor y asignación contigua/enlazada/indexada; regresiones M1–M3. |
| **18/11** | **M4 y M5 completos** | Sondeo/interrupciones/DMA y buffering; archivos/permisos; autenticación; cifrado/descifrado; comparación VM/contenedor. |
| **25/11** | **M6 completo** | Dos procesos/nodos intercambian mensajes y reportan fallo de conexión. |
| **02/12** | Versión final candidata | Todos los módulos integrados; estrés e informe listos para revisión. |
| **09/12** | **Entrega interna final propuesta** | Ejecutable, informe, instrucciones y sustentación ensayada de 10–15 minutos. |

### 5.1 Coordinación para Cumplir el Día 10

- Al empezar, acordar esquema JSON, unidades, errores y criterios de aceptación. No esperar a que la interfaz esté terminada para validar los algoritmos.
- Al cerrar cada día, integrar los cambios que pasaron revisión y ejecutar las pruebas de los módulos modificados. Repartir correcciones entre roles y mantener visibles los pendientes del hito.
- Si al día 5 no funcionan las bases de CPU, sincronización y reemplazo, reasignar apoyo al módulo bloqueado y trasladar refinamientos visuales al Sprint 5. No declarar completa la primera mitad omitiendo contenidos obligatorios.
- En los días 9 y 10 priorizar correcciones y evidencias de M1–M3; evitar incorporar funcionalidades ajenas a esos módulos.
- Los plazos organizan el trabajo; confirmar la disponibilidad real del equipo antes de comprometer horas o afirmar que la fecha está garantizada.

### 5.2 Relación con las Entregas Académicas de la Propuesta

| Evaluación de la propuesta | Contenido exigido | Evidencia que produce este plan |
|---|---|---|
| Semana 4 del curso · Laboratorio 1 | M1 completo | Código, pruebas y métricas del hito de 10 días. |
| Semana 7 del curso · Trabajo Domiciliario 1 | M2 y M3, incluido reemplazo | Sincronización y memoria completas con informe comparativo al 21/10. |
| Semana 10 del curso · Laboratorio 2 | M4 completo | E/S y archivos completos al 18/11. |
| Semana 15 del curso · Trabajo Domiciliario 2 y sustentación | M1–M6 integrados | Paquete final interno previsto para el 09/12 y demo de 10–15 minutos. |

Las fechas de desarrollo propuestas no prueban que coincidan con esas semanas del curso: contrastar las fechas académicas con el calendario del docente. Si la entrega oficial final es anterior al 09/12, adelantar el cierre interno y reasignar tareas; no desplazar automáticamente la evaluación del curso.

**Documentación y evaluación:** Mantener arquitectura, instrucciones, resultados comparativos y evidencia Git de los cuatro integrantes. Rúbrica de la propuesta: correctitud 30%, calidad de código 15%, documentación 15%, análisis 15%, sustentación 15% y trabajo en equipo/plazos 10%. En los laboratorios, la propuesta redistribuye el peso de sustentación entre los demás criterios.
