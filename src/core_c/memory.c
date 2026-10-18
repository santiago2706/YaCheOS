#include "pcb.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void run_paging_demo(int virtual_addr) {
    /* Tabla del Proceso #100 según contrato Semana 6: P0->M5, P1->M2, P2->M7, P3->M0 */
    PageTableEntry table[4] = {
        {0, 5, true},
        {1, 2, true},
        {2, 7, true},
        {3, 0, true}
    };
    int page = 0, frame = -1, offset = 0;
    int physical = translate_virtual_to_physical(table, 4, virtual_addr, &page, &frame, &offset);

    printf("{\n");
    printf("  \"pid\": 100,\n");
    printf("  \"page_size_bytes\": %d,\n", PAGE_SIZE_BYTES);
    printf("  \"virtual_address\": %d,\n", virtual_addr);
    printf("  \"page_number\": %d,\n", page);
    printf("  \"frame_number\": %d,\n", frame);
    printf("  \"offset\": %d,\n", offset);
    if (physical >= 0) {
        printf("  \"physical_address\": %d,\n", physical);
        printf("  \"status\": \"OK\"\n");
    } else {
        printf("  \"physical_address\": null,\n");
        printf("  \"status\": \"PAGE_FAULT\"\n");
    }
    printf("}\n");
}

static void run_segmentation_demo(int seg_id, int offset) {
    /* Tabla de Segmentos según contrato Semana 6 */
    SegmentEntry seg_table[3] = {
        {0, "Codigo", 1000, 400},
        {1, "Datos",  5200, 800},
        {2, "Pila",   9000, 300}
    };
    int phys = -1;
    const char *err = NULL;
    int rc = translate_segment_address(seg_table, 3, seg_id, offset, &phys, &err);

    printf("{\n");
    printf("  \"segment\": %d,\n", seg_id);
    printf("  \"offset\": %d,\n", offset);
    if (rc == 0) {
        printf("  \"physical_address\": %d,\n", phys);
        printf("  \"status\": \"OK\"\n");
    } else {
        printf("  \"physical_address\": null,\n");
        printf("  \"status\": \"%s\"\n", err ? err : "EXCEPTION_SEGMENTATION_FAULT");
    }
    printf("}\n");
}

static void run_fragmentation_case(void) {
    /* Caso DoD T2.4: 6 unidades libres distribuidas en huecos menores de 3, solicitud de 3 unidades */
    ContiguousMemory mem;
    mem_init(&mem, 10);
    alloc_first_fit(&mem, 1, 2); /* [0..2) libre luego */
    alloc_first_fit(&mem, 2, 2); /* [2..4) ocupado */
    alloc_first_fit(&mem, 3, 2); /* [4..6) libre luego */
    alloc_first_fit(&mem, 4, 2); /* [6..8) ocupado */
    alloc_first_fit(&mem, 5, 2); /* [8..10) libre luego */

    /* Liberar PIDs 1, 3 y 5: quedan 3 huecos de tamaño 2 (total libre = 6 unidades, ningún hueco >= 3) */
    free_memory(&mem, 1);
    free_memory(&mem, 3);
    free_memory(&mem, 5);

    /* Solicitar 3 unidades contiguas -> Rechaza y emite alerta de Fragmentación Externa */
    alloc_first_fit(&mem, 99, 3);
    print_contiguous_json(&mem);
}

static void run_contiguous_Processes_case(const char *fit_mode) {
    ContiguousMemory mem;
    mem_init(&mem, 512);
    if (strcmp(fit_mode, "BEST") == 0) {
        alloc_best_fit(&mem, 1, 128);
        alloc_best_fit(&mem, 2, 64);
        alloc_best_fit(&mem, 3, 256);
        free_memory(&mem, 2);
        alloc_best_fit(&mem, 4, 64);
    } else {
        alloc_first_fit(&mem, 1, 128);
        alloc_first_fit(&mem, 2, 64);
        alloc_first_fit(&mem, 3, 256);
        free_memory(&mem, 2);
        alloc_first_fit(&mem, 4, 64);
    }
    print_contiguous_json(&mem);
}

int main(int argc, char *argv[]) {
    const char *mode = "paging";
    const char *fit = "FIRST";
    int vaddr = 10000;
    int seg_id = 1;
    int offset = 300;

    for (int i = 1; i < argc; i++) {
        if (strcmp(argv[i], "--mode") == 0 && i + 1 < argc) {
            mode = argv[++i];
        } else if (strcmp(argv[i], "--vaddr") == 0 && i + 1 < argc) {
            vaddr = atoi(argv[++i]);
        } else if (strcmp(argv[i], "--seg") == 0 && i + 1 < argc) {
            seg_id = atoi(argv[++i]);
        } else if (strcmp(argv[i], "--offset") == 0 && i + 1 < argc) {
            offset = atoi(argv[++i]);
        } else if (strcmp(argv[i], "--fit") == 0 && i + 1 < argc) {
            fit = argv[++i];
        } else if (strcmp(argv[i], "--algo") == 0) {
            /* Si se invoca core_engine como planificador de CPU */
            const char *algo = "FCFS";
            const char *filepath = "data/procesos_semana4.json";
            const char *outpath = "cpu_timeline.json";
            int quantum = 4;
            int aging = 0;
            for (int j = 1; j < argc; j++) {
                if (strcmp(argv[j], "--algo") == 0 && j + 1 < argc) algo = argv[++j];
                else if (strcmp(argv[j], "--file") == 0 && j + 1 < argc) filepath = argv[++j];
                else if (strcmp(argv[j], "--quantum") == 0 && j + 1 < argc) quantum = atoi(argv[++j]);
                else if (strcmp(argv[j], "--aging") == 0 && j + 1 < argc) aging = atoi(argv[++j]);
                else if (strcmp(argv[j], "--out") == 0 && j + 1 < argc) outpath = argv[++j];
            }
            Process procs[MAX_PROCESSES];
            int n = load_processes_from_json(filepath, procs, MAX_PROCESSES);
            if (n <= 0) return EXIT_FAILURE;
            ScheduleResult res;
            if (strcmp(algo, "FCFS") == 0) schedule_fcfs(procs, n, &res);
            else if (strcmp(algo, "SJF") == 0) schedule_sjf(procs, n, &res);
            else if (strcmp(algo, "RR") == 0) schedule_rr(procs, n, quantum, &res);
            else if (strcmp(algo, "PRIORITY") == 0) schedule_priority_nonpreemptive(procs, n, &res);
            else if (strcmp(algo, "PRIORITY_PREEMPTIVE") == 0) schedule_priority_preemptive(procs, n, aging, &res);
            else return EXIT_FAILURE;
            print_schedule_json(&res, outpath);
            return EXIT_SUCCESS;
        }
    }

    if (strcmp(mode, "paging") == 0) {
        run_paging_demo(vaddr);
    } else if (strcmp(mode, "segmentation") == 0) {
        run_segmentation_demo(seg_id, offset);
    } else if (strcmp(mode, "fragmentation") == 0) {
        run_fragmentation_case();
    } else if (strcmp(mode, "contiguous") == 0) {
        run_contiguous_Processes_case(fit);
    } else {
        fprintf(stderr, "{\"error\": \"Modo de memoria desconocido: %s\"}\n", mode);
        return EXIT_FAILURE;
    }

    return EXIT_SUCCESS;
}