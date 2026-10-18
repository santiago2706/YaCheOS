#include "pcb.h"
#include <stdio.h>
#include <string.h>

void mem_init(ContiguousMemory *mem, int total_size_kb) {
    memset(mem, 0, sizeof(ContiguousMemory));
    mem->total_size_kb = total_size_kb;
    mem->block_count = 1;
    mem->blocks[0].start = 0;
    mem->blocks[0].size = total_size_kb;
    mem->blocks[0].pid = -1;
    mem->blocks[0].free = true;
    mem->total_free_kb = total_size_kb;
}

static void recalc_free_space(ContiguousMemory *mem) {
    int total = 0;
    for (int i = 0; i < mem->block_count; i++) {
        if (mem->blocks[i].free) {
            total += mem->blocks[i].size;
        }
    }
    mem->total_free_kb = total;
}

static bool allocate_at_index(ContiguousMemory *mem, int idx, int pid, int size_kb) {
    MemBlock *blk = &mem->blocks[idx];
    if (blk->size == size_kb) {
        blk->free = false;
        blk->pid = pid;
    } else {
        if (mem->block_count >= MAX_MEM_BLOCKS) return false;
        for (int i = mem->block_count; i > idx + 1; i--) {
            mem->blocks[i] = mem->blocks[i - 1];
        }
        mem->blocks[idx + 1].start = blk->start + size_kb;
        mem->blocks[idx + 1].size = blk->size - size_kb;
        mem->blocks[idx + 1].pid = -1;
        mem->blocks[idx + 1].free = true;

        blk->size = size_kb;
        blk->pid = pid;
        blk->free = false;
        mem->block_count++;
    }
    mem->external_fragmentation_alert = false;
    mem->last_failed_req_size = 0;
    recalc_free_space(mem);
    return true;
}

bool alloc_first_fit(ContiguousMemory *mem, int pid, int size_kb) {
    if (!mem || size_kb <= 0) return false;
    for (int i = 0; i < mem->block_count; i++) {
        if (mem->blocks[i].free && mem->blocks[i].size >= size_kb) {
            return allocate_at_index(mem, i, pid, size_kb);
        }
    }
    recalc_free_space(mem);
    if (mem->total_free_kb >= size_kb) {
        mem->external_fragmentation_alert = true;
        mem->last_failed_req_size = size_kb;
    }
    return false;
}

bool alloc_best_fit(ContiguousMemory *mem, int pid, int size_kb) {
    if (!mem || size_kb <= 0) return false;
    int best_idx = -1;
    for (int i = 0; i < mem->block_count; i++) {
        if (mem->blocks[i].free && mem->blocks[i].size >= size_kb) {
            if (best_idx == -1 || mem->blocks[i].size < mem->blocks[best_idx].size) {
                best_idx = i;
            }
        }
    }
    if (best_idx != -1) {
        return allocate_at_index(mem, best_idx, pid, size_kb);
    }
    recalc_free_space(mem);
    if (mem->total_free_kb >= size_kb) {
        mem->external_fragmentation_alert = true;
        mem->last_failed_req_size = size_kb;
    }
    return false;
}

bool free_memory(ContiguousMemory *mem, int pid) {
    if (!mem) return false;
    bool freed_any = false;
    for (int i = 0; i < mem->block_count; i++) {
        if (!mem->blocks[i].free && mem->blocks[i].pid == pid) {
            mem->blocks[i].free = true;
            mem->blocks[i].pid = -1;
            freed_any = true;
        }
    }
    /* Fusionar huecos libres contiguos */
    int i = 0;
    while (i < mem->block_count - 1) {
        if (mem->blocks[i].free && mem->blocks[i + 1].free) {
            mem->blocks[i].size += mem->blocks[i + 1].size;
            for (int j = i + 1; j < mem->block_count - 1; j++) {
                mem->blocks[j] = mem->blocks[j + 1];
            }
            mem->block_count--;
        } else {
            i++;
        }
    }
    recalc_free_space(mem);
    return freed_any;
}

void print_contiguous_json(const ContiguousMemory *mem) {
    printf("{\n");
    printf("  \"total_size_kb\": %d,\n", mem->total_size_kb);
    printf("  \"total_free_kb\": %d,\n", mem->total_free_kb);
    printf("  \"external_fragmentation\": %s,\n", mem->external_fragmentation_alert ? "true" : "false");
    if (mem->external_fragmentation_alert) {
        printf("  \"alert\": \"EXTERNAL_FRAGMENTATION: Solicitud de %d unidades rechazada pese a tener %d unidades libres totales en huecos no contiguos.\",\n",
               mem->last_failed_req_size, mem->total_free_kb);
    }
    printf("  \"blocks\": [\n");
    for (int i = 0; i < mem->block_count; i++) {
        const MemBlock *b = &mem->blocks[i];
        printf("    {\"start\": %d, \"size\": %d, \"pid\": %d, \"free\": %s}%s\n",
               b->start, b->size, b->pid, b->free ? "true" : "false",
               (i + 1 < mem->block_count) ? "," : "");
    }
    printf("  ]\n");
    printf("}\n");
}