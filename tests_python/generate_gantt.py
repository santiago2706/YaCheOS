"""Genera diagramas SVG desde la salida JSON del planificador u oráculo.

Uso: python tests_python/generate_gantt.py --input salida.json --output-dir docs/gantt
Acepta un resultado individual o un objeto con resultados por algoritmo.
"""

import argparse
import json
import sys
from html import escape
from pathlib import Path

from ground_truth_semana4 import ALGORITHMS


COLORS = ("#2563eb", "#db2777", "#059669", "#d97706", "#7c3aed", "#0891b2")


def _integer(value, label, minimum=0):
    if type(value) is not int or value < minimum:
        raise ValueError(f"{label} debe ser un entero >= {minimum}")
    return value


def _validate(result):
    if not isinstance(result, dict) or result.get("algorithm") not in ALGORITHMS:
        raise ValueError("Resultado sin algoritmo reconocido")
    timeline = result.get("timeline")
    processes = result.get("processes")
    if not isinstance(timeline, list) or not timeline or not isinstance(processes, list) or not processes:
        raise ValueError("Se requieren timeline y processes no vacíos")
    rows = {}
    for row in processes:
        if not isinstance(row, dict):
            raise ValueError("Cada proceso debe ser un objeto JSON")
        pid = _integer(row.get("pid"), "pid", 1)
        if pid in rows:
            raise ValueError(f"PID duplicado: {pid}")
        arrival = _integer(row.get("arrival_time"), "arrival_time")
        completion = _integer(row.get("completion"), "completion")
        burst = _integer(row.get("burst_time"), "burst_time", 1)
        if completion < arrival + burst:
            raise ValueError(f"P{pid}: finalización anterior a su ráfaga")
        rows[pid] = row

    previous_end = 0
    runtime = {pid: 0 for pid in rows}
    first_start = {}
    last_end = {}
    for block in timeline:
        if not isinstance(block, dict):
            raise ValueError("Cada bloque del timeline debe ser un objeto JSON")
        pid = _integer(block.get("pid"), "timeline.pid")
        start = _integer(block.get("start"), "timeline.start")
        end = _integer(block.get("end"), "timeline.end")
        if start != previous_end or end <= start:
            raise ValueError("El timeline debe ser continuo desde t=0 y tener bloques positivos")
        if pid == 0:
            if any(row["arrival_time"] < end and row["completion"] > start
                   for row in rows.values()):
                raise ValueError("La CPU figura inactiva mientras hay un proceso listo")
        else:
            if pid not in rows:
                raise ValueError(f"El timeline contiene P{pid} sin métricas")
            if start < rows[pid]["arrival_time"] or end > rows[pid]["completion"]:
                raise ValueError(f"P{pid}: ejecución fuera de llegada/finalización")
            runtime[pid] += end - start
            first_start.setdefault(pid, start)
            last_end[pid] = end
        previous_end = end
    for pid, row in rows.items():
        if runtime[pid] != row["burst_time"]:
            raise ValueError(f"P{pid}: duración ejecutada {runtime[pid]} != ráfaga {row['burst_time']}")
        if last_end[pid] != row["completion"]:
            raise ValueError(f"P{pid}: finalización no coincide con el último bloque")
        turnaround = row["completion"] - row["arrival_time"]
        expected = {"first_start": first_start[pid], "turnaround_time": turnaround,
                    "waiting_time": turnaround - row["burst_time"],
                    "response_time": first_start[pid] - row["arrival_time"]}
        for key, value in expected.items():
            if row.get(key) != value:
                raise ValueError(f"P{pid}: {key}={row.get(key)!r}, esperado={value}")
    if previous_end != max(row["completion"] for row in rows.values()):
        raise ValueError("El timeline continúa tras finalizar todos los procesos")
    for key in ("average_waiting_time", "average_turnaround_time", "average_response_time"):
        value = result.get(key)
        if type(value) not in (int, float) or value < 0:
            raise ValueError(f"Falta una métrica válida: {key}")
        process_key = key.removeprefix("average_")
        expected = round(sum(row[process_key] for row in rows.values()) / len(rows), 3)
        if value != expected:
            raise ValueError(f"{key}={value}, esperado={expected}")
    return rows, previous_end


