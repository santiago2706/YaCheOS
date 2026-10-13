"""Casos documentados de MMU y un caso construido de fragmentación."""

import json
from pathlib import Path

import pytest

from memory_audit import SegmentationFault, fragmentation, main, translate_page, translate_segment


CASE = json.loads((Path(__file__).resolve().parent / "fixtures/memoria_semana6.json").read_text(encoding="utf-8"))


def test_documented_paging_translation():
    assert translate_page(CASE["virtual_address"], CASE["page_size_bytes"], CASE["page_table"]) == {
        "page": 2, "offset": 1808, "frame": 7, "physical_address": 30480,
    }


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
