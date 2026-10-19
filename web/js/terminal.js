/**
 * YaKcheOS - Terminal Emulator CLI (Sprint 1 - T1.3)
 * Autor: Integrante 1 (Líder Técnico & Orquestador Web)
 * Emulador de consola con prompt 'yakcheos:~$', historial con flechas y despacho de comandos del lab.
 */

class TerminalCLI {
  constructor() {
    this.history = [];
    this.historyIndex = -1;
    this.inputEl = document.getElementById('terminal-input');
    this.historyEl = document.getElementById('terminal-history');
    this.screenEl = document.getElementById('terminal-screen');
    this.rootEl = document.getElementById('terminal-root');

    this.init();
  }

  init() {
    if (!this.inputEl || !this.historyEl) return;

    // Foco automático al hacer clic en cualquier parte de la terminal
    if (this.rootEl) {
      this.rootEl.addEventListener('click', () => {
        this.inputEl.focus();
      });
    }

    this.inputEl.addEventListener('keydown', (e) => this.handleKeyDown(e));
  }

  handleKeyDown(e) {
    if (e.key === 'Enter') {
      const rawCmd = this.inputEl.value.trim();
      this.executeCommand(rawCmd);
      if (rawCmd.length > 0) {
        this.history.push(rawCmd);
        this.historyIndex = this.history.length;
      }
      this.inputEl.value = '';
      this.scrollToBottom();
    } else if (e.key === 'ArrowUp') {
      // Navegación hacia atrás en el historial
      e.preventDefault();
      if (this.history.length > 0 && this.historyIndex > 0) {
        this.historyIndex--;
        this.inputEl.value = this.history[this.historyIndex];
      }
    } else if (e.key === 'ArrowDown') {
      // Navegación hacia adelante en el historial
      e.preventDefault();
      if (this.historyIndex < this.history.length - 1) {
        this.historyIndex++;
        this.inputEl.value = this.history[this.historyIndex];
      } else {
        this.historyIndex = this.history.length;
        this.inputEl.value = '';
      }
    }
  }

  executeCommand(rawCmd) {
    // Si se pulsa enter sin comando, solo agregar un prompt vacío
    if (!rawCmd) {
      this.appendEntry('', '');
      return;
    }

    const parts = rawCmd.split(/\s+/);
    const cmd = parts[0].toLowerCase();
    const args = parts.slice(1);

    switch (cmd) {
      case 'help':
        this.appendEntry(rawCmd, this.getHelpOutput());
        break;

      case 'clear':
      case 'cls':
        this.historyEl.innerHTML = '';
        break;

      case 'ps':
        this.handlePsCommand(rawCmd);
        break;

      case 'run':
        this.handleRunCommand(rawCmd, args);
        break;

      case 'ls':
        this.appendEntry(rawCmd, `
<div class="term-cyan">cmd/   data/   docs/   src/   tests_python/   web/   go.mod   README.md</div>
<div class="term-gray" style="margin-top:4px;">[Aviso VFS]: Sintaxis reconocida (Semana 2). La navegacion en profundidad se habilitara con el VFS en Sprint 3.</div>
`);
        break;

      case 'mkdir':
        if (args.length === 0) {
          this.appendEntry(rawCmd, `<span class="term-rose">mkdir: falta el operando de directorio. Uso: mkdir &lt;nombre&gt;</span>`);
        } else {
          this.appendEntry(rawCmd, `
<span class="term-green">Directorio simulado '${this.escapeHtml(args[0])}' registrado.</span>
<div class="term-gray">[Aviso VFS]: La persistencia en disco virtual se habilitara con el VFS en Sprint 3.</div>
`);
        }
        break;

      case 'cat':
        this.handleCatCommand(rawCmd, args);
        break;

      case 'uname':
        this.appendEntry(rawCmd, `<span class="term-cyan">YaKcheOS 0.1.0-alpha #1 SMP PREEMPT 2026 x86_64 UNI-FIIS GNU/Linux</span>`);
        break;

      case 'date':
        this.appendEntry(rawCmd, `<span class="term-amber">${new Date().toString()}</span>`);
        break;

      case 'echo':
        this.appendEntry(rawCmd, `<span>${this.escapeHtml(args.join(' '))}</span>`);
        break;

      case 'kill':
        this.handleKillCommand(rawCmd, args);
        break;

      case 'mmu':
        this.handleMmuCommand(rawCmd, args);
        break;

      case 'seg':
        this.handleSegCommand(rawCmd, args);
        break;

      case 'replace':
        this.handleReplaceCommand(rawCmd, args);
        break;

      case 'sync':
        this.handleSyncCommand(rawCmd, args);
        break;

      case 'open':
        this.handleOpenCommand(rawCmd, args);
        break;

      default:
        this.appendEntry(
          rawCmd,
          `<span class="term-rose">yakcheos: comando no encontrado: '${this.escapeHtml(cmd)}'. Escribe '<span class="term-cyan">help</span>' para ver la lista de comandos disponibles.</span>`
        );
        break;
    }
  }

