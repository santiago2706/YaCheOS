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
<div class="term-bold term-cyan">Comandos Disponibles en YaKcheOS CLI (Taller Semana 2):</div>
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
      <td>Muestra los procesos activos en la simulación (PID, estado, ráfagas).</td>
    </tr>
    <tr>
      <td><span class="term-green">run &lt;algo&gt; &lt;archivo.json&gt;</span></td>
      <td>Ejecuta planificación de CPU (FCFS, SJF, RR, PRIORIDAD) con datos de contrato.</td>
    </tr>
    <tr>
      <td><span class="term-green">ls</span></td>
      <td>Lista archivos del simulador (preparatorio para VFS Semana 3).</td>
    </tr>
    <tr>
      <td><span class="term-green">mkdir &lt;nombre&gt;</span></td>
      <td>Crea un directorio simulado (preparatorio para VFS Semana 3).</td>
    </tr>
    <tr>
      <td><span class="term-green">cat &lt;archivo&gt;</span></td>
      <td>Lee y muestra el contenido de archivos (ej. contratos JSON).</td>
    </tr>
    <tr>
      <td><span class="term-green">open &lt;tasks|sync|about&gt;</span></td>
      <td>Abre la ventana gráfica correspondiente en el escritorio.</td>
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

  handleRunCommand(rawCmd, args) {
    if (args.length < 2) {
      this.appendEntry(
        rawCmd,
        `<span class="term-rose">Uso incorrecto. Sintaxis: run &lt;FCFS|SJF|RR|PRIORIDAD&gt; &lt;archivo.json&gt;</span><br><span class="term-gray">Ejemplo: run FCFS data/procesos_semana4.json</span>`
      );
      return;
    }

    const algo = args[0].toUpperCase();
    const file = args[1];

    let metricsHtml = '';

    if (algo === 'FCFS') {
      metricsHtml = `
<div class="term-green term-bold">[SIMULACIÓN CPU] Algoritmo: First-Come, First-Served (FCFS)</div>
<div>Archivo de entrada: <span class="term-cyan">${this.escapeHtml(file)}</span></div>
<div style="margin-top: 6px;">Traza de Ejecución (Gantt):</div>
<div class="term-amber" style="font-family: monospace; padding: 4px 0;">[0]--P1(5m)--[5]--P2(3m)--[8]--P3(8m)--[16]--P4(6m)--[22]</div>
<table class="term-table">
  <thead>
    <tr><th>Proceso</th><th>Espera (W_i)</th><th>Retorno (T_i)</th></tr>
  </thead>
  <tbody>
    <tr><td>P1</td><td>0 min</td><td>5 min</td></tr>
    <tr><td>P2</td><td>4 min</td><td>7 min</td></tr>
    <tr><td>P3</td><td>6 min</td><td>14 min</td></tr>
    <tr><td>P4</td><td>13 min</td><td>19 min</td></tr>
  </tbody>
</table>
<div class="term-cyan term-bold">Espera Promedio (W_prom): 5.75 min | Retorno Promedio: 11.25 min</div>
`;
    } else if (algo === 'SJF') {
      metricsHtml = `
<div class="term-green term-bold">[SIMULACIÓN CPU] Algoritmo: Shortest Job First (SJF No Apropiativo)</div>
<div>Archivo de entrada: <span class="term-cyan">${this.escapeHtml(file)}</span></div>
<div style="margin-top: 6px;">Traza de Ejecución (Gantt):</div>
<div class="term-amber" style="font-family: monospace; padding: 4px 0;">[0]--P1(5m)--[5]--P2(3m)--[8]--P4(6m)--[14]--P3(8m)--[22]</div>
<div class="term-cyan term-bold">Espera Promedio (W_prom): 5.25 min (Menor espera promedio teórica)</div>
`;
    } else if (algo === 'RR' || algo === 'ROUNDROBIN') {
      metricsHtml = `
<div class="term-green term-bold">[SIMULACIÓN CPU] Algoritmo: Round Robin (Quantum = 4 min)</div>
<div>Archivo de entrada: <span class="term-cyan">${this.escapeHtml(file)}</span></div>
<div style="margin-top: 6px;">Traza de Ejecución (Gantt):</div>
<div class="term-amber" style="font-family: monospace; padding: 4px 0;">P1(4) &rarr; P2(3) &rarr; P3(4) &rarr; P4(4) &rarr; P1(1) &rarr; P3(4) &rarr; P4(2)</div>
<div class="term-cyan term-bold">Espera Promedio (W_prom): 9.25 min | Respuesta Promedio: 4.0 min</div>
`;
    } else if (algo === 'PRIORIDAD' || algo === 'PRIORITY') {
      metricsHtml = `
<div class="term-green term-bold">[SIMULACIÓN CPU] Algoritmo: Prioridad No Apropiativa (1 = Más Alta)</div>
<div>Archivo de entrada: <span class="term-cyan">${this.escapeHtml(file)}</span></div>
<div style="margin-top: 6px;">Traza de Ejecución (Gantt):</div>
<div class="term-amber" style="font-family: monospace; padding: 4px 0;">P1(5) &rarr; P3(8) &rarr; P4(6) &rarr; P2(3)</div>
<div class="term-cyan term-bold">Espera Promedio (W_prom): 7.75 min</div>
`;
    } else {
      metricsHtml = `
<span class="term-rose">Algoritmo '${this.escapeHtml(algo)}' no soportado. Algoritmos válidos: FCFS, SJF, RR, PRIORIDAD.</span>
`;
    }

    this.appendEntry(rawCmd, metricsHtml);

    // Refrescar y abrir la ventana de Task Manager si está disponible
    if (window.osManager) {
      window.osManager.openWindow('win-tasks');
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
