#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>

typedef struct {
    int id;
    char name[32];
    int base;
    int limit;
} Segment;

static int default_page_table[16] = {5, 2, 7, 0, 1, 3, 4, 6};

static Segment default_segments[3] = {
    {0, "Codigo", 1000, 400},
    {1, "Datos", 5200, 800},
    {2, "Pila", 9000, 300}
};

int main(int argc, char *argv[]) {
    char mode[32] = "paging";
    int virtual_addr = 10000;
    int page_size = 4096;
    int seg_id = -1;
    int seg_offset = -1;

    for (int i = 1; i < argc; i++) {
        if (strcmp(argv[i], "--mode") == 0 && i + 1 < argc) {
            strncpy(mode, argv[++i], sizeof(mode) - 1);
        } else if (strcmp(argv[i], "--virtual") == 0 && i + 1 < argc) {
            virtual_addr = atoi(argv[++i]);
        } else if (strcmp(argv[i], "--page-size") == 0 && i + 1 < argc) {
            page_size = atoi(argv[++i]);
        } else if (strcmp(argv[i], "--segment") == 0 && i + 1 < argc) {
            seg_id = atoi(argv[++i]);
            strncpy(mode, "segmentation", sizeof(mode) - 1);
        } else if (strcmp(argv[i], "--offset") == 0 && i + 1 < argc) {
            seg_offset = atoi(argv[++i]);
        }
    }

    if (strcmp(mode, "segmentation") == 0 || seg_id >= 0) {
        if (seg_id < 0 || seg_id > 2) {
            fprintf(stderr, "Error: Segmento invalido %d (Segmentos disponibles: 0=Codigo, 1=Datos, 2=Pila)\n", seg_id);
            return 1;
        }
        if (seg_offset < 0) {
            fprintf(stderr, "Error: Desplazamiento invalido %d\n", seg_offset);
            return 1;
        }

        Segment seg = default_segments[seg_id];
        if (seg_offset >= seg.limit) {
            fprintf(stderr, "EXCEPTION_SEGMENTATION_FAULT: Desplazamiento %d supera el limite %d del segmento %d (%s)\n",
                    seg_offset, seg.limit, seg_id, seg.name);
            return 2;
        }

        int physical = seg.base + seg_offset;
        printf("{\n");
        printf("  \"mode\": \"segmentation\",\n");
        printf("  \"segment_id\": %d,\n", seg_id);
        printf("  \"segment_name\": \"%s\",\n", seg.name);
        printf("  \"base\": %d,\n", seg.base);
        printf("  \"limit\": %d,\n", seg.limit);
        printf("  \"offset\": %d,\n", seg_offset);
        printf("  \"physical_address\": %d,\n", physical);
        printf("  \"valid\": true\n");
        printf("}\n");
        return 0;
    }

    // Paging translation mode
    if (page_size <= 0) page_size = 4096;
    if (virtual_addr < 0) {
        fprintf(stderr, "Error: Direccion virtual invalida %d\n", virtual_addr);
        return 1;
    }

    int page = virtual_addr / page_size;
    int offset = virtual_addr % page_size;

    if (page < 0 || page >= 8) {
        fprintf(stderr, "Error: Fallo de pagina (Page Fault). Pagina %d fuera de tabla de paginas\n", page);
        return 3;
    }

    int frame = default_page_table[page];
    int physical = (frame * page_size) + offset;

    printf("{\n");
    printf("  \"mode\": \"paging\",\n");
    printf("  \"virtual_address\": %d,\n", virtual_addr);
    printf("  \"page_size\": %d,\n", page_size);
    printf("  \"page_number\": %d,\n", page);
    printf("  \"frame_number\": %d,\n", frame);
    printf("  \"offset\": %d,\n", offset);
    printf("  \"physical_address\": %d,\n", physical);
    printf("  \"valid\": true\n");
    printf("}\n");

    return 0;
}
