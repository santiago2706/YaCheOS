"""Casos de caja negra: compara el ejecutable C con el oráculo Python."""
import json
import os
import shutil
import subprocess
from pathlib import Path

import pytest

from ground_truth_semana4 import schedule

ROOT = Path(__file__).resolve().parents[1]
CORE_DIR = Path(os.environ.get("YAKCHEOS_CORE_C_DIR", ROOT / "src/core_c"))
SPLIT_SOURCES = [CORE_DIR / file for file in
                 ("queue.c", "scheduler_nonpreemptive.c", "scheduler_preemptive.c")]
SOURCES = [CORE_DIR / "scheduler.c"] + (SPLIT_SOURCES if any(path.exists() for path in SPLIT_SOURCES) else [])
HEADERS = [CORE_DIR / "pcb.h"]
BASE_PATH = ROOT / "tests_python/fixtures/procesos_semana4.json"


@pytest.fixture(scope="session")
def binary(tmp_path_factory):
    missing = [str(source) + (" (vacio)" if source.is_file() else " (no existe)")
               for source in SOURCES + HEADERS if not source.is_file() or source.stat().st_size == 0]
    if not BASE_PATH.is_file() or not BASE_PATH.read_text(encoding="utf-8").strip():
        missing.append(str(BASE_PATH.relative_to(ROOT)) + " (sin datos)")
    if missing:
        pytest.fail("Faltan dependencias del equipo para auditar C: " + ", ".join(missing), pytrace=False)
    if shutil.which("gcc") is None:
        pytest.fail("No se encontró gcc en PATH; instala GCC o añade su carpeta bin a PATH", pytrace=False)
    path = tmp_path_factory.mktemp("core") / ("scheduler.exe" if os.name == "nt" else "scheduler")
    try:
        subprocess.run(["gcc", "-std=c11", "-Wall", "-Wextra", "-Werror", *map(str, SOURCES), "-o", str(path)],
                       check=True, capture_output=True, text=True)
    except subprocess.CalledProcessError as error:
        pytest.fail("No se pudo compilar el planificador C:\n" + error.stderr, pytrace=False)
    except OSError as error:
        pytest.fail(f"No se pudo iniciar GCC: {error}. Verifica la instalación y PATH", pytrace=False)
    return path


CASES = [
    ("base_fcfs", None, "FCFS", 4, 4, 5.75),
    ("base_sjf", None, "SJF", 4, 4, 5.25),
    ("base_rr", None, "RR", 4, 4, 9.25),
    ("base_priority", None, "PRIORITY", 4, 4, 7.75),
    ("base_preemptive", None, "PRIORITY_PREEMPTIVE", 4, 4, None),
    ("base_aging", None, "PRIORITY_AGING", 4, 2, None),
    ("same_arrival", [{"pid": 2, "arrival_time": 0, "burst_time": 2, "priority": 2},
                      {"pid": 1, "arrival_time": 0, "burst_time": 1, "priority": 1}], "FCFS", 4, 4, None),
    ("idle_cpu", [{"pid": 1, "arrival_time": 1000000, "burst_time": 2, "priority": 2}], "RR", 4, 4, None),
    ("rr_arrival_boundary", [{"pid": 1, "arrival_time": 0, "burst_time": 5, "priority": 1},
                             {"pid": 2, "arrival_time": 2, "burst_time": 1, "priority": 2}], "RR", 2, 4, None),
    ("priority_preemption", [{"pid": 1, "arrival_time": 0, "burst_time": 4, "priority": 4},
                             {"pid": 2, "arrival_time": 1, "burst_time": 1, "priority": 1}], "PRIORITY_PREEMPTIVE", 4, 4, None),
    ("rr_unsorted_arrivals", [{"pid": 3, "arrival_time": 0, "burst_time": 4, "priority": 1},
                              {"pid": 2, "arrival_time": 2, "burst_time": 1, "priority": 1},
                              {"pid": 1, "arrival_time": 1, "burst_time": 1, "priority": 1}], "RR", 3, 4, None),
    ("aging_resumes", [{"pid": 1, "arrival_time": 1, "burst_time": 4, "priority": 5},
                       {"pid": 2, "arrival_time": 7, "burst_time": 7, "priority": 4},
                       {"pid": 3, "arrival_time": 0, "burst_time": 5, "priority": 2},
                       {"pid": 4, "arrival_time": 4, "burst_time": 5, "priority": 3}], "PRIORITY_AGING", 4, 3, None),
    ("rr_long_timeline", [{"pid": 1, "arrival_time": 0, "burst_time": 300, "priority": 1},
                          {"pid": 2, "arrival_time": 0, "burst_time": 300, "priority": 1}], "RR", 1, 4, None),
]


def first_difference(expected, actual, path="$"):
    if type(expected) is not type(actual):
        return f"{path}: tipo esperado={type(expected).__name__}, obtenido={type(actual).__name__}"
    if isinstance(expected, dict):
        if expected.keys() != actual.keys():
            return f"{path}: campos esperados={sorted(expected)}, obtenidos={sorted(actual)}"
        for key in expected:
            difference = first_difference(expected[key], actual[key], f"{path}.{key}")
            if difference:
                return difference
    elif isinstance(expected, list):
        if len(expected) != len(actual):
            return f"{path}: longitud esperada={len(expected)}, obtenida={len(actual)}"
        for index, (expected_item, actual_item) in enumerate(zip(expected, actual)):
            difference = first_difference(expected_item, actual_item, f"{path}[{index}]")
            if difference:
                return difference
    elif expected != actual:
        return f"{path}: esperado={expected!r}, obtenido={actual!r}"
    return None


