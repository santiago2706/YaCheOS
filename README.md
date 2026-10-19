# YaKcheOS - Simulador Web de Sistema Operativo

**Universidad Nacional de Ingeniería (UNI)**  
**Facultad de Ingeniería Industrial y de Sistemas (FIIS) · Escuela Profesional de Ingeniería de Software**  
**Curso:** SW407 - Sistemas Operativos (2026-II)  
**Proyecto:** YaKcheOS (Simulador Educativo de Sistema Operativo en Web)  
**Stack de Desarrollo:** C (Core y Memoria), Go (Orquestación Web y Sincronización), Python (QA, Oráculo y Analítica).

---

## 🚀 Instrucciones de Arranque Rápido

Para levantar el simulador en tu navegador, ejecuta un único comando desde la raíz del proyecto:

```bash
go run ./cmd/server
```

Abre en tu navegador web:
👉 **[http://localhost:8080](http://localhost:8080)**

Para ejecutar las pruebas del servidor en Go:
```bash
go test -v ./cmd/server/...
```

---

## 🛠️ Estado del Proyecto: Sprint 1 (Días 1–5)

### 👤 Integrante 1: Líder Técnico & Orquestador Web (Completado)
* **T1.1: Inicialización de Repositorio y Servidor Base en Go:**
  - Servidor HTTP nativo en Go (`net/http`) en el puerto `:8080`.
  - Módulo Go configurado (`yakcheos`, Go 1.21).
  - Reglas de `.gitignore` para ignorar `.exe`, binarios y trazas de simulación (`cpu_timeline.json`).
  - Pruebas unitarias de endpoint raíz y salud (`cmd/server/main_test.go`).
* **T1.2: Sistema de Ventanas Flotantes de Escritorio:**
  - Maquetación de escritorio moderno con modo oscuro y *glassmorphism*.
  - Barra de tareas inferior con menú de inicio, lista de tareas activas y reloj digital en tiempo real.
  - Gestor de ventanas ligero (`web/js/app.js`): mover por arrastre de barra superior, minimizar a barra de tareas, maximizar/restaurar y cerrar.
  - Soporte de ventanas simultáneas sin bloqueos en el navegador (Terminal, Task Manager, Sincronización, Acerca de).
* **T1.3: Emulador de Terminal CLI en Web (Taller Semana 2):**
  - Pantalla negra y prompt interactivo `yakcheos:~$`.
  - Historial de comandos navegable con las flechas de teclado `Arriba` y `Abajo`.
  - Despacho de comandos del laboratorio:
    - `help`: Lista de comandos disponibles.
    - `clear`: Limpieza de pantalla.
    - `ps`: Consulta y lista de procesos activos desde el contrato JSON.
    - `run <algo> <archivo.json>`: Simulación de CPU (FCFS, SJF, RR, PRIORIDAD) con cálculo de métricas.
    - `ls`, `mkdir`, `cat`: Reconocimiento de sintaxis de sistema de archivos (preparatorio para VFS en Sprint 3).
    - `uname -a`, `date`, `echo`, `open <ventana>`.

---

## 📁 Estructura del Repositorio

```text
YaKcheOS/
├── cmd/
│   └── server/
│       ├── main.go                 # Servidor HTTP Go en :8080
│       └── main_test.go            # Pruebas HTTP del servidor
├── data/                           # Contratos JSON base acordados el Día 1
│   ├── procesos_semana4.json       # Caso base Semana 4 (P1 a P4)
│   ├── sincronizacion_semana5.json # Contrato Semana 5 (Buffer y Filósofos)
│   ├── memoria_semana6.json        # Contrato Semana 6 (Paginación y Segmentación)
│   └── page_references.json        # Secuencia para reemplazo de páginas
├── docs/                           # Informes y arquitectura
├── src/
│   ├── core_c/                     # [Integrante 2: C] Algoritmos de CPU y Memoria
│   └── sync_go/                    # [Integrante 3: Go] Concurrencia y Sincronización
├── tests_python/                   # [Integrante 4: Python] Suite de validación y oráculo
└── web/
    ├── index.html                  # Escritorio + barra de tareas + ventanas
    ├── css/
    │   ├── desktop.css             # Estilos de escritorio y ventanas
    │   └── terminal.css            # Estilos de la terminal CLI
    └── js/
        ├── app.js                  # Gestor de ventanas flotantes
        └── terminal.js             # Lógica CLI e historial
```