  appendEntry(commandText, outputHtml) {
    const entry = document.createElement('div');
    entry.className = 'term-history-entry';

    if (commandText !== '') {
      entry.innerHTML = `
        <div class="term-command-echo">
          <span class="prompt-user">yakcheos</span><span class="prompt-at">@</span><span class="prompt-host">uni</span><span class="prompt-colon">:</span><span class="prompt-path">~</span><span class="prompt-symbol">$</span>
          <span>${this.escapeHtml(commandText)}</span>
        </div>
        <div class="term-output">${outputHtml}</div>
      `;
    } else {
      entry.innerHTML = `
        <div class="term-command-echo">
          <span class="prompt-user">yakcheos</span><span class="prompt-at">@</span><span class="prompt-host">uni</span><span class="prompt-colon">:</span><span class="prompt-path">~</span><span class="prompt-symbol">$</span>
        </div>
      `;
    }

    this.historyEl.appendChild(entry);
    this.scrollToBottom();
  }

  getHelpOutput() {
    return `
<div class="term-bold term-cyan">Comandos Disponibles en YaKcheOS CLI (Módulos 1, 2 y 3 - Sprint 2):</div>
<table class="term-table">
  <thead>
    <tr>
      <th style="width: 220px;">Comando</th>
      <th>Descripcion</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><span class="term-green">help</span></td>
      <td>Lista todos los comandos disponibles del simulador.</td>
    </tr>
    <tr>
      <td><span class="term-green">clear</span></td>
      <td>Limpia la pantalla de la terminal.</td>
    </tr>
    <tr>
      <td><span class="term-green">ps</span></td>
      <td>Muestra los procesos activos en la simulación (PID, estado, ráfagas, RAM).</td>
    </tr>
    <tr>
      <td><span class="term-green">run &lt;algo&gt; [archivo] [q]</span></td>
      <td>Ejecuta planificación de CPU (FCFS, SJF, RR, PRIORIDAD) invocando el motor C.</td>
    </tr>
    <tr>
      <td><span class="term-green">kill &lt;PID&gt;</span></td>
      <td>Termina un proceso en simulación y libera su bloque de memoria RAM (T1.5).</td>
    </tr>
    <tr>
      <td><span class="term-green">mmu &lt;dir_virtual&gt; [page_size]</span></td>
      <td>Traductor MMU de direcciones virtuales a físicas por paginación (M3).</td>
    </tr>
    <tr>
      <td><span class="term-green">seg &lt;seg_id&gt; &lt;offset&gt;</span></td>
      <td>Traductor de segmentación con detección de violación de límite / Trap (M3).</td>
    </tr>
    <tr>
      <td><span class="term-green">replace [frames]</span></td>
      <td>Compara fallos FIFO, LRU, Óptimo y verifica la Anomalía de Belády (M3).</td>
    </tr>
    <tr>
      <td><span class="term-green">sync &lt;status|step|deadlock|preventive&gt;</span></td>
      <td>Controla la sincronización concurrente (filósofos y buffer en Go) (M2).</td>
    </tr>
    <tr>
      <td><span class="term-green">open &lt;tasks|sync|about&gt;</span></td>
      <td>Abre la ventana gráfica correspondiente en el escritorio.</td>
    </tr>
    <tr>
      <td><span class="term-green">cat &lt;archivo&gt;</span> | <span class="term-green">ls</span> | <span class="term-green">mkdir</span></td>
      <td>Comandos del sistema de archivos (preparatorio para VFS Semana 3).</td>
    </tr>
    <tr>
      <td><span class="term-green">uname -a</span> | <span class="term-green">date</span> | <span class="term-green">echo</span></td>
      <td>Utilidades estándar del sistema operativo simulado.</td>
    </tr>
  </tbody>
</table>
`;
  }

