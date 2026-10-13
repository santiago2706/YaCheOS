import json
import subprocess
import sys
from pathlib import Path

import pytest

from page_replacement import simulate


def test_belady_and_optimal():
    refs = [1, 2, 3, 4, 1, 2, 5, 1, 2, 3, 4, 5]
    assert simulate(refs, 3, "FIFO")["faults"] == 9
    assert simulate(refs, 4, "FIFO")["faults"] == 10
    for count in (3, 4):
        optimal = simulate(refs, count, "OPTIMAL")["faults"]
        assert optimal <= simulate(refs, count, "FIFO")["faults"]
        assert optimal <= simulate(refs, count, "LRU")["faults"]


def test_invalid_frame_count():
    with pytest.raises(ValueError):
        simulate([1, 2], True, "FIFO")


def test_trace_accounts_for_every_reference():
    refs = [1, 2, 3, 4, 1, 2, 5, 1, 2, 3, 4, 5]
    for algorithm in ("FIFO", "LRU", "OPTIMAL"):
        result = simulate(refs, 3, algorithm)
        assert len(result["trace"]) == len(refs)
        assert result["hits"] + result["faults"] == len(refs)
        assert all(len(step["frames"]) <= 3 for step in result["trace"])
        assert all(step["hit"] != step["fault"] for step in result["trace"])
    assert simulate(refs, 3, "OPTIMAL")["trace"][3]["evicted"] == 3


@pytest.mark.parametrize("frames,expected_faults", [
    (3, {"FIFO": 9, "LRU": 10, "OPTIMAL": 7}),
    (4, {"FIFO": 10, "LRU": 8, "OPTIMAL": 6}),
])
def test_documented_comparison_table(frames, expected_faults):
    refs = [1, 2, 3, 4, 1, 2, 5, 1, 2, 3, 4, 5]
    for algorithm, faults in expected_faults.items():
        result = simulate(refs, frames, algorithm)
        assert result["faults"] == faults
        assert result["hits"] == len(refs) - faults


@pytest.mark.parametrize("frames,expected", [(3, (9, 10, 7)), (4, (10, 8, 6))])
def test_cli_emits_json_contract_for_go(frames, expected):
    script = Path(__file__).resolve().parent / "page_replacement.py"
    case = Path(__file__).resolve().parent / "fixtures" / f"referencias_belady_{frames}marcos.json"
    completed = subprocess.run([sys.executable, str(script), "--file", str(case)],
                               check=True, capture_output=True, text=True)
    result = json.loads(completed.stdout)
    assert tuple(result[name]["faults"] for name in ("FIFO", "LRU", "OPTIMAL")) == expected
    assert all(len(result[name]["trace"]) == 12 for name in result)


def test_cli_reports_invalid_request_without_partial_result(tmp_path):
    script = Path(__file__).resolve().parent / "page_replacement.py"
    case = tmp_path / "invalid.json"
    case.write_text('{"references": [1, 2], "frames": 0}', encoding="utf-8")
    completed = subprocess.run([sys.executable, str(script), "--file", str(case)],
                               capture_output=True, text=True)
    assert completed.returncode == 2
    assert not completed.stdout
    assert "error" in json.loads(completed.stderr)


def test_cli_accepts_request_on_stdin():
    script = Path(__file__).resolve().parent / "page_replacement.py"
    request = {"references": [1, 2, 1], "frames": 2}
    completed = subprocess.run([sys.executable, str(script)], input=json.dumps(request),
                               check=True, capture_output=True, text=True)
    result = json.loads(completed.stdout)
    assert {name: (entry["hits"], entry["faults"]) for name, entry in result.items()} == {
        "FIFO": (1, 2), "LRU": (1, 2), "OPTIMAL": (1, 2),
    }
    assert not completed.stderr