def complete_idle_intervals(segments):
    """El motor C omite intervalos inactivos; reconstruirlos desde los tiempos emitidos."""
    timeline = []
    previous_end = 0
    for segment in segments:
        start, end = segment["start"], segment["end"]
        assert start >= previous_end and end > start, "Traza C solapada o con bloque vacío"
        if start > previous_end:
            timeline.append({"pid": 0, "start": previous_end, "end": start})
        timeline.append({key: segment[key] for key in ("pid", "start", "end")})
        previous_end = end
    return timeline


@pytest.mark.parametrize("name,processes,algorithm,quantum,aging,expected_wait", CASES, ids=[case[0] for case in CASES])
def test_c_matches_oracle(binary, tmp_path, name, processes, algorithm, quantum, aging, expected_wait):
    if processes is None:
        try:
            processes = json.loads(BASE_PATH.read_text(encoding="utf-8"))
        except json.JSONDecodeError as error:
            pytest.fail(f"{BASE_PATH}: JSON inválido: {error}", pytrace=False)
    input_file = tmp_path / f"{name}.json"
    output_file = tmp_path / "cpu_timeline.json"
    input_file.write_text(json.dumps(processes), encoding="utf-8")
    c_algorithm = "PRIORITY_PREEMPTIVE" if algorithm == "PRIORITY_AGING" else algorithm
    c_aging = aging if algorithm == "PRIORITY_AGING" else 0
    try:
        completed = subprocess.run([str(binary), "--algo", c_algorithm, "--file", str(input_file),
                                    "--quantum", str(quantum), "--aging", str(c_aging),
                                    "--out", str(output_file)], capture_output=True, text=True,
                                   check=True, timeout=60)
    except subprocess.CalledProcessError as error:
        pytest.fail(f"{name}: el planificador C terminó con código {error.returncode}:\n{error.stderr}", pytrace=False)
    except subprocess.TimeoutExpired:
        pytest.fail(f"{name}: el planificador C excedió 60 segundos", pytrace=False)
    if not output_file.is_file():
        pytest.fail(f"{name}: el planificador C no escribió {output_file.name}", pytrace=False)
    actual = json.loads(output_file.read_text(encoding="utf-8"))
    assert actual == json.loads(completed.stdout), "stdout y cpu_timeline.json difieren"
    expected = schedule(processes, algorithm, quantum, aging)
    expected_c_algorithm = "PRIORITY_PREEMPTIVE_AGING" if algorithm == "PRIORITY_AGING" else algorithm
    assert actual["algorithm"] == expected_c_algorithm
    if algorithm == "RR":
        assert actual["quantum"] == quantum
    if algorithm == "PRIORITY_AGING":
        assert actual["aging_interval"] == aging
    assert actual["total_time"] == max(row["completion"] for row in expected["processes"])
    for key in ("waiting_time", "turnaround_time", "response_time"):
        # C publica dos decimales; el oráculo conserva tres. Comparar con el
        # promedio exacto evita rechazar 1.33 frente a 1.333 o aceptar un
        # redondeo incorrecto basado en una media ya redondeada.
        exact = sum(row[key] for row in expected["processes"]) / len(expected["processes"])
        reported = actual[f"average_{key}"]
        assert reported == round(reported, 2)
        assert abs(reported - exact) <= 0.005000001, (
            f"{name}: average_{key}={reported}, promedio exacto={exact}"
        )
        assert actual["metrics"][f"avg_{key}"] == actual[f"average_{key}"]
    timeline = complete_idle_intervals(actual["timeline"])
    assert timeline == expected["timeline"], f"{name}: {first_difference(expected['timeline'], timeline)}"
    assert all(segment["duration"] == segment["end"] - segment["start"]
               for segment in actual["timeline"])
    rows = [{"pid": row["pid"], "arrival_time": row["arrival_time"],
             "burst_time": row["burst_time"], "priority": row["priority"],
             "first_start": row["start_time"], "completion": row["completion_time"],
             "waiting_time": row["waiting_time"], "turnaround_time": row["turnaround_time"],
             "response_time": row["response_time"]} for row in actual["processes"]]
    assert rows == expected["processes"], f"{name}: {first_difference(expected['processes'], rows)}"
    assert all(row["arrival"] == row["arrival_time"] and row["burst"] == row["burst_time"]
               for row in actual["processes"])
    if expected_wait is not None:
        assert actual["average_waiting_time"] == expected_wait
    if name == "base_rr":
        assert actual["average_response_time"] == 4.0


@pytest.mark.parametrize("processes", [
    [{"pid": 1, "arrival_time": 0, "burst_time": 2, "priority": 1},
     {"pid": 1, "arrival_time": 1, "burst_time": 1, "priority": 1}],
    [{"pid": 0, "arrival_time": 0, "burst_time": 1, "priority": 1}],
    [{"pid": -2, "arrival_time": 0, "burst_time": 1, "priority": 1}],
    [{"pid": 1, "arrival_time": 0, "burst_time": 1, "priority": 0}],
], ids=["pid_duplicado", "pid_cero", "pid_negativo", "prioridad_cero"])
def test_c_rejects_invalid_processes_without_result(binary, tmp_path, processes):
    source = tmp_path / "invalid.json"
    output = tmp_path / "cpu_timeline.json"
    source.write_text(json.dumps(processes), encoding="utf-8")
    with pytest.raises(ValueError):
        schedule(processes, "RR")
    completed = subprocess.run([str(binary), "--algo", "RR", "--file", str(source),
                                "--out", str(output)], capture_output=True, text=True)
    assert completed.returncode != 0
    assert not completed.stdout
    assert not output.exists()