  async handlePsCommand(rawCmd) {
    try {
      const res = await fetch('/data/procesos_semana4.json');
      const data = await res.json();
      const list = Array.isArray(data) ? data : (data.procesos || []);

      let rows = '';
      list.forEach((p) => {
        const pid = p.pid;
        const name = p.name || p.nombre;
        const arr = p.arrival_time !== undefined ? p.arrival_time : p.llegada;
        const burst = p.burst_time !== undefined ? p.burst_time : p.rafaga;
        const prio = p.priority !== undefined ? p.priority : p.prioridad;
        const mem = p.memory_kb ? p.memory_kb + 'KB' : '128KB';
        rows += `
          <tr>
            <td>${pid}</td>
            <td><span class="term-cyan">${name}</span></td>
            <td>READY</td>
            <td>${arr}m</td>
            <td>${burst}m</td>
            <td>${prio}</td>
            <td>${mem}</td>
          </tr>
        `;
      });

      const output = `
<table class="term-table">
  <thead>
    <tr>
      <th>PID</th>
      <th>CMD</th>
      <th>STAT</th>
      <th>LLEGADA</th>
      <th>BURST</th>
      <th>PRIO</th>
      <th>MEM</th>
    </tr>
  </thead>
  <tbody>
    ${rows}
  </tbody>
</table>
<div class="term-gray" style="margin-top:4px;">Total: ${list.length} procesos cargados en el PCB de simulación.</div>
`;
      this.appendEntry(rawCmd, output);
    } catch (err) {
      this.appendEntry(rawCmd, `<span class="term-rose">Error al consultar procesos del contrato: ${err.message}</span>`);
    }
  }

  async handleRunCommand(rawCmd, args) {
    if (args.length < 1) {
      this.appendEntry(
        rawCmd,
        `<span class="term-rose">Uso: run &lt;FCFS|SJF|RR|PRIORIDAD&gt; [archivo.json] [quantum]</span><br><span class="term-gray">Ejemplo: run FCFS data/procesos_semana4.json</span>`
      );
      return;
    }

    const algo = args[0].toUpperCase();
    const file = args[1] || 'data/procesos_semana4.json';
    const quantum = args[2] || '4';

    try {
      const res = await fetch(`/api/v1/cpu/schedule?algo=${encodeURIComponent(algo)}&file=${encodeURIComponent(file)}&quantum=${encodeURIComponent(quantum)}`);
      const data = await res.json();

      if (!res.ok) {
        this.appendEntry(rawCmd, `<span class="term-rose">Error C (${res.status}): ${this.escapeHtml(data.error || 'Fallo de simulacion')}</span><br><span class="term-gray">${this.escapeHtml(data.details || '')}</span>`);
        return;
      }

      let ganttStr = '';
      if (data.timeline && data.timeline.length > 0) {
        ganttStr = data.timeline.map(t => `[${t.start}]--${t.name}(${t.duration}m)--[${t.end}]`).join('');
      }

      let rows = '';
      if (data.processes && data.processes.length > 0) {
        data.processes.forEach(p => {
          rows += `<tr>
            <td>${p.name} (PID ${p.pid})</td>
            <td>${p.waiting_time} min</td>
            <td>${p.turnaround_time} min</td>
            <td>${p.response_time !== undefined ? p.response_time + ' min' : '-'}</td>
          </tr>`;
        });
      }

      const metricsHtml = `
<div class="term-green term-bold">[MOTOR C OK - Latencia: ${data.latency_ms || 0}ms] Planificación: ${data.algorithm}</div>
<div>Archivo: <span class="term-cyan">${this.escapeHtml(file)}</span> ${algo === 'RR' ? `| Quantum: <span class="term-amber">${quantum} min</span>` : ''}</div>
<div style="margin-top: 6px;">Traza de Ejecución (Gantt determinista):</div>
<div class="term-amber" style="font-family: monospace; padding: 4px 0; word-break: break-all;">${ganttStr}</div>
<table class="term-table">
  <thead>
    <tr><th>Proceso</th><th>Espera (W_i)</th><th>Retorno (T_i)</th><th>Respuesta (R_i)</th></tr>
  </thead>
  <tbody>
    ${rows}
  </tbody>
</table>
<div class="term-cyan term-bold">
  Espera Promedio: ${data.average_waiting_time} min | Retorno Promedio: ${data.average_turnaround_time} min | Respuesta Promedio: ${data.average_response_time} min
</div>
`;
      this.appendEntry(rawCmd, metricsHtml);

      // Sincronizar y abrir la ventana de Task Manager
      if (window.osManager) {
        window.osManager.openWindow('win-tasks');
      }
      if (window.taskManager) {
        const select = document.getElementById('select-sim-algo');
        if (select) select.value = algo;
        window.taskManager.currentAlgo = algo;
        await window.taskManager.fetchScheduleData();
        window.taskManager.reset();
      }
    } catch (err) {
      this.appendEntry(rawCmd, `<span class="term-rose">Fallo de conexion con API /api/v1/cpu/schedule: ${this.escapeHtml(err.message)}</span>`);
    }
  }

