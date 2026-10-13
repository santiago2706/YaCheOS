"""Oráculo independiente para la planificación de CPU del contrato Semana 4."""
import argparse
import json
from collections import deque
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
ALGORITHMS = ("FCFS", "SJF", "RR", "PRIORITY", "PRIORITY_PREEMPTIVE", "PRIORITY_AGING")


def schedule(processes, algorithm, quantum=4, aging_interval=4):
    if (algorithm not in ALGORITHMS or type(quantum) is not int or type(aging_interval) is not int
            or not 1 <= quantum <= 1000000 or not 1 <= aging_interval <= 1000000):
        raise ValueError("Algoritmo o parámetros inválidos")
    if not isinstance(processes, list) or not 1 <= len(processes) <= 128:
        raise ValueError("Se requieren entre 1 y 128 procesos")
    rows = []
    for raw in processes:
        if not isinstance(raw, dict) or any(type(raw.get(key)) is not int
                                            for key in ("pid", "arrival_time", "burst_time", "priority")):
            raise ValueError("Cada proceso requiere PID, llegada, ráfaga y prioridad enteros")
        p = {key: raw[key] for key in ("pid", "arrival_time", "burst_time", "priority")}
        if (not 1 <= p["pid"] <= 1000000 or not 0 <= p["arrival_time"] <= 1000000
                or not 1 <= p["burst_time"] <= 1000000 or not 1 <= p["priority"] <= 1000000):
            raise ValueError("PID, llegada, ráfaga o prioridad inválida")
        p.update(remaining=p["burst_time"], first_start=None, completion=None, waited=0)
        rows.append(p)
    if len({p["pid"] for p in rows}) != len(rows):
        raise ValueError("PID duplicado")

    timeline = []

    def append(pid, start, end):
        if timeline and timeline[-1]["pid"] == pid and timeline[-1]["end"] == start:
            timeline[-1]["end"] = end
        else:
            timeline.append(dict(pid=pid, start=start, end=end))

    time = 0
    done = 0
    pending = sorted(rows, key=lambda p: (p["arrival_time"], p["pid"]))
    ready = deque()
    next_arrival = 0

    def enqueue_arrivals(until):
        nonlocal next_arrival
        while next_arrival < len(pending) and pending[next_arrival]["arrival_time"] <= until:
            ready.append(pending[next_arrival])
            next_arrival += 1

    while done < len(rows):
        if algorithm == "RR":
            enqueue_arrivals(time)
            if not ready:
                until = pending[next_arrival]["arrival_time"]
                append(0, time, until)
                time = until
                continue
            p = ready.popleft()
            if p["first_start"] is None:
                p["first_start"] = time
            span = min(quantum, p["remaining"])
            append(p["pid"], time, time + span)
            time += span
            p["remaining"] -= span
            enqueue_arrivals(time)  # llegadas antes de reinsertar el desalojado
            if p["remaining"]:
                ready.append(p)
            else:
                p["completion"] = time
                done += 1
            continue

        candidates = [p for p in rows if p["arrival_time"] <= time and p["remaining"]]
        if not candidates:
            until = min(p["arrival_time"] for p in rows if p["remaining"])
            append(0, time, until)
            time = until
            continue
        tie = lambda p: (p["arrival_time"], p["pid"])
        if algorithm == "FCFS":
            p = min(candidates, key=tie)
        elif algorithm == "SJF":
            p = min(candidates, key=lambda x: (x["burst_time"], *tie(x)))
        elif algorithm == "PRIORITY":
            p = min(candidates, key=lambda x: (x["priority"], *tie(x)))
        else:
            p = min(candidates, key=lambda x: (max(1, x["priority"] - x["waited"] // aging_interval)
                             if algorithm == "PRIORITY_AGING" else x["priority"], *tie(x)))
        if p["first_start"] is None:
            p["first_start"] = time
        if algorithm in ("FCFS", "SJF", "PRIORITY"):
            span = p["remaining"]
        else:
            span = 1
            for other in candidates:
                if other is not p:
                    other["waited"] += 1
        append(p["pid"], time, time + span)
        time += span
        p["remaining"] -= span
        if not p["remaining"]:
            p["completion"] = time
            done += 1

    result = []
    for p in rows:
        turnaround = p["completion"] - p["arrival_time"]
        result.append({key: p[key] for key in ("pid", "arrival_time", "burst_time", "priority")}
                      | dict(first_start=p["first_start"], completion=p["completion"],
                             waiting_time=turnaround - p["burst_time"], turnaround_time=turnaround,
                             response_time=p["first_start"] - p["arrival_time"]))
    average = lambda key: round(sum(p[key] for p in result) / len(result), 3)
    return dict(algorithm=algorithm, quantum=quantum, aging_interval=aging_interval,
                timeline=timeline, processes=result,
                average_waiting_time=average("waiting_time"),
                average_turnaround_time=average("turnaround_time"),
                average_response_time=average("response_time"))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--file", type=Path, default=ROOT / "data/procesos_semana4.json")
    parser.add_argument("--algo", choices=ALGORITHMS)
    parser.add_argument("--quantum", type=int, default=4)
    parser.add_argument("--aging-interval", type=int, default=4)
    args = parser.parse_args()
    cases = json.loads(args.file.read_text(encoding="utf-8"))
    results = {a: schedule(cases, a, args.quantum, args.aging_interval)
               for a in ([args.algo] if args.algo else ALGORITHMS)}
    print(json.dumps(results, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    main()
