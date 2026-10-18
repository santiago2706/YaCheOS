#include "pcb.h"
#include <string.h>

static void append_or_extend_segment(ScheduleResult *out, const Process *p, int start, int end) {
    if (start >= end) return;
    if (out->segment_count > 0) {
        TimelineSegment *last = &out->segments[out->segment_count - 1];
        if (last->pid == p->pid && last->end_time == start) {
            last->end_time = end;
            return;
        }
    }
    if (out->segment_count < MAX_TIMELINE_SEGMENTS) {
        TimelineSegment *seg = &out->segments[out->segment_count++];
        seg->pid = p->pid;
        strncpy(seg->name, p->name, MAX_NAME_LEN - 1);
        seg->start_time = start;
        seg->end_time = end;
    }
}

static void finalize_metrics(ScheduleResult *out) {
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

void schedule_rr(const Process *input, int n, int quantum, ScheduleResult *out) {
    if (quantum <= 0) quantum = 4;
    memset(out, 0, sizeof(ScheduleResult));
    strncpy(out->algorithm, "RR", sizeof(out->algorithm) - 1);
    out->quantum = quantum;
    out->process_count = n;

    bool in_queue[MAX_PROCESSES] = {false};
    for (int i = 0; i < n; i++) {
        out->processes[i] = input[i];
        out->processes[i].remaining_time = input[i].burst_time;
        out->processes[i].effective_priority = input[i].priority;
        out->processes[i].state = STATE_NEW;
        out->processes[i].start_time = -1;
    }

    Queue ready;
    queue_init(&ready);
    int current_time = 0;
    int completed = 0;

    /* Encolar procesos que llegan en t = 0 ordenados por llegada y PID */
    for (int i = 0; i < n; i++) {
        if (out->processes[i].arrival_time <= current_time && !in_queue[i]) {
            out->processes[i].state = STATE_READY;
            queue_enqueue(&ready, &out->processes[i]);
            in_queue[i] = true;
        }
    }

    while (completed < n) {
        if (queue_is_empty(&ready)) {
            current_time++;
            for (int i = 0; i < n; i++) {
                if (!in_queue[i] && out->processes[i].state != STATE_TERMINATED &&
                    out->processes[i].arrival_time <= current_time) {
                    out->processes[i].state = STATE_READY;
                    queue_enqueue(&ready, &out->processes[i]);
                    in_queue[i] = true;
                }
            }
            continue;
        }

        Process *curr = queue_dequeue(&ready);
        curr->state = STATE_RUNNING;
        if (curr->start_time == -1) {
            curr->start_time = current_time;
        }

        int slice = (curr->remaining_time < quantum) ? curr->remaining_time : quantum;
        int start = current_time;
        current_time += slice;
        curr->remaining_time -= slice;
        append_or_extend_segment(out, curr, start, current_time);

        /* Regla del contrato: encolar llegadas ocurridas hasta el final del quantum antes de reinsertar */
        for (int i = 0; i < n; i++) {
            if (!in_queue[i] && out->processes[i].state != STATE_TERMINATED &&
                out->processes[i].arrival_time <= current_time) {
                out->processes[i].state = STATE_READY;
                queue_enqueue(&ready, &out->processes[i]);
                in_queue[i] = true;
            }
        }

        if (curr->remaining_time == 0) {
            curr->state = STATE_TERMINATED;
            curr->completion_time = current_time;
            completed++;
        } else {
            curr->state = STATE_READY;
            queue_enqueue(&ready, curr);
        }
    }

    queue_clear(&ready);
    finalize_metrics(out);
}

void schedule_priority_preemptive(const Process *input, int n, int aging_interval, ScheduleResult *out) {
    memset(out, 0, sizeof(ScheduleResult));
    strncpy(out->algorithm, (aging_interval > 0) ? "PRIORITY_PREEMPTIVE_AGING" : "PRIORITY_PREEMPTIVE",
            sizeof(out->algorithm) - 1);
    out->aging_interval = aging_interval;
    out->process_count = n;

    for (int i = 0; i < n; i++) {
        out->processes[i] = input[i];
        out->processes[i].remaining_time = input[i].burst_time;
        out->processes[i].effective_priority = input[i].priority;
        out->processes[i].state = STATE_NEW;
        out->processes[i].start_time = -1;
        out->processes[i].wait_ticks_since_boost = 0;
    }

    int current_time = 0;
    int completed = 0;

    while (completed < n) {
        for (int i = 0; i < n; i++) {
            if (out->processes[i].state == STATE_NEW && out->processes[i].arrival_time <= current_time) {
                out->processes[i].state = STATE_READY;
            }
        }

        int best = -1;
        for (int i = 0; i < n; i++) {
            Process *p = &out->processes[i];
            if (p->state != STATE_TERMINATED && p->arrival_time <= current_time) {
                if (best == -1 ||
                    p->effective_priority < out->processes[best].effective_priority ||
                    (p->effective_priority == out->processes[best].effective_priority &&
                     p->arrival_time < out->processes[best].arrival_time) ||
                    (p->effective_priority == out->processes[best].effective_priority &&
                     p->arrival_time == out->processes[best].arrival_time &&
                     p->pid < out->processes[best].pid)) {
                    best = i;
                }
            }
        }

        if (best == -1) {
            current_time++;
            continue;
        }

        Process *curr = &out->processes[best];
        curr->state = STATE_RUNNING;
        if (curr->start_time == -1) {
            curr->start_time = current_time;
        }

        /* Ejecutar 1 tick */
        append_or_extend_segment(out, curr, current_time, current_time + 1);
        curr->remaining_time--;
        curr->wait_ticks_since_boost = 0;

        /* Envejecimiento (aging) para los procesos que esperaron en READY durante este tick */
        if (aging_interval > 0) {
            for (int i = 0; i < n; i++) {
                if (i != best && out->processes[i].state == STATE_READY &&
                    out->processes[i].arrival_time <= current_time) {
                    out->processes[i].wait_ticks_since_boost++;
                    if (out->processes[i].wait_ticks_since_boost >= aging_interval) {
                        if (out->processes[i].effective_priority > 1) {
                            out->processes[i].effective_priority--;
                        }
                        out->processes[i].wait_ticks_since_boost = 0;
                    }
                }
            }
        }

        current_time++;
        if (curr->remaining_time == 0) {
            curr->state = STATE_TERMINATED;
            curr->completion_time = current_time;
            completed++;
        } else {
            curr->state = STATE_READY;
        }
    }

    finalize_metrics(out);
}