  handleKillCommand(rawCmd, args) {
    if (args.length === 0) {
      this.appendEntry(rawCmd, `<span class="term-rose">Uso: kill &lt;PID&gt;</span>`);
      return;
    }
    const pid = parseInt(args[0], 10);
    if (isNaN(pid)) {
      this.appendEntry(rawCmd, `<span class="term-rose">kill: argumento PID inválido: '${this.escapeHtml(args[0])}'</span>`);
      return;
    }
    if (window.taskManager) {
      const p = window.taskManager.processes.find(x => x.pid === pid);
      if (!p) {
        this.appendEntry(rawCmd, `<span class="term-rose">kill: no existe proceso con PID ${pid} en la simulación</span>`);
        return;
      }
      if (p.state === 'TERMINATED') {
        this.appendEntry(rawCmd, `<span class="term-amber">kill: el proceso ${p.name} (PID ${pid}) ya se encuentra TERMINATED.</span>`);
        return;
      }
      window.taskManager.killProcess(pid);
      this.appendEntry(rawCmd, `<span class="term-green">[SIGTERM] Proceso ${p.name} (PID ${pid}) terminado exitosamente. Bloque de memoria liberado y huecos contiguos fusionados.</span>`);
    } else {
      this.appendEntry(rawCmd, `<span class="term-rose">Task Manager no inicializado.</span>`);
    }
  }

  async handleMmuCommand(rawCmd, args) {
    const virtualAddr = args[0] || '10000';
    const pageSize = args[1] || '4096';
    try {
      const res = await fetch(`/api/v1/memory/translate?virtual=${encodeURIComponent(virtualAddr)}&page_size=${encodeURIComponent(pageSize)}`);
      const data = await res.json();
      if (!res.ok) {
        this.appendEntry(rawCmd, `<span class="term-rose">Fallo MMU (${res.status}): ${this.escapeHtml(data.error || '')}</span>`);
        return;
      }
      this.appendEntry(rawCmd, `
<div class="term-green term-bold">[MMU Paginación - Motor C OK]</div>
<div>Dirección Virtual: <span class="term-cyan">${data.virtual_address}</span> | Tamaño Página: <span class="term-amber">${data.page_size} B</span></div>
<div>Número de Página (p): <span class="term-cyan">${data.page_number}</span> &rarr; Marco Asignado (f): <span class="term-green">${data.frame_number}</span></div>
<div>Desplazamiento (d): <span class="term-amber">${data.offset}</span></div>
<div>Dirección Física (f &times; 4096 + d): <strong class="term-green" style="font-size:1.1em;">${data.physical_address}</strong></div>
      `);
    } catch (err) {
      this.appendEntry(rawCmd, `<span class="term-rose">Error al consultar MMU: ${this.escapeHtml(err.message)}</span>`);
    }
  }

