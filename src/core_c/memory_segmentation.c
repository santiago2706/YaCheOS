#include "pcb.h"
#include <stddef.h>

int translate_segment_address(const SegmentEntry *table, int table_size,
                              int seg_id, int offset, int *out_physical,
                              const char **out_error) {
    if (!table || offset < 0) {
        if (out_error) *out_error = "EXCEPTION_SEGMENTATION_FAULT";
        return -1;
    }

    for (int i = 0; i < table_size; i++) {
        if (table[i].seg_id == seg_id) {
            if (offset >= table[i].limit) {
                if (out_error) *out_error = "EXCEPTION_SEGMENTATION_FAULT";
                return -1;
            }
            int phys = table[i].base + offset;
            if (out_physical) *out_physical = phys;
            if (out_error) *out_error = "OK";
            return 0;
        }
    }

    if (out_error) *out_error = "EXCEPTION_INVALID_SEGMENT";
    return -1;
}