/**
 * YaKcheOS - Desktop Window Manager & OS Shell (Sprint 1 - T1.2)
 * Autor: Integrante 1 (Líder Técnico & Orquestador Web)
 * Gestiona ventanas flotantes, arrastre suave, minimizar, maximizar, cerrar y barra de tareas.
 */

class WindowManager {
  constructor() {
    this.windows = new Map();
    this.highestZIndex = 100;
    this.activeWindowId = null;
    this.draggedWindow = null;
    this.resizedWindow = null;
    this.dragOffset = { x: 0, y: 0 };
    this.resizeStart = { x: 0, y: 0, width: 0, height: 0 };

    this.init();
  }

  init() {
    this.bindGlobalEvents();
    this.setupDesktopIcons();
    this.setupStartMenu();
    this.startClock();
    this.registerExistingWindows();
    this.setupTaskManagerEvents();
  }

  /**
   * Registra las ventanas declaradas en el DOM
   */
  registerExistingWindows() {
    document.querySelectorAll('.os-window').forEach((winEl) => {
      const id = winEl.id;
      const titleSpan = winEl.querySelector('.window-title span');
      const title = titleSpan ? titleSpan.innerText : id;

      const winData = {
        id,
        element: winEl,
        title,
        isMinimized: false,
        isMaximized: false,
        prevBounds: {
          top: winEl.style.top || '60px',
          left: winEl.style.left || '60px',
          width: winEl.style.width || '650px',
          height: winEl.style.height || '420px',
        },
      };

      this.windows.set(id, winData);
      this.attachWindowControls(winData);
    });

    this.updateTaskbar();

    // Activar terminal por defecto
    if (this.windows.has('win-terminal')) {
      this.bringToFront('win-terminal');
    }
  }

  /**
   * Asocia controladores de arrastre, redimensión y botones a la ventana
   */
  attachWindowControls(win) {
    const el = win.element;

    // Foco al hacer clic
    el.addEventListener('pointerdown', () => {
      this.bringToFront(win.id);
    });

    // Barra de título y arrastre (drag)
    const titlebar = el.querySelector('.window-titlebar');
    if (titlebar) {
      titlebar.addEventListener('pointerdown', (e) => {
        if (e.target.closest('.win-btn')) return; // No arrastrar si se pulsó un botón

        this.bringToFront(win.id);

        if (win.isMaximized) {
          // Si está maximizada, restaurar al arrastrar
          this.maximizeWindow(win.id);
          const rect = el.getBoundingClientRect();
          el.style.left = `${Math.max(10, e.clientX - rect.width / 2)}px`;
          el.style.top = `${Math.max(10, e.clientY - 20)}px`;
        }

        const rect = el.getBoundingClientRect();
        this.draggedWindow = win;
        this.dragOffset = {
          x: e.clientX - rect.left,
          y: e.clientY - rect.top,
        };

        titlebar.setPointerCapture(e.pointerId);
      });

      titlebar.addEventListener('pointermove', (e) => {
        if (this.draggedWindow && this.draggedWindow.id === win.id) {
          const newX = e.clientX - this.dragOffset.x;
          const newY = e.clientY - this.dragOffset.y;

          // Mantener dentro de los bordes visibles
          const boundedX = Math.max(-200, Math.min(window.innerWidth - 80, newX));
          const boundedY = Math.max(0, Math.min(window.innerHeight - 80, newY));

          el.style.left = `${boundedX}px`;
          el.style.top = `${boundedY}px`;
        }
      });

      const endDrag = (e) => {
        if (this.draggedWindow && this.draggedWindow.id === win.id) {
          this.draggedWindow = null;
          try {
            titlebar.releasePointerCapture(e.pointerId);
          } catch (_) {}
        }
      };

      titlebar.addEventListener('pointerup', endDrag);
      titlebar.addEventListener('pointercancel', endDrag);

      // Doble clic para maximizar/restaurar
      titlebar.addEventListener('dblclick', (e) => {
        if (e.target.closest('.win-btn')) return;
        this.maximizeWindow(win.id);
      });
    }

    // Botones de control
    const btnClose = el.querySelector('.win-btn-close');
    const btnMin = el.querySelector('.win-btn-minimize');
    const btnMax = el.querySelector('.win-btn-maximize');

    if (btnClose) {
      btnClose.addEventListener('click', (e) => {
        e.stopPropagation();
        this.closeWindow(win.id);
      });
    }

    if (btnMin) {
      btnMin.addEventListener('click', (e) => {
        e.stopPropagation();
        this.minimizeWindow(win.id);
      });
    }

    if (btnMax) {
      btnMax.addEventListener('click', (e) => {
        e.stopPropagation();
        this.maximizeWindow(win.id);
      });
    }

    // Manija de redimensión (resize handle)
    const resizeHandle = el.querySelector('.window-resize-handle');
    if (resizeHandle) {
      resizeHandle.addEventListener('pointerdown', (e) => {
        e.stopPropagation();
        this.bringToFront(win.id);
        this.resizedWindow = win;
        const rect = el.getBoundingClientRect();
        this.resizeStart = {
          x: e.clientX,
          y: e.clientY,
          width: rect.width,
          height: rect.height,
        };
        resizeHandle.setPointerCapture(e.pointerId);
      });

      resizeHandle.addEventListener('pointermove', (e) => {
        if (this.resizedWindow && this.resizedWindow.id === win.id) {
          const deltaX = e.clientX - this.resizeStart.x;
          const deltaY = e.clientY - this.resizeStart.y;
          const newWidth = Math.max(340, this.resizeStart.width + deltaX);
          const newHeight = Math.max(220, this.resizeStart.height + deltaY);

          el.style.width = `${newWidth}px`;
          el.style.height = `${newHeight}px`;
        }
      });

      const endResize = (e) => {
        if (this.resizedWindow && this.resizedWindow.id === win.id) {
          this.resizedWindow = null;
          try {
            resizeHandle.releasePointerCapture(e.pointerId);
          } catch (_) {}
        }
      };

      resizeHandle.addEventListener('pointerup', endResize);
      resizeHandle.addEventListener('pointercancel', endResize);
    }
  }

