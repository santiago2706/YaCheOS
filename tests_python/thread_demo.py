"""Dos hilos reales comparten PID; cada hilo tiene su propio identificador.

El PCB del simulador C representa un proceso modelado, no uno de estos hilos.
"""
import json
import os
import threading


def main():
    records = []
    lock = threading.Lock()
    barrier = threading.Barrier(2)

    def worker(number):
        barrier.wait()
        with lock:
            records.append({"worker": number, "pid": os.getpid(),
                            "thread_ident": threading.get_ident()})

    threads = [threading.Thread(target=worker, args=(i,), name=f"worker-{i}") for i in (1, 2)]
    for thread in threads:
        thread.start()
    for thread in threads:
        thread.join()
    print(json.dumps({"threads": sorted(records, key=lambda x: x["worker"]),
                      "same_process": len(records) == 2 and len({record["pid"] for record in records}) == 1,
                      "distinct_thread_ids": len(records) == 2 and len({record["thread_ident"] for record in records}) == 2,
                      "all_terminated": len(records) == 2 and all(not t.is_alive() for t in threads)}))


if __name__ == "__main__":
    main()
