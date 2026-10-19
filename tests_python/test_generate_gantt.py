"""Comprueba que el gráfico procede de las métricas y traza, sin valores fijos."""

import json
import subprocess
import sys
from pathlib import Path
from xml.etree import ElementTree

import pytest

from generate_gantt import render_svg
from ground_truth_semana4 import schedule
from test_scheduler import binary


FIXTURE = Path(__file__).resolve().parent / "fixtures/procesos_semana4.json"
SVG = "{http://www.w3.org/2000/svg}"


def blocks(root, kind):
    return [node.attrib for node in root.iter(f"{SVG}rect") if node.get("data-kind") == kind]


def test_fcfs_svg_has_exact_running_and_waiting_blocks():
    processes = json.loads(FIXTURE.read_text(encoding="utf-8"))
    root = ElementTree.fromstring(render_svg(schedule(processes, "FCFS")))
    assert [(int(b["data-pid"]), int(b["data-start"]), int(b["data-end"]))
            for b in blocks(root, "cpu")] == [(1, 0, 5), (2, 5, 8), (3, 8, 16), (4, 16, 22)]
    assert [(int(b["data-pid"]), int(b["data-start"]), int(b["data-end"]))
            for b in blocks(root, "wait")] == [(2, 1, 5), (3, 2, 8), (4, 3, 16)]
    text = " ".join(node.text or "" for node in root.iter())
    assert "5.750 min" in text
    assert "11.250 min" in text


def test_rr_svg_shows_repeated_cpu_bursts_and_waiting():
    processes = json.loads(FIXTURE.read_text(encoding="utf-8"))
    root = ElementTree.fromstring(render_svg(schedule(processes, "RR")))
    assert [(int(b["data-start"]), int(b["data-end"])) for b in blocks(root, "run")
            if b["data-pid"] == "1"] == [(0, 4), (15, 16)]
    assert [(int(b["data-start"]), int(b["data-end"])) for b in blocks(root, "wait")
            if b["data-pid"] == "1"] == [(4, 15)]


@pytest.mark.parametrize("algorithm,expected", [
    ("SJF", [(1, 0, 5), (2, 5, 8), (4, 8, 14), (3, 14, 22)]),
    ("RR", [(1, 0, 4), (2, 4, 7), (3, 7, 11), (4, 11, 15),
            (1, 15, 16), (3, 16, 20), (4, 20, 22)]),
    ("PRIORITY", [(1, 0, 5), (3, 5, 13), (4, 13, 19), (2, 19, 22)]),
])
def test_documented_gantt_segments(algorithm, expected):
    processes = json.loads(FIXTURE.read_text(encoding="utf-8"))
    root = ElementTree.fromstring(render_svg(schedule(processes, algorithm)))
    actual = [(int(b["data-pid"]), int(b["data-start"]), int(b["data-end"]))
              for b in blocks(root, "cpu")]
    assert actual == expected


def test_svg_rejects_overlapping_timeline():
    processes = json.loads(FIXTURE.read_text(encoding="utf-8"))
    result = schedule(processes, "FCFS")
    result["timeline"][1]["start"] = 4
    with pytest.raises(ValueError, match="continuo"):
        render_svg(result)


def test_svg_rejects_idle_cpu_while_process_is_ready():
    result = schedule([{"pid": 1, "arrival_time": 3, "burst_time": 1, "priority": 1}], "FCFS")
    result["timeline"] = [{"pid": 0, "start": 0, "end": 4},
                          {"pid": 1, "start": 4, "end": 5}]
    result["processes"][0].update(first_start=4, completion=5, waiting_time=1,
                                   turnaround_time=2, response_time=1)
    result.update(average_waiting_time=1, average_turnaround_time=2,
                  average_response_time=1)
    with pytest.raises(ValueError, match="inactiva"):
        render_svg(result)


