"""Cálculos independientes de QA para contrastar la futura salida del motor C."""

import argparse
import json
from pathlib import Path


BASE_CASE = Path(__file__).resolve().parent / "fixtures/memoria_semana6.json"


class SegmentationFault(ValueError):
    """Acceso válido en forma, pero fuera de los límites del segmento."""


def _integer(value, label, minimum=0):
    if type(value) is not int or value < minimum:
        raise ValueError(f"{label} debe ser un entero >= {minimum}")
    return value


def translate_page(virtual_address, page_size, page_table):
    """Separa página/desplazamiento y usa la tabla, sin simular una MMU real."""
    _integer(virtual_address, "Dirección virtual")
    _integer(page_size, "Tamaño de página", 1)
    if not isinstance(page_table, dict):
        raise ValueError("La tabla de páginas debe ser un objeto")
    page, offset = divmod(virtual_address, page_size)
    try:
        frame = page_table[str(page)]
    except KeyError as error:
        raise ValueError(f"Página {page} sin marco asignado") from error
    _integer(frame, "Marco")
    return {"page": page, "offset": offset, "frame": frame,
            "physical_address": frame * page_size + offset}


def translate_segment(segment_id, offset, segments):
    """El límite es longitud: el desplazamiento igual al límite causa trap."""
    _integer(segment_id, "Segmento")
    if type(offset) is not int:
        raise ValueError("Desplazamiento debe ser un entero")
    if not isinstance(segments, list):
        raise ValueError("Los segmentos deben ser una lista")
    matches = [segment for segment in segments if isinstance(segment, dict)
               and type(segment.get("id")) is int and segment["id"] == segment_id]
    if len(matches) != 1:
        raise ValueError(f"Segmento {segment_id} ausente o duplicado")
    segment = matches[0]
    base = _integer(segment.get("base"), "Base")
    limit = _integer(segment.get("limit"), "Límite", 1)
    if not 0 <= offset < limit:
        raise SegmentationFault("EXCEPTION_SEGMENTATION_FAULT")
    return base + offset


def fragmentation(blocks, request_size):
    """Mide huecos contiguos; no altera el mapa recibido."""
    _integer(request_size, "Solicitud", 1)
    if not isinstance(blocks, list) or not blocks:
        raise ValueError("Se requiere un mapa de bloques no vacío")
    total_free = largest_hole = current_hole = 0
    previous_end = None
    for block in blocks:
        if not isinstance(block, dict):
            raise ValueError("Cada bloque debe ser un objeto")
        start = _integer(block.get("start"), "Inicio")
        size = _integer(block.get("size"), "Tamaño", 1)
        if type(block.get("free")) is not bool:
            raise ValueError("free debe ser booleano")
        if previous_end is not None and start != previous_end:
            raise ValueError("El mapa debe ser contiguo y ordenado")
        if block["free"]:
            total_free += size
            current_hole += size
            largest_hole = max(largest_hole, current_hole)
        else:
            current_hole = 0
        previous_end = start + size
    return {"request_size": request_size, "total_free": total_free,
            "largest_hole": largest_hole,
            "can_allocate": largest_hole >= request_size,
            "external_fragmentation": total_free >= request_size > largest_hole}


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--file", type=Path, default=BASE_CASE)
    args = parser.parse_args(argv)
    try:
        case = json.loads(args.file.read_text(encoding="utf-8"))
        page = translate_page(case["virtual_address"], case["page_size_bytes"], case["page_table"])
        segments = {}
        for offset in (300, 900):
            try:
                segments[f"offset_{offset}"] = translate_segment(1, offset, case["segments"])
            except SegmentationFault as error:
                segments[f"offset_{offset}"] = str(error)
        print(json.dumps({"paging": page, "segmentation": segments}, ensure_ascii=False, indent=2))
    except (OSError, ValueError, KeyError, TypeError) as error:
        parser.exit(2, f"Error de auditoría: {error}\n")


if __name__ == "__main__":
    main()