  async handleSegCommand(rawCmd, args) {
    if (args.length < 2) {
      this.appendEntry(rawCmd, `<span class="term-rose">Uso: seg &lt;segment_id: 0=Codigo, 1=Datos, 2=Pila&gt; &lt;offset&gt;</span><br><span class="term-gray">Ejemplo válido: seg 1 300 &rarr; 5500 | Ejemplo Trap: seg 1 900 &rarr; Exception</span>`);
      return;
    }
    const segId = args[0];
    const offset = args[1];
    try {
      const res = await fetch(`/api/v1/memory/translate?mode=segmentation&segment=${encodeURIComponent(segId)}&offset=${encodeURIComponent(offset)}`);
      const data = await res.json();
      if (!res.ok) {
        this.appendEntry(rawCmd, `
<div class="term-rose term-bold">🛑 [TRAP / EXCEPCIÓN DETECTADA]</div>
<div class="term-rose">${this.escapeHtml(data.details || data.error || 'Violación de segmento')}</div>
<div class="term-gray">El desplazamiento ${offset} supera el límite del segmento ${segId}. Señal SIGSEGV simulada.</div>
        `);
        return;
      }
      this.appendEntry(rawCmd, `
<div class="term-green term-bold">[Segmentación MMU - Motor C OK]</div>
<div>Segmento: <span class="term-cyan">${data.segment_name} (ID: ${data.segment_id})</span> | Base: <span class="term-cyan">${data.base}</span> | Límite: <span class="term-amber">${data.limit}</span></div>
<div>Desplazamiento solicitado: <span class="term-amber">${data.offset}</span> (&lt; límite ${data.limit} &check;)</div>
<div>Dirección Física (Base + Offset): <strong class="term-green" style="font-size:1.1em;">${data.physical_address}</strong></div>
      `);
    } catch (err) {
      this.appendEntry(rawCmd, `<span class="term-rose">Error al consultar segmentación: ${this.escapeHtml(err.message)}</span>`);
    }
  }

  async handleReplaceCommand(rawCmd, args) {
    const frames = args[0] || '3';
    try {
      const res = await fetch(`/api/v1/memory/replace?frames=${encodeURIComponent(frames)}`);
      const data = await res.json();
      if (!res.ok) {
        this.appendEntry(rawCmd, `<span class="term-rose">Error en reemplazo de páginas: ${this.escapeHtml(data.error || '')}</span>`);
        return;
      }
      const b = data.belady_anomaly;
      this.appendEntry(rawCmd, `
<div class="term-green term-bold">[Reemplazo de Páginas - Algoritmos M3]</div>
<div>Patrón de Referencias: <span class="term-cyan">[${data.references.join(', ')}]</span> | Marcos: <span class="term-amber">${data.frames}</span></div>
<table class="term-table">
  <thead>
    <tr><th>Algoritmo</th><th>Fallos de Página</th><th>Aciertos</th></tr>
  </thead>
  <tbody>
    <tr><td>FIFO (${frames} marcos)</td><td><span class="term-rose">${data.fifo.faults}</span></td><td>${data.fifo.hits}</td></tr>
    <tr><td>LRU (${frames} marcos)</td><td><span class="term-amber">${data.lru.faults}</span></td><td>${data.lru.hits}</td></tr>
    <tr><td>Óptimo (${frames} marcos)</td><td><span class="term-green">${data.optimal.faults}</span></td><td>${data.optimal.hits}</td></tr>
  </tbody>
</table>
<div class="term-cyan" style="margin-top:4px;">
  <strong>Anomalía de Belády:</strong> FIFO (3 marcos) = <span class="term-green">${b.fifo_3_frames_faults} fallos</span> vs FIFO (4 marcos) = <span class="term-rose">${b.fifo_4_frames_faults} fallos</span>.
  <br><span class="term-gray">${b.note}</span>
</div>
      `);
    } catch (err) {
      this.appendEntry(rawCmd, `<span class="term-rose">Error al consultar reemplazo: ${this.escapeHtml(err.message)}</span>`);
    }
  }