  /**
   * Traer una ventana al frente (mayor z-index)
   */
  bringToFront(winId) {
    const win = this.windows.get(winId);
    if (!win) return;

    this.highestZIndex += 1;
    win.element.style.zIndex = this.highestZIndex;

    document.querySelectorAll('.os-window').forEach((w) => {
      w.classList.remove('active-window');
    });
    win.element.classList.add('active-window');

    if (win.isMinimized) {
      win.isMinimized = false;
      win.element.classList.remove('minimized');
    }

    this.activeWindowId = winId;
    this.updateTaskbar();
  }

  /**
   * Minimizar ventana a la barra de tareas
   */
  minimizeWindow(winId) {
    const win = this.windows.get(winId);
    if (!win) return;

    win.isMinimized = true;
    win.element.classList.add('minimized');
    win.element.classList.remove('active-window');

    if (this.activeWindowId === winId) {
      this.activeWindowId = null;
      this.focusNextWindow();
    }

    this.updateTaskbar();
  }

  /**
   * Maximizar / Restaurar ventana
   */
  maximizeWindow(winId) {
    const win = this.windows.get(winId);
    if (!win) return;

    const el = win.element;
    if (!win.isMaximized) {
      win.prevBounds = {
        top: el.style.top,
        left: el.style.left,
        width: el.style.width,
        height: el.style.height,
      };
      win.isMaximized = true;
      el.classList.add('maximized');
    } else {
      win.isMaximized = false;
      el.classList.remove('maximized');
      el.style.top = win.prevBounds.top;
      el.style.left = win.prevBounds.left;
      el.style.width = win.prevBounds.width;
      el.style.height = win.prevBounds.height;
    }

    this.bringToFront(winId);
  }

  /**
   * Cerrar ventana
   */
  closeWindow(winId) {
    const win = this.windows.get(winId);
    if (!win) return;

    win.element.style.display = 'none';
    win.isMinimized = true;
    win.element.classList.remove('active-window');

    if (this.activeWindowId === winId) {
      this.activeWindowId = null;
      this.focusNextWindow();
    }

    this.updateTaskbar();
  }

  /**
   * Abrir o revelar una ventana
   */
  openWindow(winId) {
    const win = this.windows.get(winId);
    if (!win) return;

    win.element.style.display = 'flex';
    win.isMinimized = false;
    win.element.classList.remove('minimized');
    this.bringToFront(winId);
  }

  /**
   * Alternar estado desde la barra de tareas
   */
  toggleWindow(winId) {
    const win = this.windows.get(winId);
    if (!win) return;

    if (win.element.style.display === 'none') {
      this.openWindow(winId);
    } else if (win.isMinimized) {
      this.bringToFront(winId);
    } else if (this.activeWindowId === winId) {
      this.minimizeWindow(winId);
    } else {
      this.bringToFront(winId);
    }
  }

  /**
   * Dar foco a la siguiente ventana visible
   */
  focusNextWindow() {
    let topWin = null;
    let highestZ = -1;

    this.windows.forEach((w) => {
      if (w.element.style.display !== 'none' && !w.isMinimized) {
        const z = parseInt(w.element.style.zIndex || '0', 10);
        if (z > highestZ) {
          highestZ = z;
          topWin = w;
        }
      }
    });

    if (topWin) {
      this.bringToFront(topWin.id);
    }
  }

  /**
   * Sincronizar botones de la barra de tareas con las ventanas abiertas
   */
  updateTaskbar() {
    const container = document.getElementById('taskbar-tasks');
    if (!container) return;

    container.innerHTML = '';

    this.windows.forEach((win) => {
      if (win.element.style.display !== 'none') {
        const item = document.createElement('div');
        const isActive = this.activeWindowId === win.id && !win.isMinimized;
        item.className = `task-item ${isActive ? 'active' : ''}`;
        item.innerHTML = `
          <div class="status-dot"></div>
          <span>${win.title}</span>
        `;
        item.addEventListener('click', () => {
          this.toggleWindow(win.id);
        });
        container.appendChild(item);
      }
    });
  }

  /**
   * Configuración de iconos de escritorio
   */
  setupDesktopIcons() {
    document.querySelectorAll('.desktop-icon').forEach((icon) => {
      const targetWinId = icon.getAttribute('data-window-target');
      if (!targetWinId) return;

      icon.addEventListener('click', () => {
        document.querySelectorAll('.desktop-icon').forEach((i) => i.classList.remove('selected'));
        icon.classList.add('selected');
        this.openWindow(targetWinId);
      });
    });

    // Deseleccionar iconos al pulsar el escritorio
    const desktop = document.getElementById('desktop');
    if (desktop) {
      desktop.addEventListener('pointerdown', (e) => {
        if (e.target === desktop) {
          document.querySelectorAll('.desktop-icon').forEach((i) => i.classList.remove('selected'));
        }
      });
    }
  }

