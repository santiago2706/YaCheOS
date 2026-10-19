"""Casos documentados de MMU y un caso construido de fragmentación."""

import json
import os
import shutil
import subprocess
from pathlib import Path

import pytest

from memory_audit import SegmentationFault, fragmentation, main, translate_page, translate_segment


CASE = json.loads((Path(__file__).resolve().parent / "fixtures/memoria_semana6.json").read_text(encoding="utf-8"))
CORE_DIR = Path(__file__).resolve().parents[1] / "src/core_c"


def compile_c_memory(tmp_path_factory, name, files, defines=()):
    if shutil.which("gcc") is None:
        pytest.fail("No se encontró gcc en PATH; instala GCC o añade su carpeta bin a PATH", pytrace=False)
    sources = [CORE_DIR / file for file in files]
    missing = [str(path) for path in sources + [CORE_DIR / "pcb.h"] if not path.is_file() or not path.stat().st_size]
    if missing:
        pytest.fail("Faltan fuentes C de memoria: " + ", ".join(missing), pytrace=False)
    binary = tmp_path_factory.mktemp(name) / (name + (".exe" if os.name == "nt" else ""))
    completed = subprocess.run(["gcc", "-std=c11", "-Wall", "-Wextra", "-Werror", *defines,
                                *map(str, sources), "-o", str(binary)], capture_output=True, text=True)
    if completed.returncode:
        pytest.fail("No se pudo compilar el motor C de memoria:\n" + completed.stderr, pytrace=False)
    return binary


@pytest.fixture(scope="module")
def memory_translate_binary(tmp_path_factory):
    return compile_c_memory(tmp_path_factory, "memory_translate", ["memory_translate.c"])


@pytest.fixture(scope="module")
def memory_cli_binary(tmp_path_factory):
    return compile_c_memory(tmp_path_factory, "memory_cli",
                            ["memory.c", "memory_contigous.c", "memory_paging.c", "memory_segmentation.c",
                             "scheduler.c", "queue.c", "scheduler_nonpreemptive.c", "scheduler_preemptive.c"],
                            defines=("-DCORE_ENGINE_UNIFIED",))


def test_documented_paging_translation():
    assert translate_page(CASE["virtual_address"], CASE["page_size_bytes"], CASE["page_table"]) == {
        "page": 2, "offset": 1808, "frame": 7, "physical_address": 30480,
    }


def test_shared_memory_fixture_matches_documented_case():
    shared = json.loads((CORE_DIR.parents[1] / "data/memoria_semana6.json").read_text(encoding="utf-8"))
    paging = shared["paginacion"]
    assert paging["tamano_pagina_bytes"] == CASE["page_size_bytes"]
    assert paging["tabla_paginas_proceso_100"] == CASE["page_table"]
    assert paging["caso_base_prueba"]["direccion_virtual"] == CASE["virtual_address"]
    for shared_segment, qa_segment in zip(shared["segmentacion"]["segmentos"], CASE["segments"], strict=True):
        assert (shared_segment["id"], shared_segment["base"], shared_segment["limite"]) == (
            qa_segment["id"], qa_segment["base"], qa_segment["limit"])


def test_documented_segment_and_trap():
    assert translate_segment(1, 300, CASE["segments"]) == 5500
    with pytest.raises(SegmentationFault, match="EXCEPTION_SEGMENTATION_FAULT"):
        translate_segment(1, 900, CASE["segments"])
    with pytest.raises(SegmentationFault, match="EXCEPTION_SEGMENTATION_FAULT"):
        translate_segment(1, 800, CASE["segments"])
    with pytest.raises(SegmentationFault, match="EXCEPTION_SEGMENTATION_FAULT"):
        translate_segment(1, -1, CASE["segments"])


def test_cli_reports_documented_segment_probes(capsys):
    main([])
    segments = json.loads(capsys.readouterr().out)["segmentation"]
    assert segments == {"offset_300": 5500,
                        "offset_900": "EXCEPTION_SEGMENTATION_FAULT"}


def test_invalid_segment_contract_is_not_reported_as_a_trap(tmp_path, capsys):
    case = dict(CASE, segments=[{"id": 1, "base": 5200, "limit": 0}])
    source = tmp_path / "memoria.json"
    source.write_text(json.dumps(case), encoding="utf-8")
    with pytest.raises(SystemExit) as exit_info:
        main(["--file", str(source)])
    assert exit_info.value.code == 2
    assert "Límite" in capsys.readouterr().err


