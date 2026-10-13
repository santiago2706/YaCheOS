# Informe técnico Beta — planificación y memoria

**Estado de la evidencia:** cálculos y gráficos de QA reproducibles. La comparación con el motor C y la visualización integrada en la web quedan pendientes de los módulos de los integrantes 1 y 2. Este documento no presenta resultados de esos motores como si ya se hubieran ejecutado.

## 1. Alcance, datos y método

Se usaron únicamente los casos numéricos del `plan_de_sprints_yakcheos_6_modulos.md`, secciones 2.1, 2.3 y tarea S3.1. Los datos de CPU están transcritos en [`tests_python/fixtures/procesos_semana4.json`](../tests_python/fixtures/procesos_semana4.json); los de paginación y segmentación, en [`tests_python/fixtures/memoria_semana6.json`](../tests_python/fixtures/memoria_semana6.json). El patrón de reemplazo es `1,2,3,4,1,2,5,1,2,3,4,5`, con 3 y 4 marcos. Los programas de Python son oráculos independientes; todavía no prueban que el código C produzca los mismos resultados.

Para reproducir la evidencia desde la raíz del repositorio:

```powershell
python tests_python/ground_truth_semana4.py | python tests_python/generate_gantt.py --input - --output-dir docs/gantt --algorithms FCFS SJF RR PRIORITY
python tests_python/memory_audit.py
python tests_python/page_replacement.py --file tests_python/fixtures/referencias_belady_3marcos.json
python tests_python/page_replacement.py --file tests_python/fixtures/referencias_belady_4marcos.json
python -m pytest -q tests_python/test_ground_truth.py tests_python/test_page_replacement.py tests_python/test_generate_gantt.py tests_python/test_memory_audit.py
```

El generador de Gantt también acepta un JSON individual de `cpu_timeline.json` por `--input`. Los SVG enlazados abajo se generaron con la salida del oráculo sobre el caso documentado; deben regenerarse con la salida C cuando esté disponible.

## 2. Planificación de CPU y Gantt

Los minutos son ticks simulados. Para cada proceso, `retorno = finalización − llegada`, `espera = retorno − ráfaga` y `respuesta = primer inicio − llegada`. El oráculo emplea llegada y PID para desempatar, y en RR coloca las llegadas ocurridas hasta el final del quantum antes de reinsertar el proceso desalojado.

| Algoritmo | Espera media (min) | Retorno medio (min) | Respuesta media (min) | Diagrama |
|---|---:|---:|---:|---|
| FCFS | 5.750 | 11.250 | 5.750 | [SVG](gantt/gantt_fcfs.svg) |
| SJF no apropiativo | 5.250 | 10.750 | 5.250 | [SVG](gantt/gantt_sjf.svg) |
| RR, quantum 4 | 9.250 | 14.750 | 4.000 | [SVG](gantt/gantt_rr.svg) |
| Prioridad no apropiativa | 7.750 | 13.250 | 7.750 | [SVG](gantt/gantt_priority.svg) |

Los promedios de espera coinciden con los cuatro resultados exigidos por el plan. Los SVG muestran la secuencia de CPU, los tramos de espera entre llegada y ejecución, eje de tiempo, color por proceso y los tres promedios.

![Gantt FCFS generado con el oráculo](gantt/gantt_fcfs.svg)

![Gantt SJF generado con el oráculo](gantt/gantt_sjf.svg)

![Gantt RR generado con el oráculo](gantt/gantt_rr.svg)

![Gantt de prioridad generado con el oráculo](gantt/gantt_priority.svg)

## 3. Auditoría de memoria

### 3.1 Asignación contigua y fragmentación externa

