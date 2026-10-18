#include "pcb.h"

int translate_virtual_to_physical(const PageTableEntry *table, int table_size,
                                  int virtual_address, int *out_page,
                                  int *out_frame, int *out_offset) {
    if (!table || virtual_address < 0) return -1;

    int page = virtual_address / PAGE_SIZE_BYTES;
    int offset = virtual_address % PAGE_SIZE_BYTES;

    if (out_page) *out_page = page;
    if (out_offset) *out_offset = offset;

    for (int i = 0; i < table_size; i++) {
        if (table[i].page_number == page && table[i].valid) {
            int frame = table[i].frame_number;
            if (out_frame) *out_frame = frame;
            return (frame * PAGE_SIZE_BYTES) + offset;
        }
    }
    return -1; /* Page Fault / Página inválida */
}