  /**
   * Menú de inicio
   */
  setupStartMenu() {
    const startBtn = document.getElementById('start-button');
    const startMenu = document.getElementById('start-menu');

    if (startBtn && startMenu) {
      startBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        startMenu.classList.toggle('show');
      });

      document.addEventListener('click', (e) => {
        if (!startMenu.contains(e.target) && !startBtn.contains(e.target)) {
          startMenu.classList.remove('show');
        }
      });

      startMenu.querySelectorAll('.start-app-item').forEach((item) => {
        item.addEventListener('click', () => {
          const target = item.getAttribute('data-window-target');
          if (target) {
            this.openWindow(target);
            startMenu.classList.remove('show');
          }
        });
      });
    }
  }

  /**
   * Reloj digital del System Tray
   */
  startClock() {
    const timeEl = document.getElementById('tray-time');
    const dateEl = document.getElementById('tray-date');

    const update = () => {
      const now = new Date();
      if (timeEl) {
        timeEl.textContent = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
      }
      if (dateEl) {
        dateEl.textContent = now.toLocaleDateString([], { month: 'short', day: 'numeric', year: 'numeric' });
      }
    };

    update();
    setInterval(update, 1000);
  }

  /**
   * Carga dinámica de procesos desde data/procesos_semana4.json
   */
  setupTaskManagerEvents() {
    const btnReload = document.getElementById('btn-reload-tasks');
    if (btnReload) {
      btnReload.addEventListener('click', () => this.refreshProcessesTable());
    }
  }

  async refreshProcessesTable() {
    const tbody = document.getElementById('tbody-processes');
    if (!tbody) return;

    try {
      const res = await fetch('/data/procesos_semana4.json');
      if (!res.ok) throw new Error('HTTP ' + res.status);
      const data = await res.json();
      const list = Array.isArray(data) ? data : (data.procesos || []);

      tbody.innerHTML = '';
      list.forEach((p) => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
          <td>${p.pid}</td>
          <td><strong style="color: #38bdf8;">${p.name || p.nombre}</strong></td>
          <td>${p.arrival_time !== undefined ? p.arrival_time : p.llegada} min</td>
          <td>${p.burst_time !== undefined ? p.burst_time : p.rafaga} min</td>
          <td>${p.priority !== undefined ? p.priority : p.prioridad}</td>
          <td>${p.memory_kb ? p.memory_kb + ' KB' : '128 KB'}</td>
          <td><span class="pm-badge" style="background: rgba(16,185,129,0.2); color: #34d399;">READY</span></td>
        `;
        tbody.appendChild(tr);
      });
    } catch (err) {
      console.warn('No se pudo recargar procesos desde API:', err);
    }
  }

  bindGlobalEvents() {
    window.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        const startMenu = document.getElementById('start-menu');
        if (startMenu) startMenu.classList.remove('show');
      }
    });
  }
}

/**
 * TaskManager (Sprint 2 - T1.5)
 * Gestiona simulación paso a paso por ticks, tabla en vivo, ordenación,
 * botón kill para liberar memoria, y mapa gráfico de bloques de RAM.
 */
class TaskManager {
  constructor() {
    this.processes = [];
    this.currentAlgo = 'FCFS';
    this.currentTick = 0;
    this.isRunning = false;
    this.timer = null;
    this.speed = 1000;
    this.sortCol = 'pid';
    this.sortAsc = true;
    this.timeline = [];
    this.totalRamKb = 1024;
    this.memoryBlocks = [];

    this.init();
  }

  async init() {
    this.bindEvents();
    await this.fetchScheduleData();
    this.initMemoryBlocks();
    this.render();
  }

  bindEvents() {
    const btnPlay = document.getElementById('btn-sim-play');
    const btnStep = document.getElementById('btn-sim-step');
    const btnReset = document.getElementById('btn-sim-reset');
    const selectAlgo = document.getElementById('select-sim-algo');

    if (btnPlay) {
      btnPlay.addEventListener('click', () => this.togglePlay());
    }
    if (btnStep) {
      btnStep.addEventListener('click', () => this.step());
    }
    if (btnReset) {
      btnReset.addEventListener('click', () => this.reset());
    }
    if (selectAlgo) {
      selectAlgo.addEventListener('change', async (e) => {
        this.currentAlgo = e.target.value;
        await this.fetchScheduleData();
        this.reset();
      });
    }

    // Cabeceras ordenables
    document.querySelectorAll('#table-processes th.sortable-th').forEach((th) => {
      th.addEventListener('click', () => {
        const col = th.getAttribute('data-sort');
        if (this.sortCol === col) {
          this.sortAsc = !this.sortAsc;
        } else {
          this.sortCol = col;
          this.sortAsc = true;
        }
        this.sortProcesses();
        this.renderTable();
      });
    });
  }

  async fetchScheduleData() {
    try {
      const res = await fetch(`/api/v1/cpu/schedule?algo=${this.currentAlgo}`);
      if (!res.ok) throw new Error('HTTP ' + res.status);
      const data = await res.json();
      this.timeline = data.timeline || [];

      // Actualizar métricas del dashboard
      const wEl = document.getElementById('sim-w-prom');
      if (wEl) wEl.textContent = `${data.average_waiting_time}m`;

      const gTxt = document.getElementById('gantt-summary-txt');
      if (gTxt) {
        gTxt.textContent = `${this.currentAlgo}: W_prom ${data.average_waiting_time} min | T_prom ${data.average_turnaround_time} min`;
      }

      this.renderGantt(this.timeline);
    } catch (err) {
      console.warn('Fallo al invocar /api/v1/cpu/schedule:', err);
    }
  }

  initMemoryBlocks() {
    // Procesos base de Semana 4
    this.processes = [
      { pid: 1, name: 'P1', arrival: 0, burst: 5, remaining: 5, priority: 3, memory: 128, state: 'READY' },
      { pid: 2, name: 'P2', arrival: 1, burst: 3, remaining: 3, priority: 4, memory: 64,  state: 'READY' },
      { pid: 3, name: 'P3', arrival: 2, burst: 8, remaining: 8, priority: 1, memory: 256, state: 'READY' },
      { pid: 4, name: 'P4', arrival: 3, burst: 6, remaining: 6, priority: 2, memory: 128, state: 'READY' },
    ];

    // Asignación contigua inicial en RAM (1024 KB total)
    let curAddr = 0;
    this.memoryBlocks = [];
    const colors = { 1: '#3b82f6', 2: '#10b981', 3: '#f59e0b', 4: '#8b5cf6' };

    this.processes.forEach((p) => {
      this.memoryBlocks.push({
        start: curAddr,
        size: p.memory,
        pid: p.pid,
        name: p.name,
        color: colors[p.pid] || '#06b6d4',
        free: false,
      });
      curAddr += p.memory;
    });

    // Hueco libre final
    if (curAddr < this.totalRamKb) {
      this.memoryBlocks.push({
        start: curAddr,
        size: this.totalRamKb - curAddr,
        pid: null,
        name: 'LIBRE',
        color: null,
        free: true,
      });
    }
  }

  togglePlay() {
    this.isRunning = !this.isRunning;
    const btn = document.getElementById('btn-sim-play');
    if (this.isRunning) {
      if (btn) btn.textContent = '⏸ Pausar';
      this.timer = setInterval(() => this.step(), this.speed);
    } else {
      if (btn) btn.textContent = '▶ Iniciar';
      clearInterval(this.timer);
    }
  }

  step() {
    this.currentTick++;

    // Determinar qué proceso está en ejecución en el tick actual según la línea de tiempo
    let activeSegment = this.timeline.find((seg) => this.currentTick > seg.start && this.currentTick <= seg.end);
    let activePid = activeSegment ? activeSegment.pid : null;

    // Actualizar procesos
    let allFinished = true;
    this.processes.forEach((p) => {
      if (p.state !== 'TERMINATED') {
        if (p.arrival > this.currentTick) {
          p.state = 'WAITING';
          allFinished = false;
        } else if (p.pid === activePid) {
          p.state = 'RUNNING';
          if (p.remaining > 0) p.remaining--;
          if (p.remaining === 0) {
            p.state = 'TERMINATED';
          } else {
            allFinished = false;
          }
        } else {
          p.state = 'READY';
          allFinished = false;
        }
      }
    });

    const tickEl = document.getElementById('sim-tick-count');
    if (tickEl) tickEl.textContent = this.currentTick;

    const cpuEl = document.getElementById('sim-cpu-status');
    if (cpuEl) {
      if (activeSegment) {
        cpuEl.textContent = `${activeSegment.name} [PID ${activeSegment.pid}]`;
      } else if (allFinished) {
        cpuEl.textContent = 'COMPLETADO';
        if (this.isRunning) this.togglePlay();
      } else {
        cpuEl.textContent = 'IDLE';
      }
    }

    this.render();
  }

  reset() {
    if (this.isRunning) this.togglePlay();
    this.currentTick = 0;
    const tickEl = document.getElementById('sim-tick-count');
    if (tickEl) tickEl.textContent = '0';
    const cpuEl = document.getElementById('sim-cpu-status');
    if (cpuEl) cpuEl.textContent = 'READY';

    this.initMemoryBlocks();
    this.render();
  }

  /**
   * Termina un proceso (kill <PID>) y libera su memoria asociada en RAM (T1.5)
   */
  killProcess(pid) {
    const p = this.processes.find((x) => x.pid === pid);
    if (!p) return;

    p.state = 'TERMINATED';
    p.remaining = 0;

    // Liberar memoria en el mapa de bloques y fusionar huecos contiguos
    this.memoryBlocks.forEach((b) => {
      if (b.pid === pid) {
        b.free = true;
        b.pid = null;
        b.name = 'LIBRE';
        b.color = null;
      }
    });

    this.mergeAdjacentHoles();
    this.render();

    console.log(`[Task Manager] Proceso ${p.name} (PID ${pid}) terminado. Memoria liberada.`);
  }

  mergeAdjacentHoles() {
    const merged = [];
    for (let i = 0; i < this.memoryBlocks.length; i++) {
      const b = this.memoryBlocks[i];
      if (merged.length > 0 && merged[merged.length - 1].free && b.free) {
        merged[merged.length - 1].size += b.size;
      } else {
        merged.push({ ...b });
      }
    }
    this.memoryBlocks = merged;
  }

  sortProcesses() {
    this.processes.sort((a, b) => {
      let valA = a[this.sortCol];
      let valB = b[this.sortCol];
      if (typeof valA === 'string') {
        return this.sortAsc ? valA.localeCompare(valB) : valB.localeCompare(valA);
      }
      return this.sortAsc ? valA - valB : valB - valA;
    });
  }

  render() {
    this.renderTable();
    this.renderMemoryMap();
  }

  renderTable() {
    const tbody = document.getElementById('tbody-processes');
    if (!tbody) return;

    tbody.innerHTML = '';
    this.processes.forEach((p) => {
      const tr = document.createElement('tr');
      let badgeStyle = 'background: rgba(16,185,129,0.2); color: #34d399;';
      if (p.state === 'RUNNING') badgeStyle = 'background: rgba(56,189,248,0.25); color: #38bdf8; font-weight:700;';
      else if (p.state === 'TERMINATED') badgeStyle = 'background: rgba(148,163,184,0.2); color: #94a3b8;';
      else if (p.state === 'WAITING') badgeStyle = 'background: rgba(245,158,11,0.2); color: #f59e0b;';

      tr.innerHTML = `
        <td>${p.pid}</td>
        <td><strong style="color: #38bdf8;">${p.name}</strong></td>
        <td>${p.arrival} min</td>
        <td>${p.burst} min</td>
        <td><strong style="color: #f1f5f9;">${p.remaining} min</strong></td>
        <td>${p.priority}</td>
        <td>${p.memory} KB</td>
        <td><span class="pm-badge" style="${badgeStyle}">${p.state}</span></td>
        <td style="text-align: center;">
          ${p.state !== 'TERMINATED' ? `<button class="btn-kill-proc" onclick="window.taskManager.killProcess(${p.pid})">Kill</button>` : `<span style="font-size:0.7rem; color:#64748b;">Inactivo</span>`}
        </td>
      `;
      tbody.appendChild(tr);
    });
  }

  renderMemoryMap() {
    const bar = document.getElementById('memory-map-bar');
    if (!bar) return;

    bar.innerHTML = '';
    let usedKb = 0;
    let freeKb = 0;

    this.memoryBlocks.forEach((b) => {
      const pct = (b.size / this.totalRamKb) * 100;
      const el = document.createElement('div');
      el.className = `mem-block ${b.free ? 'free-hole' : ''}`;
      el.style.width = `${pct}%`;
      if (!b.free) {
        el.style.background = b.color;
        usedKb += b.size;
      } else {
        freeKb += b.size;
      }
      el.title = `${b.name}: ${b.size} KB (Dirección: 0x${b.start.toString(16)})`;
      el.innerHTML = `<span>${b.name}</span><span style="font-size:0.6rem; opacity:0.8;">${b.size}K</span>`;
      bar.appendChild(el);
    });

    const usedTxt = document.getElementById('mem-used-txt');
    const freeTxt = document.getElementById('mem-free-txt');
    if (usedTxt) usedTxt.textContent = `${usedKb} KB`;
    if (freeTxt) freeTxt.textContent = `${freeKb} KB`;
  }

  renderGantt(timeline) {
    const bar = document.getElementById('gantt-bar');
    if (!bar || !timeline || timeline.length === 0) return;

    bar.innerHTML = '';
    const colors = { 1: '#3b82f6', 2: '#10b981', 3: '#f59e0b', 4: '#8b5cf6' };

    timeline.forEach((seg) => {
      const el = document.createElement('div');
      el.className = 'gantt-segment';
      el.style.flex = seg.duration;
      el.style.background = colors[seg.pid] || '#06b6d4';
      el.title = `${seg.name}: [${seg.start}-${seg.end}]`;
      el.textContent = `${seg.name} (${seg.duration}m)`;
      bar.appendChild(el);
    });
  }
}

/**
 * SyncManager (Sprint 2 - T1.6)
 * Gestiona la simulación de los 5 Filósofos Comensales (Deadlock vs Preventiva),
 * el buffer Productor-Consumidor y llamadas MMU.
 */
class SyncManager {
  constructor() {
    this.philoMode = 'deadlock';
    this.philoStates = ['PENSANDO', 'PENSANDO', 'PENSANDO', 'PENSANDO', 'PENSANDO'];
    this.forksOwner = [-1, -1, -1, -1, -1];
    this.eatCount = 0;
    this.isPhiloRunning = false;
    this.philoTimer = null;

    // Buffer Productor - Consumidor
    this.buffer = [];
    this.capacity = 5;
    this.mutex = 1;
    this.empty = 5;
    this.full = 0;
    this.isProdConsRunning = false;
    this.prodConsTimer = null;

    this.init();
  }

  init() {
    this.setupTabs();
    this.setupPhilosophersUI();
    this.setupProdConsUI();
    this.setupMMUUI();
    this.renderPhilosophers();
    this.renderProdCons();
  }

  setupTabs() {
    document.querySelectorAll('.window-tab').forEach((tab) => {
      tab.addEventListener('click', () => {
        document.querySelectorAll('.window-tab').forEach((t) => t.classList.remove('active'));
        document.querySelectorAll('.tab-pane').forEach((p) => (p.style.display = 'none'));

        tab.classList.add('active');
        const targetId = tab.getAttribute('data-tab-target');
        const targetPane = document.getElementById(targetId);
        if (targetPane) targetPane.style.display = 'block';
      });
    });
  }

  setupPhilosophersUI() {
    const selectMode = document.getElementById('select-philo-mode');
    const btnPlay = document.getElementById('btn-philo-play');
    const btnStep = document.getElementById('btn-philo-step');
    const btnReset = document.getElementById('btn-philo-reset');

    if (selectMode) {
      selectMode.addEventListener('change', (e) => {
        this.philoMode = e.target.value;
        const badge = document.getElementById('philo-mode-badge');
        if (badge) {
          if (this.philoMode === 'deadlock') {
            badge.textContent = 'Espera Circular';
            badge.style.color = '#f87171';
          } else {
            badge.textContent = 'Orden Asimétrico';
            badge.style.color = '#34d399';
          }
        }
        this.resetPhilosophers();
      });
    }

    if (btnPlay) {
      btnPlay.addEventListener('click', () => this.togglePhilosophersPlay());
    }
    if (btnStep) {
      btnStep.addEventListener('click', () => this.stepPhilosophers());
    }
    if (btnReset) {
      btnReset.addEventListener('click', () => this.resetPhilosophers());
    }
  }

  togglePhilosophersPlay() {
    this.isPhiloRunning = !this.isPhiloRunning;
    const btn = document.getElementById('btn-philo-play');
    if (this.isPhiloRunning) {
      if (btn) btn.textContent = '⏸ Pausar';
      this.philoTimer = setInterval(() => this.stepPhilosophers(), 900);
    } else {
      if (btn) btn.textContent = '▶ Iniciar';
      clearInterval(this.philoTimer);
    }
  }

  async stepPhilosophers() {
    try {
      const res = await fetch(`/api/v1/sync/step?actor=philosopher&mode=${encodeURIComponent(this.philoMode)}`);
      if (res.ok) {
        const data = await res.json();
        const ph = data.philosophers;
        this.philoStates = ph.states;
        this.forksOwner = ph.forks_owner;
        this.eatCount = Array.isArray(ph.eat_count) ? ph.eat_count.reduce((a, b) => a + b, 0) : ph.eat_count;

        const banner = document.getElementById('deadlock-alert-banner');
        if (banner) {
          if (ph.deadlock) {
            banner.classList.add('show');
            if (this.isPhiloRunning) this.togglePhilosophersPlay();
          } else {
            banner.classList.remove('show');
          }
        }
        this.renderPhilosophers();
        return;
      }
    } catch (_) {}

    // Fallback de cálculo local
    if (this.philoMode === 'deadlock') {
      let allTookLeft = true;
      for (let i = 0; i < 5; i++) {
        if (this.philoStates[i] === 'PENSANDO') {
          this.philoStates[i] = 'CON_HAMBRE';
          allTookLeft = false;
        } else if (this.philoStates[i] === 'CON_HAMBRE') {
          this.forksOwner[i] = i;
          this.philoStates[i] = 'INTERBLOQUEADO';
        }
      }

      const deadlockedCount = this.philoStates.filter((s) => s === 'INTERBLOQUEADO').length;
      if (deadlockedCount === 5) {
        const banner = document.getElementById('deadlock-alert-banner');
        if (banner) banner.classList.add('show');
        if (this.isPhiloRunning) this.togglePhilosophersPlay();
      }
    } else {
      const banner = document.getElementById('deadlock-alert-banner');
      if (banner) banner.classList.remove('show');

      const nextToEat = Math.floor(Math.random() * 5);
      for (let i = 0; i < 5; i++) {
        const left = i;
        const right = (i + 1) % 5;

        if (this.philoStates[i] === 'COMIENDO') {
          this.philoStates[i] = 'PENSANDO';
          if (this.forksOwner[left] === i) this.forksOwner[left] = -1;
          if (this.forksOwner[right] === i) this.forksOwner[right] = -1;
        } else if (i === nextToEat && this.forksOwner[left] === -1 && this.forksOwner[right] === -1) {
          this.forksOwner[left] = i;
          this.forksOwner[right] = i;
          this.philoStates[i] = 'COMIENDO';
          this.eatCount++;
        } else if (this.philoStates[i] === 'PENSANDO') {
          this.philoStates[i] = 'CON_HAMBRE';
        }
      }
    }

    this.renderPhilosophers();
  }

  async resetPhilosophers() {
    if (this.isPhiloRunning) this.togglePhilosophersPlay();
    try {
      await fetch('/api/v1/sync/step?action=reset');
    } catch (_) {}

    this.philoStates = ['PENSANDO', 'PENSANDO', 'PENSANDO', 'PENSANDO', 'PENSANDO'];
    this.forksOwner = [-1, -1, -1, -1, -1];
    this.eatCount = 0;
    const banner = document.getElementById('deadlock-alert-banner');
    if (banner) banner.classList.remove('show');
    this.renderPhilosophers();
  }

  renderPhilosophers() {
    const arena = document.getElementById('philosophers-arena');
    if (!arena) return;

    // Remover nodos previos excepto el plato central
    arena.querySelectorAll('.philo-node, .fork-node').forEach((el) => el.remove());

    const centerX = 360;
    const centerY = 125;
    const radiusPhilo = 88;
    const radiusFork = 54;

    const stateMap = {
      PENSANDO: { icon: '🧠', cls: 'philo-thinking', txt: 'Pensando' },
      CON_HAMBRE: { icon: '⏳', cls: 'philo-hungry', txt: 'Con Hambre' },
      COMIENDO: { icon: '🍝', cls: 'philo-eating', txt: 'Comiendo' },
      INTERBLOQUEADO: { icon: '🛑', cls: 'philo-deadlocked', txt: 'Deadlock' },
    };

    // Renderizar 5 filósofos y 5 tenedores en disposición circular
    for (let i = 0; i < 5; i++) {
      const angle = (i * 72 - 90) * (Math.PI / 180);
      const x = centerX + radiusPhilo * Math.cos(angle);
      const y = centerY + radiusPhilo * Math.sin(angle);

      const st = stateMap[this.philoStates[i]] || stateMap.PENSANDO;

      const pNode = document.createElement('div');
      pNode.className = `philo-node ${st.cls}`;
      pNode.style.left = `${x}px`;
      pNode.style.top = `${y}px`;
      pNode.innerHTML = `
        <span style="font-size:1.1rem;">${st.icon}</span>
        <span class="philo-name">F${i + 1}</span>
        <span class="philo-state">${st.txt}</span>
      `;
      arena.appendChild(pNode);

      // Tenedor entre i y (i+1)%5
      const forkAngle = (i * 72 - 90 + 36) * (Math.PI / 180);
      const fx = centerX + radiusFork * Math.cos(forkAngle);
      const fy = centerY + radiusFork * Math.sin(forkAngle);

      const isTaken = this.forksOwner[i] !== -1;
      const fNode = document.createElement('div');
      fNode.className = `fork-node ${isTaken ? 'fork-taken' : 'fork-free'}`;
      fNode.style.left = `${fx}px`;
      fNode.style.top = `${fy}px`;
      fNode.title = `Tenedor T${i} ${isTaken ? '(Retenido por F' + (this.forksOwner[i] + 1) + ')' : '(Libre)'}`;
      fNode.textContent = `T${i}`;
      arena.appendChild(fNode);
    }

    const stats = document.getElementById('philo-stats-txt');
    if (stats) stats.textContent = `Platos consumidos: ${this.eatCount}`;
  }

  setupProdConsUI() {
    const btnProd = document.getElementById('btn-prod-item');
    const btnCons = document.getElementById('btn-cons-item');
    const btnAuto = document.getElementById('btn-prodcons-auto');

    if (btnProd) btnProd.addEventListener('click', () => this.produceItem());
    if (btnCons) btnCons.addEventListener('click', () => this.consumeItem());
    if (btnAuto) btnAuto.addEventListener('click', () => this.toggleProdConsAuto());
  }

  async produceItem() {
    try {
      const res = await fetch('/api/v1/sync/step?action=produce');
      if (res.ok) {
        const data = await res.json();
        const pc = data.prod_cons;
        this.buffer = pc.items;
        this.mutex = pc.semaphores.mutex;
        this.empty = pc.semaphores.vacios;
        this.full = pc.semaphores.llenos;
        if (pc.events && pc.events.length > 0) {
          const last = pc.events[pc.events.length - 1];
          this.logSyncEvent(`[Go Backend] ${last.actor} &rarr; ${last.action}: ${last.details}`);
        }
        this.renderProdCons();
        return;
      }
    } catch (_) {}

    // Fallback de cálculo local
    if (this.empty > 0 && this.mutex === 1) {
      this.empty--;
      const item = Math.floor(Math.random() * 900) + 100;
      this.buffer.push(item);
      this.full++;
      this.logSyncEvent(`Productor produjo elemento #${item}. Ocupación: ${this.buffer.length}/${this.capacity}`);
    } else {
      this.logSyncEvent(`⚠️ Productor BLOQUEADO: Buffer lleno (vacios=0)`);
    }
    this.renderProdCons();
  }

  async consumeItem() {
    try {
      const res = await fetch('/api/v1/sync/step?action=consume');
      if (res.ok) {
        const data = await res.json();
        const pc = data.prod_cons;
        this.buffer = pc.items;
        this.mutex = pc.semaphores.mutex;
        this.empty = pc.semaphores.vacios;
        this.full = pc.semaphores.llenos;
        if (pc.events && pc.events.length > 0) {
          const last = pc.events[pc.events.length - 1];
          this.logSyncEvent(`[Go Backend] ${last.actor} &rarr; ${last.action}: ${last.details}`);
        }
        this.renderProdCons();
        return;
      }
    } catch (_) {}

    // Fallback de cálculo local
    if (this.full > 0 && this.mutex === 1) {
      this.full--;
      const item = this.buffer.shift();
      this.empty++;
      this.logSyncEvent(`Consumidor retiró elemento #${item}. Ocupación: ${this.buffer.length}/${this.capacity}`);
    } else {
      this.logSyncEvent(`⚠️ Consumidor BLOQUEADO: Buffer vacío (llenos=0)`);
    }
    this.renderProdCons();
  }

  toggleProdConsAuto() {
    this.isProdConsRunning = !this.isProdConsRunning;
    const btn = document.getElementById('btn-prodcons-auto');
    if (this.isProdConsRunning) {
      if (btn) btn.textContent = '⏸ Pausar';
      this.prodConsTimer = setInterval(() => {
        if (Math.random() > 0.45) this.produceItem();
        else this.consumeItem();
      }, 700);
    } else {
      if (btn) btn.textContent = '▶ Auto-Run';
      clearInterval(this.prodConsTimer);
    }
  }

  logSyncEvent(msg) {
    const box = document.getElementById('sync-log-box');
    if (!box) return;
    const time = new Date().toLocaleTimeString();
    const line = document.createElement('div');
    line.className = 'sync-log-line';
    line.innerHTML = `<span style="color:#64748b;">[${time}]</span> ${msg}`;
    box.appendChild(line);
    box.scrollTop = box.scrollHeight;
  }

  renderProdCons() {
    const semMutex = document.getElementById('sem-val-mutex');
    const semEmpty = document.getElementById('sem-val-empty');
    const semFull = document.getElementById('sem-val-full');

    if (semMutex) semMutex.textContent = this.mutex;
    if (semEmpty) semEmpty.textContent = this.empty;
    if (semFull) semFull.textContent = this.full;

    const grid = document.getElementById('prodcons-slots-grid');
    if (!grid) return;

    grid.innerHTML = '';
    for (let i = 0; i < this.capacity; i++) {
      const isOccupied = i < this.buffer.length;
      const val = isOccupied ? `#${this.buffer[i]}` : 'VACÍO';
      const slot = document.createElement('div');
      slot.style.background = isOccupied ? 'rgba(56, 189, 248, 0.2)' : 'rgba(30, 41, 59, 0.7)';
      slot.style.border = isOccupied ? '1px solid #38bdf8' : '1px dashed rgba(255,255,255,0.15)';
      slot.style.borderRadius = '8px';
      slot.style.padding = '10px 4px';
      slot.style.textAlign = 'center';

      slot.innerHTML = `
        <span style="font-size:0.68rem; color:#94a3b8;">Slot ${i}</span>
        <div style="font-size:0.8rem; font-weight:700; color:${isOccupied ? '#38bdf8' : '#64748b'}; margin-top:2px;">${val}</div>
      `;
      grid.appendChild(slot);
    }
  }

  setupMMUUI() {
    const btnMMU = document.getElementById('btn-calc-mmu');
    const btnReplace = document.getElementById('btn-calc-replace');

    if (btnMMU) {
      btnMMU.addEventListener('click', async () => {
        const input = document.getElementById('input-mmu-virtual');
        const v = input ? input.value : 10000;
        try {
          const res = await fetch(`/api/v1/memory/translate?virtual=${v}&page_size=4096`);
          const data = await res.json();
          const box = document.getElementById('mmu-result-box');
          if (box) {
            box.innerHTML = `
              <div>Página: <span style="color:#38bdf8;">${data.page_number}</span> (Marco: <span style="color:#34d399;">${data.frame_number}</span>)</div>
              <div>Desplazamiento: <span style="color:#fbbf24;">${data.offset}</span></div>
              <div>Dirección Física: <strong style="color:#34d399; font-size: 0.85rem;">${data.physical_address}</strong></div>
            `;
          }
        } catch (err) {
          console.warn('Fallo MMU:', err);
        }
      });
    }

    if (btnReplace) {
      btnReplace.addEventListener('click', async () => {
        try {
          const res = await fetch('/api/v1/memory/replace');
          const data = await res.json();
          const box = document.getElementById('replace-result-box');
          if (box && data.belady_anomaly) {
            box.innerHTML = `
              <div>FIFO (3 marcos): <strong style="color:#f87171;">${data.belady_anomaly.fifo_3_frames_faults} fallos</strong></div>
              <div>FIFO (4 marcos): <strong style="color:#ef4444;">${data.belady_anomaly.fifo_4_frames_faults} fallos</strong> (¡Belady Confirmado!)</div>
              <div>Óptimo (3 marcos): <strong style="color:#34d399;">${data.optimal.faults} fallos</strong></div>
            `;
          }
        } catch (err) {
          console.warn('Fallo replace:', err);
        }
      });
    }

    // Handlers para Segmentación y Detección de Trap (T2.6 / T1.4)
    const btnSeg = document.getElementById('btn-calc-seg');
    const btnSegValid = document.getElementById('btn-test-seg-valid');
    const btnSegTrap = document.getElementById('btn-test-seg-trap');

    const runSegmentation = async (segId, offset) => {
      const box = document.getElementById('seg-result-box');
      try {
        const res = await fetch(`/api/v1/memory/translate?mode=segmentation&segment=${segId}&offset=${offset}`);
        const data = await res.json();
        if (!res.ok) {
          if (box) {
            box.innerHTML = `
              <div style="color:#f87171; font-weight:700;">🛑 TRAP / EXCEPCIÓN DETECTADA</div>
              <div style="color:#fca5a5; font-size:0.68rem;">${data.details || data.error}</div>
              <div style="color:#94a3b8; font-size:0.65rem;">Offset ${offset} &gt; Límite. Señal SIGSEGV simulada.</div>
            `;
          }
          return;
        }
        if (box) {
          box.innerHTML = `
            <div>Segmento: <span style="color:#38bdf8;">${data.segment_name} (ID: ${data.segment_id})</span></div>
            <div>Base: ${data.base} | Límite: ${data.limit}</div>
            <div>Dirección Física: <strong style="color:#34d399; font-size: 0.85rem;">${data.physical_address}</strong> (Válido)</div>
          `;
        }
      } catch (err) {
        if (box) box.innerHTML = `<div style="color:#f87171;">Error al consultar API: ${err.message}</div>`;
      }
    };

    if (btnSeg) {
      btnSeg.addEventListener('click', () => {
        const segId = document.getElementById('select-seg-id')?.value || 1;
        const offset = document.getElementById('input-seg-offset')?.value || 300;
        runSegmentation(segId, offset);
      });
    }

    if (btnSegValid) {
      btnSegValid.addEventListener('click', () => {
        const sel = document.getElementById('select-seg-id');
        const inp = document.getElementById('input-seg-offset');
        if (sel) sel.value = '1';
        if (inp) inp.value = '300';
        runSegmentation(1, 300);
      });
    }

    if (btnSegTrap) {
      btnSegTrap.addEventListener('click', () => {
        const sel = document.getElementById('select-seg-id');
        const inp = document.getElementById('input-seg-offset');
        if (sel) sel.value = '1';
        if (inp) inp.value = '900';
        runSegmentation(1, 900);
      });
    }
  }
}

// Inicialización global al cargar el DOM
window.addEventListener('DOMContentLoaded', () => {
  window.osManager = new WindowManager();
  window.taskManager = new TaskManager();
  window.syncManager = new SyncManager();
});

