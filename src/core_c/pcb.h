#ifndef PCB_H
#define PCB_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>

typedef enum {
    STATE_NEW,
    STATE_READY,
    STATE_RUNNING,
    STATE_BLOCKED,
    STATE_TERMINATED
} ProcessState;

typedef struct {
    int pid;
    char name[32];
    int arrival_time;
    int burst_time;
    int remaining_time;
    int priority;
    int memory_kb;
    ProcessState state;
    int start_time;
    int finish_time;
    int waiting_time;
    int turnaround_time;
    int response_time;
} Process;

typedef struct {
    int pid;
    char name[32];
    int start;
    int end;
    int duration;
} TimelineSegment;

#endif // PCB_H
