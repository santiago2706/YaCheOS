import pytest

from ground_truth_semana4 import schedule


def test_cpu_input_requires_exact_integers():
    for value in (0, -1, 1.5, "1", True):
        process = {"pid": value, "arrival_time": 0, "burst_time": 1, "priority": 1}
        with pytest.raises(ValueError):
            schedule([process], "FCFS")
