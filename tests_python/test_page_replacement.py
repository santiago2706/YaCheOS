from page_replacement import simulate
import pytest


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
