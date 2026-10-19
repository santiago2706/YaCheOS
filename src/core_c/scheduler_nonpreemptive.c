#include "pcb.h"
#include <string.h>

static void init_result(const char *algo, const Process *input, int n, ScheduleResult *out) {
    memset(out, 0, sizeof(ScheduleResult));
    strncpy(out->algorithm, algo, sizeof(out->algorithm) - 1);
    out->process_count = n;
    for (int i = 0; i < n; i++) {
        out->processes[i] = input[i];
        out->processes[i].remaining_time = input[i].burst_time;
        out->processes[i].effective_priority = input[i].priority;
        out->processes[i].state = STATE_NEW;
        out->processes[i].start_time = -1;
        out->processes[i].completion_time = 0;
        out->processes[i].waiting_time = 0;
        out->processes[i].turnaround_time = 0;
        out->processes[i].response_time = 0;
        out->processes[i].wait_ticks_since_boost = 0;
    }
}

static void compute_averages(ScheduleResult *out) {
    double total_wait = 0.0, total_turn = 0.0, total_resp = 0.0;
    for (int i = 0; i < out->process_count; i++) {
        Process *p = &out->processes[i];
        p->turnaround_time = p->completion_time - p->arrival_time;
        p->waiting_time = p->turnaround_time - p->burst_time;
        p->response_time = p->start_time - p->arrival_time;
        total_wait += p->waiting_time;
        total_turn += p->turnaround_time;
        total_resp += p->response_time;
    }
    if (out->process_count > 0) {
        out->avg_waiting_time = total_wait / out->process_count;
        out->avg_turnaround_time = total_turn / out->process_count;
        out->avg_response_time = total_resp / out->process_count;
    }
}

static void record_segment(ScheduleResult *out, const Process *p, int start, int end) {
    if (start >= end || out->segment_count >= MAX_TIMELINE_SEGMENTS) return;
    TimelineSegment *seg = &out->segments[out->segment_count++];
    seg->pid = p->pid;
    strncpy(seg->name, p->name, MAX_NAME_LEN - 1);
    seg->start_time = start;
    seg->end_time = end;
}

void schedule_fcfs(const Process *input, int n, ScheduleResult *out) {
    init_result("FCFS", input, n, out);
    int completed = 0;
    int current_time = 0;

    while (completed < n) {
        int idx = -1;
        for (int i = 0; i < n; i++) {
            Process *p = &out->processes[i];
            if (p->state != STATE_TERMINATED && p->arrival_time <= current_time) {
                if (idx == -1 ||
                    p->arrival_time < out->processes[idx].arrival_time ||
                    (p->arrival_time == out->processes[idx].arrival_time && p->pid < out->processes[idx].pid)) {
                    idx = i;
                }
            }
        }
        if (idx == -1) {
            current_time++;
            continue;
        }
        Process *p = &out->processes[idx];
        p->state = STATE_RUNNING;
        p->start_time = current_time;
        int start = current_time;
        current_time += p->burst_time;
        p->remaining_time = 0;
        p->completion_time = current_time;
        p->state = STATE_TERMINATED;
        record_segment(out, p, start, current_time);
        completed++;
    }
    compute_averages(out);
}

void schedule_sjf(const Process *input, int n, ScheduleResult *out) {
    init_result("SJF", input, n, out);
    int completed = 0;
    int current_time = 0;

    while (completed < n) {
        int idx = -1;
        for (int i = 0; i < n; i++) {
            Process *p = &out->processes[i];
            if (p->state != STATE_TERMINATED && p->arrival_time <= current_time) {
                if (idx == -1 ||
                    p->burst_time < out->processes[idx].burst_time ||
                    (p->burst_time == out->processes[idx].burst_time &&
                     p->arrival_time < out->processes[idx].arrival_time) ||
                    (p->burst_time == out->processes[idx].burst_time &&
                     p->arrival_time == out->processes[idx].arrival_time &&
                     p->pid < out->processes[idx].pid)) {
                    idx = i;
                }
            }
        }
        if (idx == -1) {
            current_time++;
            continue;
        }
        Process *p = &out->processes[idx];
        p->state = STATE_RUNNING;
        p->start_time = current_time;
        int start = current_time;
        current_time += p->burst_time;
        p->remaining_time = 0;
        p->completion_time = current_time;
        p->state = STATE_TERMINATED;
        record_segment(out, p, start, current_time);
        completed++;
    }
    compute_averages(out);
}

void schedule_priority_nonpreemptive(const Process *input, int n, ScheduleResult *out) {
    init_result("PRIORITY", input, n, out);
    int completed = 0;
    int current_time = 0;

    while (completed < n) {
        int idx = -1;
        for (int i = 0; i < n; i++) {
            Process *p = &out->processes[i];
            if (p->state != STATE_TERMINATED && p->arrival_time <= current_time) {
                if (idx == -1 ||
                    p->priority < out->processes[idx].priority ||
                    (p->priority == out->processes[idx].priority &&
                     p->arrival_time < out->processes[idx].arrival_time) ||
                    (p->priority == out->processes[idx].priority &&
                     p->arrival_time == out->processes[idx].arrival_time &&
                     p->pid < out->processes[idx].pid)) {
                    idx = i;
                }
            }
        }
        if (idx == -1) {
            current_time++;
            continue;
        }
        Process *p = &out->processes[idx];
        p->state = STATE_RUNNING;
        p->start_time = current_time;
        int start = current_time;
        current_time += p->burst_time;
        p->remaining_time = 0;
        p->completion_time = current_time;
        p->state = STATE_TERMINATED;
        record_segment(out, p, start, current_time);
        completed++;
    }
    compute_averages(out);
}