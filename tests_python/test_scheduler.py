"""Diez casos de caja negra: compara el ejecutable C con el oráculo Python."""
import json
import os
import shutil
import subprocess
from pathlib import Path

import pytest

from ground_truth_semana4 import schedule

ROOT = Path(__file__).resolve().parents[1]
SOURCES = [ROOT / "src/core_c" / file for file in
           ("scheduler.c", "queue.c", "scheduler_nonpreemptive.c", "scheduler_preemptive.c")]
BASE_PATH = ROOT / "tests_python/fixtures/procesos_semana4.json"


@pytest.fixture(scope="session")
def binary(tmp_path_factory):
    missing = [str(source.relative_to(ROOT)) for source in SOURCES if not source.is_file()]
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
    try:
        completed = subprocess.run([str(binary), "--algo", algorithm, "--file", str(input_file),
                                    "--quantum", str(quantum), "--aging-interval", str(aging),
                                    "--output", str(output_file)], capture_output=True, text=True,
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
    assert actual == expected, f"{name}: {first_difference(expected, actual)}"
    if expected_wait is not None:
        assert actual["average_waiting_time"] == expected_wait
    if name == "base_rr":
        assert actual["average_response_time"] == 4.0