def test_cli_labels_segment_probes_by_offset_not_assumed_result(tmp_path, capsys):
    case = dict(CASE, segments=[{"id": 1, "base": 5200, "limit": 1000}])
    source = tmp_path / "memoria.json"
    source.write_text(json.dumps(case), encoding="utf-8")
    main(["--file", str(source)])
    segments = json.loads(capsys.readouterr().out)["segmentation"]
    assert segments == {"offset_300": 5500, "offset_900": 6100}


def test_external_fragmentation_with_six_free_units_in_small_holes():
    # Escenario de prueba construido según el criterio del plan; no es salida de C.
    blocks = [
        {"start": 0, "size": 2, "free": False},
        {"start": 2, "size": 2, "free": True},
        {"start": 4, "size": 1, "free": False},
        {"start": 5, "size": 2, "free": True},
        {"start": 7, "size": 1, "free": False},
        {"start": 8, "size": 2, "free": True},
    ]
    assert fragmentation(blocks, 3) == {
        "request_size": 3, "total_free": 6, "largest_hole": 2,
        "can_allocate": False, "external_fragmentation": True,
    }
    assert blocks[1]["free"] is True  # El oráculo no modifica el mapa inicial.


def test_adjacent_free_blocks_form_one_hole():
    blocks = [{"start": 0, "size": 2, "free": True},
              {"start": 2, "size": 2, "free": True}]
    assert fragmentation(blocks, 3)["can_allocate"] is True


def test_c_paging_matches_python_oracle(memory_translate_binary, memory_cli_binary):
    expected = translate_page(CASE["virtual_address"], CASE["page_size_bytes"], CASE["page_table"])
    commands = [
        [str(memory_translate_binary), "--mode", "paging", "--virtual", "10000", "--page-size", "4096"],
        [str(memory_cli_binary), "--mode", "paging", "--vaddr", "10000"],
    ]
    for command in commands:
        completed = subprocess.run(command, check=True, capture_output=True, text=True)
        actual = json.loads(completed.stdout)
        assert (actual["page_number"], actual["frame_number"], actual["offset"],
                actual["physical_address"]) == (expected["page"], expected["frame"],
                                                expected["offset"], expected["physical_address"])


def test_c_segmentation_matches_python_oracle(memory_translate_binary, memory_cli_binary):
    expected = translate_segment(1, 300, CASE["segments"])
    valid_commands = [
        [str(memory_translate_binary), "--mode", "segmentation", "--segment", "1", "--offset", "300"],
        [str(memory_cli_binary), "--mode", "segmentation", "--seg", "1", "--offset", "300"],
    ]
    for command in valid_commands:
        actual = json.loads(subprocess.run(command, check=True, capture_output=True, text=True).stdout)
        assert actual["physical_address"] == expected

    for offset in (800, 900):
        with pytest.raises(SegmentationFault):
            translate_segment(1, offset, CASE["segments"])
        standalone = subprocess.run([str(memory_translate_binary), "--mode", "segmentation",
                                     "--segment", "1", "--offset", str(offset)],
                                    capture_output=True, text=True)
        assert standalone.returncode == 2
        assert "EXCEPTION_SEGMENTATION_FAULT" in standalone.stderr
        integrated = subprocess.run([str(memory_cli_binary), "--mode", "segmentation",
                                     "--seg", "1", "--offset", str(offset)],
                                    check=True, capture_output=True, text=True)
        assert json.loads(integrated.stdout)["status"] == "EXCEPTION_SEGMENTATION_FAULT"


def test_c_fragmentation_matches_python_oracle(memory_cli_binary):
    completed = subprocess.run([str(memory_cli_binary), "--mode", "fragmentation"],
                               check=True, capture_output=True, text=True)
    actual = json.loads(completed.stdout)
    expected = fragmentation(actual["blocks"], 3)
    assert actual["total_size_kb"] == sum(block["size"] for block in actual["blocks"]) == 10
    assert actual["total_free_kb"] == expected["total_free"] == 6
    assert expected["largest_hole"] == 2
    assert actual["external_fragmentation"] is expected["external_fragmentation"] is True
    assert expected["can_allocate"] is False
    assert "Solicitud de 3 unidades" in actual["alert"]
