"""Diez casos de caja negra: compara el ejecutable C con el oráculo Python."""
import json
import os
import subprocess
from pathlib import Path

import pytest

from ground_truth_semana4 import schedule

ROOT = Path(__file__).resolve().parents[1]
SOURCES = [ROOT / "src/core_c" / file for file in
           ("scheduler.c", "queue.c", "scheduler_nonpreemptive.c", "scheduler_preemptive.c")]
BASE = json.loads((ROOT / "data/procesos_semana4.json").read_text(encoding="utf-8"))


@pytest.fixture(scope="session")
def binary(tmp_path_factory):
    path = tmp_path_factory.mktemp("core") / ("scheduler.exe" if os.name == "nt" else "scheduler")
    subprocess.run(["gcc", "-std=c11", "-Wall", "-Wextra", "-Werror", *map(str, SOURCES), "-o", str(path)],
                   check=True, capture_output=True, text=True)
    return path


CASES = [
    ("base_fcfs", BASE, "FCFS", 4, 4, 5.75),
    ("base_sjf", BASE, "SJF", 4, 4, 5.25),
    ("base_rr", BASE, "RR", 4, 4, 9.25),
    ("base_priority", BASE, "PRIORITY", 4, 4, 7.75),
    ("base_preemptive", BASE, "PRIORITY_PREEMPTIVE", 4, 4, None),
    ("base_aging", BASE, "PRIORITY_AGING", 4, 2, None),
    ("same_arrival", [{"name": "pid", "pid": 2, "arrival_time": 0, "burst_time": 2, "priority": 2},
                      {"pid": 1, "arrival_time": 0, "burst_time": 1, "priority": 1}], "FCFS", 4, 4, None),
    ("idle_cpu", [{"pid": 1, "arrival_time": 1000000, "burst_time": 2, "priority": 2}], "RR", 4, 4, None),
    ("rr_arrival_boundary", [{"pid": 1, "arrival_time": 0, "burst_time": 5, "priority": 1},
                             {"pid": 2, "arrival_time": 2, "burst_time": 1, "priority": 2}], "RR", 2, 4, None),
    ("priority_preemption", [{"pid": 1, "arrival_time": 0, "burst_time": 4, "priority": 4},
                             {"pid": 2, "arrival_time": 1, "burst_time": 1, "priority": 1}], "PRIORITY_PREEMPTIVE", 4, 4, None),
]


@pytest.mark.parametrize("name,processes,algorithm,quantum,aging,expected_wait", CASES, ids=[case[0] for case in CASES])
def test_c_matches_oracle(binary, tmp_path, name, processes, algorithm, quantum, aging, expected_wait):
    input_file = tmp_path / f"{name}.json"
    output_file = tmp_path / "cpu_timeline.json"
    input_file.write_text(json.dumps(processes), encoding="utf-8")
    completed = subprocess.run([str(binary), "--algo", algorithm, "--file", str(input_file),
                                "--quantum", str(quantum), "--aging-interval", str(aging),
                                "--output", str(output_file)], capture_output=True, text=True, check=True)
    actual = json.loads(output_file.read_text(encoding="utf-8"))
    assert actual == json.loads(completed.stdout), "stdout y cpu_timeline.json difieren"
    expected = schedule(processes, algorithm, quantum, aging)
    assert actual == expected, f"{name}: desfase en timeline o métricas\nC={actual}\nPython={expected}"
    if expected_wait is not None:
        assert actual["average_waiting_time"] == expected_wait
    if name == "base_rr":
        assert actual["average_response_time"] == 4.0