First Fit elige el primer hueco suficiente; Best Fit elige el más pequeño que alcance para la solicitud. Ambos requieren un hueco **contiguo**. Sumar toda la memoria libre no basta para decidir si cabe un bloque. Esta distinción corresponde a la asignación contigua descrita en la sección 9.2 de *Operating System Concepts* y al capítulo de administración de memoria de *Modern Operating Systems*; el mapa siguiente es un **caso de prueba construido** para el criterio de fragmentación del plan, no una salida observada del motor C. [Silberschatz, Galvin y Gagne (Wiley)](https://www.wiley-vch.de/en/areas-interest/computing-computer-sciences/computer-science-17cs/general-introductory-computer-science-17cs0/silberschatz-39-s-operating-system-concepts-978-1-119-45408-3), [Tanenbaum y Bos (Pearson)](https://www.pearson.com/en-us/subject-catalog/p/modern-operating-systems/P200000003295/9780137618934).

| Bloque inicial (unidades abstractas) | Inicio | Tamaño | Estado |
|---|---:|---:|---|
| 1 | 0 | 2 | Ocupado |
| 2 | 2 | 2 | Libre |
| 3 | 4 | 1 | Ocupado |
| 4 | 5 | 2 | Libre |
| 5 | 7 | 1 | Ocupado |
| 6 | 8 | 2 | Libre |

**Solicitud:** 3 unidades contiguas. **Resultado esperado:** rechazo por fragmentación externa; hay 6 unidades libres en total, pero el hueco mayor tiene 2. El mapa no cambia tras el rechazo. `fragmentation()` reproduce ese dictamen y une huecos libres adyacentes antes de medir su tamaño. Falta contrastarlo con `memory_map.json` y la alerta del motor C del integrante 2.

### 3.2 Paginación y traducción

Con páginas y marcos de **4096 bytes**, la dirección virtual **10000** se descompone en página `10000 // 4096 = 2` y desplazamiento `10000 % 4096 = 1808`. La tabla documentada asigna la página 2 al marco 7; por tanto, `7 × 4096 + 1808 = 30480` es la dirección física. El oráculo [`memory_audit.py`](../tests_python/memory_audit.py) y su prueba reproducen el cálculo. Falta comprobar que `memory_paging.c` emita exactamente esos valores cuando exista. El marco teórico de paginación está en la sección 9.3 de [*Operating System Concepts* (Wiley)](https://www.wiley-vch.de/en/areas-interest/computing-computer-sciences/computer-science-17cs/general-introductory-computer-science-17cs0/silberschatz-39-s-operating-system-concepts-978-1-119-45408-3).

### 3.3 Segmentación y límites

El límite es la **longitud** del segmento: solo son válidos los desplazamientos `0 <= offset < límite`. La tabla del plan es:

| Segmento | Base | Límite | Rango físico válido |
|---|---:|---:|---|
| 0, Código | 1000 | 400 | 1000–1399 |
| 1, Datos | 5200 | 800 | 5200–5999 |
| 2, Pila | 9000 | 300 | 9000–9299 |

Para Datos, `(1, 300)` se traduce a `5200 + 300 = 5500`. `(1, 900)` excede el límite y produce `EXCEPTION_SEGMENTATION_FAULT`. El desplazamiento **800**, igual al límite, también debe fallar. Estos resultados están probados en Python; la comparación con el módulo C está pendiente.

## 4. Reemplazo de páginas: FIFO, LRU y Óptimo

Los tres algoritmos reciben exactamente la misma secuencia y capacidad. FIFO expulsa la página residente más antigua; LRU, la menos usada recientemente; Óptimo elige la página cuyo próximo uso está más lejos, usando el futuro de la secuencia **solo como referencia teórica**. Se cuentan fallos cuando la página solicitada no está en los marcos. Las salidas completas contienen, para cada referencia, posición, página, marcos, acierto/fallo y víctima. El marco teórico de reemplazo corresponde a la sección 10.4 de [*Operating System Concepts* (Wiley)](https://www.wiley-vch.de/en/areas-interest/computing-computer-sciences/computer-science-17cs/general-introductory-computer-science-17cs0/silberschatz-39-s-operating-system-concepts-978-1-119-45408-3) y al capítulo 3 de [*Modern Operating Systems* (Pearson)](https://www.pearson.com/en-us/subject-catalog/p/modern-operating-systems/P200000003295/9780137618934).

| Marcos | FIFO: fallos / aciertos | LRU: fallos / aciertos | Óptimo: fallos / aciertos | Traza JSON |
|---:|---:|---:|---:|---|
| 3 | 9 / 3 | 10 / 2 | 7 / 5 | [Traza completa](traces/reemplazo_3marcos.json) |
| 4 | 10 / 2 | 8 / 4 | 6 / 6 | [Traza completa](traces/reemplazo_4marcos.json) |

Al subir de 3 a 4 marcos, FIFO pasa de **9 a 10 fallos**: reproduce la anomalía de Belady requerida. Óptimo tiene menos fallos que FIFO y LRU en ambas capacidades. Esto no implica que LRU supere siempre a FIFO: en esta secuencia con 3 marcos, LRU tiene un fallo más.

Detalle de FIFO con 3 marcos (cada fila es un acceso; `–` indica marco aún libre):

| Índice | Página | Marcos tras acceso | Resultado | Víctima |
|---:|---:|---|---|---:|
| 0 | 1 | 1, –, – | Fallo | – |
| 1 | 2 | 1, 2, – | Fallo | – |
| 2 | 3 | 1, 2, 3 | Fallo | – |
| 3 | 4 | 4, 2, 3 | Fallo | 1 |
| 4 | 1 | 4, 1, 3 | Fallo | 2 |
| 5 | 2 | 4, 1, 2 | Fallo | 3 |
| 6 | 5 | 5, 1, 2 | Fallo | 4 |
| 7 | 1 | 5, 1, 2 | Acierto | – |
| 8 | 2 | 5, 1, 2 | Acierto | – |
| 9 | 3 | 5, 3, 2 | Fallo | 1 |
| 10 | 4 | 5, 3, 4 | Fallo | 2 |
| 11 | 5 | 5, 3, 4 | Acierto | – |

Las entradas para repetirlo están en [`referencias_belady_3marcos.json`](../tests_python/fixtures/referencias_belady_3marcos.json) y [`referencias_belady_4marcos.json`](../tests_python/fixtures/referencias_belady_4marcos.json). `page_replacement.py` acepta esos JSON por `--file` o por `stdin`; emite un objeto con FIFO, LRU y OPTIMAL apto para que el servidor Go lo exponga. La vista web y el endpoint `/api/v1/memory/replace` corresponden al integrante 1 y aún no se han verificado en esta rama.

## 5. Estado de validación e integración

| Entregable | Evidencia en esta rama | Pendiente externo |
|---|---|---|
| Gantt exacto y métricas | Generador SVG y cuatro gráficos del caso documentado | Regenerar y comparar desde la salida C; presentar en la interfaz web |
| Paginación y segmentación | Oráculo y pruebas de `30480`, `5500` y trap | Contrastar con `memory_paging.c` y el motor de segmentación |
| Asignación contigua | Caso reproducible de fragmentación externa y dictamen teórico | Comprobar mapa inicial/final y alerta emitidos por C |
| Reemplazo | FIFO/LRU/Óptimo, tabla, trazas y caso de Belady | Conectar endpoint y mostrar tabla/traza en la web |

**Criterio para cerrar el informe Beta:** sustituir cada pendiente por una salida real del módulo correspondiente, conservar entrada y comando usados, comprobar las diferencias contra los oráculos y actualizar esta tabla. Hasta entonces, el informe documenta la evidencia de QA disponible y las verificaciones por integrar.