def _tick_step(duration):
    step = 1
    while duration / step > 16:
        step *= 2 if step in (1, 5) else 2.5
    return int(step)


def render_svg(result):
    """Devuelve un SVG autocontenido; no necesita matplotlib ni recursos externos."""
    rows, duration = _validate(result)
    plot_width = min(1300, max(620, duration * 28))
    scale = plot_width / duration
    left, right = 110, 32
    width = left + plot_width + right
    cpu_y, row_start, row_height = 116, 181, 46
    axis_bottom = row_start + len(rows) * row_height + 6
    metric_y = axis_bottom + 66
    height = metric_y + 43
    parts = [
        f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width:g} {height:g}" role="img" aria-labelledby="title desc">',
        f'<title id="title">Gantt {escape(result["algorithm"])}</title>',
        '<desc id="desc">Línea de CPU y tramos de ejecución y espera por proceso, en minutos simulados.</desc>',
        '<defs><pattern id="waiting" width="7" height="7" patternUnits="userSpaceOnUse" patternTransform="rotate(45)"><rect width="7" height="7" fill="#e2e8f0"/><line x1="0" y1="0" x2="0" y2="7" stroke="#94a3b8" stroke-width="2"/></pattern></defs>',
        f'<rect width="{width:g}" height="{height:g}" fill="#fff"/>',
        f'<text x="{left}" y="37" font-family="Arial,sans-serif" font-size="22" font-weight="700" fill="#0f172a">Planificación {escape(result["algorithm"])}</text>',
        f'<text x="{left}" y="62" font-family="Arial,sans-serif" font-size="12" fill="#475569">Tiempo simulado (min) · color: ejecución · trama: espera</text>',
    ]
    sorted_pids = sorted(rows)
    colors = {pid: COLORS[index % len(COLORS)] for index, pid in enumerate(sorted_pids)}

    step = _tick_step(duration)
    ticks = list(range(0, duration + 1, step))
    if ticks[-1] != duration:
        ticks.append(duration)
    for tick in ticks:
        x = left + tick * scale
        parts.append(f'<line x1="{x:g}" y1="82" x2="{x:g}" y2="{axis_bottom:g}" stroke="#e2e8f0" stroke-width="1"/>')
        parts.append(f'<text x="{x:g}" y="78" text-anchor="middle" font-family="Arial,sans-serif" font-size="11" fill="#475569">{tick}</text>')

    parts.append(f'<text x="{left - 12}" y="{cpu_y + 21}" text-anchor="end" font-family="Arial,sans-serif" font-size="13" font-weight="700" fill="#0f172a">CPU</text>')
    for block in result["timeline"]:
        pid, start, end = block["pid"], block["start"], block["end"]
        x, block_width = left + start * scale, (end - start) * scale
        label = f'P{pid}' if pid else 'Inactivo'
        color = colors.get(pid, "#cbd5e1")
        parts.append(f'<g><title>{escape(label)}: {start}–{end} min</title><rect data-kind="cpu" data-pid="{pid}" data-start="{start}" data-end="{end}" x="{x:g}" y="{cpu_y}" width="{block_width:g}" height="34" fill="{color}" stroke="#fff"/></g>')
        if block_width >= 27:
            parts.append(f'<text x="{x + block_width / 2:g}" y="{cpu_y + 22}" text-anchor="middle" font-family="Arial,sans-serif" font-size="12" font-weight="700" fill="{("#334155" if pid == 0 else "#fff")}">{escape(label)}</text>')

    for index, pid in enumerate(sorted_pids):
        row = rows[pid]
        y = row_start + index * row_height
        parts.append(f'<text x="{left - 12}" y="{y + 20}" text-anchor="end" font-family="Arial,sans-serif" font-size="13" font-weight="700" fill="{colors[pid]}">P{pid}</text>')
        runs = [(block["start"], block["end"]) for block in result["timeline"] if block["pid"] == pid]
        cursor = row["arrival_time"]
        for start, end in runs:
            if cursor < start:
                x = left + cursor * scale
                parts.append(f'<g><title>P{pid} espera: {cursor}–{start} min</title><rect data-kind="wait" data-pid="{pid}" data-start="{cursor}" data-end="{start}" x="{x:g}" y="{y}" width="{(start - cursor) * scale:g}" height="28" fill="url(#waiting)" stroke="#94a3b8"/></g>')
            x = left + start * scale
            parts.append(f'<g><title>P{pid} ejecuta: {start}–{end} min</title><rect data-kind="run" data-pid="{pid}" data-start="{start}" data-end="{end}" x="{x:g}" y="{y}" width="{(end - start) * scale:g}" height="28" fill="{colors[pid]}" stroke="#fff"/></g>')
            cursor = end
        if cursor < row["completion"]:
            x = left + cursor * scale
            parts.append(f'<g><title>P{pid} espera: {cursor}–{row["completion"]} min</title><rect data-kind="wait" data-pid="{pid}" data-start="{cursor}" data-end="{row["completion"]}" x="{x:g}" y="{y}" width="{(row["completion"] - cursor) * scale:g}" height="28" fill="url(#waiting)" stroke="#94a3b8"/></g>')

    parts.append(f'<line x1="{left}" y1="{axis_bottom + 15:g}" x2="{width - right:g}" y2="{axis_bottom + 15:g}" stroke="#cbd5e1"/>')
    metrics = (("Espera media", result["average_waiting_time"]),
               ("Retorno medio", result["average_turnaround_time"]),
               ("Respuesta media", result["average_response_time"]))
    for index, (label, value) in enumerate(metrics):
        x = left + index * plot_width / 3
        parts.append(f'<text x="{x:g}" y="{metric_y}" font-family="Arial,sans-serif" font-size="12" fill="#475569">{label}</text>')
        parts.append(f'<text x="{x:g}" y="{metric_y + 23}" font-family="Arial,sans-serif" font-size="17" font-weight="700" fill="#0f172a">{value:.3f} min</text>')
    parts.append('</svg>')
    return "\n".join(parts) + "\n"


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", required=True, help="JSON del planificador u oráculo; '-' lee stdin")
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--algorithms", nargs="+", choices=ALGORITHMS)
    args = parser.parse_args(argv)
    try:
        with sys.stdin if args.input == "-" else open(args.input, encoding="utf-8") as source:
            payload = json.load(source)
        results = {payload["algorithm"]: payload} if isinstance(payload, dict) and "algorithm" in payload else payload
        if not isinstance(results, dict) or not results:
            raise ValueError("El JSON debe contener un resultado o un mapa de algoritmos")
        selected = args.algorithms or list(results)
        if any(name not in results or name not in ALGORITHMS for name in selected):
            raise ValueError("Algoritmo solicitado ausente en el JSON")
        if any(not isinstance(results[name], dict) or results[name].get("algorithm") != name
               for name in selected):
            raise ValueError("La clave del algoritmo no coincide con el resultado")
        rendered = {name: render_svg(results[name]) for name in selected}
        args.output_dir.mkdir(parents=True, exist_ok=True)
        for name, svg in rendered.items():
            target = args.output_dir / f"gantt_{name.lower()}.svg"
            target.write_text(svg, encoding="utf-8")
            print(target)
    except (OSError, ValueError, KeyError, TypeError, json.JSONDecodeError) as error:
        parser.exit(2, f"Error al generar Gantt: {error}\n")


if __name__ == "__main__":
    main()
