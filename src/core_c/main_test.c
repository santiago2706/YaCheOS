#include "pcb.h"
#include <assert.h>
#include <stdio.h>

int main(void) {
    Queue queue;
    Process processes[5] = {{0}};
    queue_init(&queue);
    for (int i = 0; i < 5; i++) {
        processes[i].pid = i + 1;
        queue_enqueue(&queue, &processes[i]);
        assert(queue.size == i + 1);
    }
    for (int tick = 0; tick < 5; tick++) {
        Process *next = queue_dequeue(&queue);
        assert(next != NULL && next->pid == tick + 1);
        assert(queue.size == 4 - tick);
    }
    assert(queue_is_empty(&queue));
    queue_clear(&queue);

    ContiguousMemory memory;
    mem_init(&memory, 10);
    assert(alloc_first_fit(&memory, 1, 2));
    assert(alloc_first_fit(&memory, 2, 2));
    assert(alloc_first_fit(&memory, 3, 2));
    assert(free_memory(&memory, 2));
    assert(memory.blocks[1].free && memory.blocks[1].start == 2);
    assert(alloc_best_fit(&memory, 4, 2));
    assert(memory.blocks[1].pid == 4);
    assert(free_memory(&memory, 1));
    assert(free_memory(&memory, 4));
    assert(free_memory(&memory, 3));
    assert(memory.block_count == 1 && memory.total_free_kb == 10);
    puts("queue and contiguous memory: OK");
    return 0;
}
