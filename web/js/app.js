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

// Inicialización global
window.addEventListener('DOMContentLoaded', () => {
  window.osManager = new WindowManager();
});
