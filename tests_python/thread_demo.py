"""Dos hilos reales de Python; sus identificadores no son PIDs simulados."""
import json
import threading


def main():
    records = []
    lock = threading.Lock()
    barrier = threading.Barrier(2)

    def worker(number):
        barrier.wait()
        with lock:
            records.append({"worker": number, "thread_ident": threading.get_ident()})

    threads = [threading.Thread(target=worker, args=(i,), name=f"worker-{i}") for i in (1, 2)]
    for thread in threads:
        thread.start()
    for thread in threads:
        thread.join()
    print(json.dumps({"threads": sorted(records, key=lambda x: x["worker"]),
                      "all_terminated": all(not t.is_alive() for t in threads)}))


if __name__ == "__main__":
    main()
