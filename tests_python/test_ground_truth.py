import json
from pathlib import Path

import pytest

from ground_truth_semana4 import schedule


def process(pid, arrival, burst, priority=1):
    return {"pid": pid, "arrival_time": arrival, "burst_time": burst, "priority": priority}


@pytest.mark.parametrize("algorithm,waiting,turnaround,response", [
    ("FCFS", 5.75, 11.25, 5.75),
    ("SJF", 5.25, 10.75, 5.25),
    ("RR", 9.25, 14.75, 4.0),
    ("PRIORITY", 7.75, 13.25, 7.75),
])
def test_semana4_matches_documented_metrics(algorithm, waiting, turnaround, response):
    fixture = Path(__file__).resolve().parent / "fixtures/procesos_semana4.json"
    processes = json.loads(fixture.read_text(encoding="utf-8"))
    result = schedule(processes, algorithm)
    assert result["average_waiting_time"] == waiting
    assert result["average_turnaround_time"] == turnaround
    assert result["average_response_time"] == response


def test_fcfs_ties_use_pid_and_metrics_follow_completion():
    result = schedule([process(2, 0, 2), process(1, 0, 1)], "FCFS")
    assert result["timeline"] == [
        {"pid": 1, "start": 0, "end": 1},
        {"pid": 2, "start": 1, "end": 3},
    ]
    by_pid = {row["pid"]: row for row in result["processes"]}
    assert (by_pid[1]["waiting_time"], by_pid[1]["turnaround_time"], by_pid[1]["response_time"]) == (0, 1, 0)
    assert (by_pid[2]["waiting_time"], by_pid[2]["turnaround_time"], by_pid[2]["response_time"]) == (1, 3, 1)
    assert result["average_waiting_time"] == 0.5


def test_sjf_is_nonpreemptive_with_staggered_arrivals():
    result = schedule([process(1, 0, 4), process(2, 1, 1), process(3, 1, 2)], "SJF")
    assert result["timeline"] == [
        {"pid": 1, "start": 0, "end": 4},
        {"pid": 2, "start": 4, "end": 5},
        {"pid": 3, "start": 5, "end": 7},
    ]
    assert result["average_waiting_time"] == 2.333
    assert result["average_turnaround_time"] == 4.667


def test_rr_enqueues_arrival_before_preempted_process():
    result = schedule([process(1, 0, 5), process(2, 2, 1)], "RR", quantum=2)
    assert result["timeline"] == [
        {"pid": 1, "start": 0, "end": 2},
        {"pid": 2, "start": 2, "end": 3},
        {"pid": 1, "start": 3, "end": 6},
    ]


def test_priority_modes_distinguish_preemption():
    cases = [process(1, 0, 3, 4), process(2, 1, 1, 1)]
    assert schedule(cases, "PRIORITY")["timeline"] == [
        {"pid": 1, "start": 0, "end": 3},
        {"pid": 2, "start": 3, "end": 4},
    ]
    assert schedule(cases, "PRIORITY_PREEMPTIVE")["timeline"] == [
        {"pid": 1, "start": 0, "end": 1},
        {"pid": 2, "start": 1, "end": 2},
        {"pid": 1, "start": 2, "end": 4},
    ]


def test_idle_time_is_recorded():
    result = schedule([process(1, 3, 1)], "FCFS")
    assert result["timeline"] == [
        {"pid": 0, "start": 0, "end": 3},
        {"pid": 1, "start": 3, "end": 4},
    ]
    assert result["average_waiting_time"] == 0


def test_aging_promotes_a_waiting_process():
    cases = [process(2, 0, 4, 1), process(1, 0, 1, 3)]
    assert schedule(cases, "PRIORITY_AGING", aging_interval=1)["timeline"] == [
        {"pid": 2, "start": 0, "end": 2},
        {"pid": 1, "start": 2, "end": 3},
        {"pid": 2, "start": 3, "end": 5},
    ]


def test_aging_resets_partial_wait_after_process_runs():
    cases = [process(1, 1, 4, 5), process(2, 7, 7, 4),
             process(3, 0, 5, 2), process(4, 4, 5, 3)]
    assert schedule(cases, "PRIORITY_AGING", aging_interval=3)["timeline"] == [
        {"pid": 3, "start": 0, "end": 5},
        {"pid": 4, "start": 5, "end": 7},
        {"pid": 1, "start": 7, "end": 10},
        {"pid": 4, "start": 10, "end": 13},
        {"pid": 1, "start": 13, "end": 14},
        {"pid": 2, "start": 14, "end": 21},
    ]


def test_cpu_input_requires_exact_integers():
    for value in (0, -1, 1.5, "1", True):
        process = {"pid": value, "arrival_time": 0, "burst_time": 1, "priority": 1}
        with pytest.raises(ValueError):
            schedule([process], "FCFS")
