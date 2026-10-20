"""FIFO, LRU y Óptimo con traza por referencia y desempate determinista."""
import argparse
import json
import sys
from collections import defaultdict, deque


def simulate(references, frame_count, algorithm):
    if algorithm not in ("FIFO", "LRU", "OPTIMAL") or not isinstance(frame_count, int) or isinstance(frame_count, bool) or frame_count < 1:
        raise ValueError("Algoritmo o cantidad de marcos inválida")
    if not isinstance(references, list) or any(not isinstance(p, int) or isinstance(p, bool) or p < 0 for p in references):
        raise ValueError("Las referencias deben ser una lista de páginas enteras no negativas")
    frames, loaded, last_used, trace = [], {}, {}, []
    future = defaultdict(deque)
    if algorithm == "OPTIMAL":
        for position, page in enumerate(references):
            future[page].append(position)
    hits = faults = 0
    for index, page in enumerate(references):
        if algorithm == "OPTIMAL":
            future[page].popleft()
        hit = page in frames
        victim = None
        if hit:
            hits += 1
        else:
            faults += 1
            if len(frames) < frame_count:
                frames.append(page)
            else:
                if algorithm == "FIFO":
                    victim = min(frames, key=lambda p: loaded[p])
                elif algorithm == "LRU":
                    victim = min(frames, key=lambda p: last_used[p])
                else:
                    def next_use(p):
                        return future[p][0] if future[p] else float("inf")
                    victim = max(frames, key=next_use)
                frames[frames.index(victim)] = page
            loaded[page] = index
        last_used[page] = index
        trace.append({"index": index, "page": page, "frames": frames.copy(), "hit": hit, "fault": not hit, "evicted": victim})
    return {"algorithm": algorithm, "frame_count": frame_count, "hits": hits, "faults": faults, "trace": trace}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--file", help="JSON con references y frames; por defecto stdin")
    args = parser.parse_args()
    try:
        with open(args.file, encoding="utf-8") if args.file else sys.stdin as source:
            request = json.load(source)
        result = {algo: simulate(request["references"], request["frames"], algo)
                  for algo in ("FIFO", "LRU", "OPTIMAL")}
        print(json.dumps(result, ensure_ascii=False))
    except (OSError, ValueError, KeyError, TypeError) as error:
        print(json.dumps({"error": str(error)}), file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