def test_c_two_decimal_average_is_checked_and_displayed_to_three_decimals():
    processes = [{"pid": 1, "arrival_time": 0, "burst_time": 1, "priority": 1},
                 {"pid": 2, "arrival_time": 1, "burst_time": 1, "priority": 1},
                 {"pid": 3, "arrival_time": 1, "burst_time": 1, "priority": 1}]
    result = schedule(processes, "FCFS")
    result["metrics"] = {}
    for key in ("average_waiting_time", "average_turnaround_time", "average_response_time"):
        result[key] = round(result[key], 2)
        result["metrics"][key.replace("average_", "avg_")] = result[key]
    root = ElementTree.fromstring(render_svg(result))
    text = " ".join(node.text or "" for node in root.iter())
    assert "0.333 min" in text
    result["average_waiting_time"] = 0.34
    with pytest.raises(ValueError, match="average_waiting_time"):
        render_svg(result)


@pytest.mark.parametrize("wrapped", [False, True])
def test_cli_accepts_scheduler_or_oracle_json(tmp_path, wrapped):
    processes = json.loads(FIXTURE.read_text(encoding="utf-8"))
    result = schedule(processes, "FCFS")
    source = tmp_path / "cpu_timeline.json"
    source.write_text(json.dumps({"FCFS": result} if wrapped else result), encoding="utf-8")
    output = tmp_path / "gantt"
    script = Path(__file__).resolve().parent / "generate_gantt.py"
    subprocess.run([sys.executable, str(script), "--input", str(source), "--output-dir", str(output)],
                   check=True, capture_output=True, text=True)
    root = ElementTree.parse(output / "gantt_fcfs.svg").getroot()
    assert len(blocks(root, "cpu")) == 4


def test_cli_rejects_mismatched_algorithm_key(tmp_path):
    processes = json.loads(FIXTURE.read_text(encoding="utf-8"))
    source = tmp_path / "cpu_timeline.json"
    source.write_text(json.dumps({"SJF": schedule(processes, "FCFS")}), encoding="utf-8")
    output = tmp_path / "gantt"
    script = Path(__file__).resolve().parent / "generate_gantt.py"
    completed = subprocess.run([sys.executable, str(script), "--input", str(source),
                                "--output-dir", str(output)], capture_output=True, text=True)
    assert completed.returncode == 2
    assert "no coincide" in completed.stderr
    assert not output.exists()


@pytest.mark.parametrize("algorithm,aging,processes", [
    ("FCFS", 0, None),
    ("SJF", 0, None),
    ("RR", 0, None),
    ("PRIORITY", 0, None),
    ("FCFS", 0, [{"pid": 1, "arrival_time": 3, "burst_time": 1, "priority": 1}]),
    ("PRIORITY_AGING", 2, None),
])
def test_cli_renders_actual_c_scheduler_json(binary, tmp_path, algorithm, aging, processes):
    if processes is None:
        processes = json.loads(FIXTURE.read_text(encoding="utf-8"))
    source = tmp_path / "processes.json"
    source.write_text(json.dumps(processes), encoding="utf-8")
    c_json = tmp_path / "cpu_timeline.json"
    c_algorithm = "PRIORITY_PREEMPTIVE" if algorithm == "PRIORITY_AGING" else algorithm
    subprocess.run([str(binary), "--algo", c_algorithm, "--file", str(source),
                    "--aging", str(aging), "--out", str(c_json)],
                   check=True, capture_output=True, text=True)
    output = tmp_path / "gantt"
    script = Path(__file__).resolve().parent / "generate_gantt.py"
    subprocess.run([sys.executable, str(script), "--input", str(c_json),
                    "--output-dir", str(output)], check=True, capture_output=True, text=True)
    root = ElementTree.parse(output / f"gantt_{algorithm.lower()}.svg").getroot()
    actual = [(int(block["data-pid"]), int(block["data-start"]), int(block["data-end"]))
              for block in blocks(root, "cpu")]
    expected = schedule(processes, algorithm, aging_interval=aging or 4)
    assert actual == [(block["pid"], block["start"], block["end"])
                      for block in expected["timeline"]]