  async handleSyncCommand(rawCmd, args) {
    const sub = (args[0] || 'status').toLowerCase();
    try {
      let url = '/api/v1/sync/status';
      if (sub === 'step') {
        url = '/api/v1/sync/step?actor=philosopher';
      } else if (sub === 'deadlock') {
        url = '/api/v1/sync/step?mode=deadlock';
      } else if (sub === 'preventive' || sub === 'preventiva') {
        url = '/api/v1/sync/step?mode=preventiva';
      } else if (sub === 'produce') {
        url = '/api/v1/sync/step?action=produce';
      } else if (sub === 'consume') {
        url = '/api/v1/sync/step?action=consume';
      }

      const res = await fetch(url);
      const data = await res.json();
      const ph = data.philosophers;
      const pc = data.prod_cons;

      this.appendEntry(rawCmd, `
<div class="term-green term-bold">[Sincronización Go - Concurrencia M2]</div>
<div><strong>Filósofos:</strong> Modo: <span class="term-cyan">${ph.mode}</span> | Deadlock RAG: ${ph.deadlock ? '<span class="term-rose term-bold">SÍ (Ciclo: [' + ph.cycle.join(', ') + '])</span>' : '<span class="term-green">NO</span>'}</div>
<div>Estados: [${ph.states.join(', ')}] | Platos: ${Array.isArray(ph.eat_count) ? ph.eat_count.join(', ') : ph.eat_count}</div>
<div style="margin-top:4px;"><strong>Buffer Productor-Consumidor:</strong> Ocupación: <span class="term-cyan">${pc.buffer_len}/${pc.capacity}</span> | Items: [${pc.items.join(', ')}]</div>
<div>Semáforos: mutex=<span class="term-green">${pc.semaphores.mutex}</span>, vacios=<span class="term-cyan">${pc.semaphores.vacios}</span>, llenos=<span class="term-amber">${pc.semaphores.llenos}</span></div>
      `);

      if (window.osManager) {
        window.osManager.openWindow('win-sync');
      }
      if (window.syncManager) {
        window.syncManager.philoMode = ph.mode;
        window.syncManager.philoStates = ph.states;
        window.syncManager.forksOwner = ph.forks_owner;
        window.syncManager.renderPhilosophers();

        window.syncManager.buffer = pc.items;
        window.syncManager.mutex = pc.semaphores.mutex;
        window.syncManager.empty = pc.semaphores.vacios;
        window.syncManager.full = pc.semaphores.llenos;
        window.syncManager.renderProdCons();
      }
    } catch (err) {
      this.appendEntry(rawCmd, `<span class="term-rose">Error sync: ${this.escapeHtml(err.message)}</span>`);
    }
  }
      window.osManager.refreshProcessesTable();
    }
  }

  async handleCatCommand(rawCmd, args) {
    if (args.length === 0) {
      this.appendEntry(rawCmd, `<span class="term-rose">cat: falta el nombre del archivo. Uso: cat &lt;archivo&gt;</span>`);
      return;
    }

    let filePath = args[0];
    if (filePath.startsWith('data/') || filePath.endsWith('.json')) {
      const cleanPath = filePath.startsWith('/') ? filePath : '/' + filePath;
      try {
        const res = await fetch(cleanPath);
        if (!res.ok) throw new Error('HTTP ' + res.status);
        const text = await res.text();
        this.appendEntry(rawCmd, `<pre class="term-cyan" style="white-space: pre-wrap; font-family: inherit;">${this.escapeHtml(text)}</pre>`);
        return;
      } catch (err) {
        this.appendEntry(rawCmd, `<span class="term-rose">cat: no se pudo leer '${this.escapeHtml(filePath)}': ${err.message}</span>`);
        return;
      }
    }

    this.appendEntry(
      rawCmd,
      `<span class="term-gray">[Aviso VFS]: Sintaxis 'cat ${this.escapeHtml(filePath)}' reconocida. La lectura de archivos virtuales creados en tiempo de ejecución se habilitará con el VFS en Sprint 3.</span>`
    );
  }

  handleOpenCommand(rawCmd, args) {
    if (args.length === 0) {
      this.appendEntry(rawCmd, `<span class="term-rose">Uso: open &lt;terminal|tasks|sync|about&gt;</span>`);
      return;
    }

    const target = args[0].toLowerCase();
    const map = {
      terminal: 'win-terminal',
      tasks: 'win-tasks',
      sync: 'win-sync',
      about: 'win-about',
    };

    const winId = map[target];
    if (winId && window.osManager) {
      window.osManager.openWindow(winId);
      this.appendEntry(rawCmd, `<span class="term-green">Ventana '${target}' abierta en el escritorio.</span>`);
    } else {
      this.appendEntry(rawCmd, `<span class="term-rose">Ventana desconocida: '${this.escapeHtml(target)}'. Opciones: terminal, tasks, sync, about</span>`);
    }
  }

  scrollToBottom() {
    if (this.screenEl) {
      this.screenEl.scrollTop = this.screenEl.scrollHeight;
    }
  }

  escapeHtml(str) {
    if (!str) return '';
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }
}

// Inicialización de la terminal al cargar el DOM
window.addEventListener('DOMContentLoaded', () => {
  window.terminalCLI = new TerminalCLI();
});
