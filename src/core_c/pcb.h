#ifndef YACHEOS_PCB_H
#define YACHEOS_PCB_H

#include <stdbool.h>
#include <stddef.h>

#define MAX_NAME_LEN 32
#define MAX_PROCESSES 128
#define INITIAL_TIMELINE_CAPACITY 512
#define PAGE_SIZE_BYTES 4096

typedef enum {
    STATE_NEW = 0,
    STATE_READY,
    STATE_RUNNING,
    STATE_BLOCKED,
    STATE_TERMINATED
} ProcessState;

typedef struct Process {
    int pid;
    char name[MAX_NAME_LEN];
    int arrival_time;
    int burst_time;
    int remaining_time;
    int priority;
    int effective_priority;
    int memory_kb;
    ProcessState state;

    int start_time;
    int completion_time;
    int waiting_time;
    int turnaround_time;
    int response_time;
    int wait_ticks_since_boost;
} Process;

typedef struct QueueNode {
    Process *process;
    struct QueueNode *next;
} QueueNode;

typedef struct Queue {
    QueueNode *head;
    QueueNode *tail;
    int size;
} Queue;

typedef struct TimelineSegment {
    int pid;
    char name[MAX_NAME_LEN];
    int start_time;
    int end_time;
} TimelineSegment;

typedef struct ScheduleResult {
    char algorithm[32];
    int quantum;
    int aging_interval;
    Process processes[MAX_PROCESSES];
    int process_count;
    TimelineSegment *segments;
    int segment_count;
    int segment_capacity;
    bool timeline_failed;
    double avg_waiting_time;
    double avg_turnaround_time;
    double avg_response_time;
} ScheduleResult;

/* Operaciones de Cola (queue.c) */
void queue_init(Queue *q);
bool queue_is_empty(const Queue *q);
void queue_enqueue(Queue *q, Process *p);
Process *queue_dequeue(Queue *q);
void queue_clear(Queue *q);

/* Planificadores No Apropiativos (scheduler_nonpreemptive.c) */
void schedule_fcfs(const Process *input, int n, ScheduleResult *out);
void schedule_sjf(const Process *input, int n, ScheduleResult *out);
void schedule_priority_nonpreemptive(const Process *input, int n, ScheduleResult *out);

/* Planificadores Apropiativos (scheduler_preemptive.c) */
void schedule_rr(const Process *input, int n, int quantum, ScheduleResult *out);
void schedule_priority_preemptive(const Process *input, int n, int aging_interval, ScheduleResult *out);

/* Utilidades de Planificación */
int load_processes_from_json(const char *filepath, Process *out_procs, int max_procs);
TimelineSegment *append_timeline_segment(ScheduleResult *res);
void free_schedule_result(ScheduleResult *res);
void print_schedule_json(const ScheduleResult *res, const char *out_filepath);

/* --- Módulo 3: Gestión de Memoria --- */
#define MAX_MEM_BLOCKS 64

typedef struct MemBlock {
    int start;
    int size;
    int pid;
    bool free;
} MemBlock;

typedef struct ContiguousMemory {
    int total_size_kb;
    MemBlock blocks[MAX_MEM_BLOCKS];
    int block_count;
    bool external_fragmentation_alert;
    int last_failed_req_size;
    int total_free_kb;
} ContiguousMemory;

void mem_init(ContiguousMemory *mem, int total_size_kb);
bool alloc_first_fit(ContiguousMemory *mem, int pid, int size_kb);
bool alloc_best_fit(ContiguousMemory *mem, int pid, int size_kb);
bool free_memory(ContiguousMemory *mem, int pid);
void print_contiguous_json(const ContiguousMemory *mem);

/* Paginación (memory_paging.c) */
typedef struct PageTableEntry {
    int page_number;
    int frame_number;
    bool valid;
} PageTableEntry;

int translate_virtual_to_physical(const PageTableEntry *table, int table_size,
                                  int virtual_address, int *out_page,
                                  int *out_frame, int *out_offset);

/* Segmentación (memory_segmentation.c) */
typedef struct SegmentEntry {
    int seg_id;
    const char *name;
    int base;
    int limit;
} SegmentEntry;

int translate_segment_address(const SegmentEntry *table, int table_size,
                              int seg_id, int offset, int *out_physical,
                              const char **out_error);

#endif /* YACHEOS_PCB_H */
