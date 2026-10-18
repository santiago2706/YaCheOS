#include "pcb.h"

// Parse simple integer from JSON string key
static int parse_json_int(const char *json, const char *key, int default_val) {
    char search[64];
    snprintf(search, sizeof(search), "\"%s\":", key);
    char *pos = strstr(json, search);
    if (!pos) {
        snprintf(search, sizeof(search), "\"%s\" :", key);
        pos = strstr(json, search);
    }
    if (!pos) return default_val;
    pos += strlen(search);
    while (*pos == ' ' || *pos == '\t') pos++;
    return atoi(pos);
}

// Parse string from JSON key
static void parse_json_str(const char *json, const char *key, char *out, size_t out_size, const char *default_val) {
    char search[64];
    snprintf(search, sizeof(search), "\"%s\":", key);
    char *pos = strstr(json, search);
    if (!pos) {
        snprintf(search, sizeof(search), "\"%s\" :", key);
        pos = strstr(json, search);
    }
    if (!pos) {
        strncpy(out, default_val, out_size - 1);
        out[out_size - 1] = '\0';
        return;
    }
    pos += strlen(search);
    while (*pos == ' ' || *pos == '\t' || *pos == '"') pos++;
    size_t i = 0;
    while (*pos && *pos != '"' && *pos != ',' && *pos != '}' && i < out_size - 1) {
        out[i++] = *pos++;
    }
    out[i] = '\0';
}

static int load_processes(const char *filename, Process procs[], int max_procs) {
    FILE *f = fopen(filename, "r");
    if (!f) {
        // Fallback default processes if file cannot be opened (Semana 4 case)
        procs[0] = (Process){ .pid = 1, .arrival_time = 0, .burst_time = 5, .remaining_time = 5, .priority = 3, .memory_kb = 128 };
        strncpy(procs[0].name, "P1", sizeof(procs[0].name));
        procs[1] = (Process){ .pid = 2, .arrival_time = 1, .burst_time = 3, .remaining_time = 3, .priority = 4, .memory_kb = 64 };
        strncpy(procs[1].name, "P2", sizeof(procs[1].name));
        procs[2] = (Process){ .pid = 3, .arrival_time = 2, .burst_time = 8, .remaining_time = 8, .priority = 1, .memory_kb = 256 };
        strncpy(procs[2].name, "P3", sizeof(procs[2].name));
        procs[3] = (Process){ .pid = 4, .arrival_time = 3, .burst_time = 6, .remaining_time = 6, .priority = 2, .memory_kb = 128 };
        strncpy(procs[3].name, "P4", sizeof(procs[3].name));
        return 4;
    }

    char buffer[16384];
    size_t read_bytes = fread(buffer, 1, sizeof(buffer) - 1, f);
    buffer[read_bytes] = '\0';
    fclose(f);

    int count = 0;
    char *cursor = buffer;
    while ((cursor = strchr(cursor, '{')) != NULL && count < max_procs) {
        char *end = strchr(cursor, '}');
        if (!end) break;
        char obj[1024];
        size_t len = (size_t)(end - cursor + 1);
        if (len >= sizeof(obj)) len = sizeof(obj) - 1;
        strncpy(obj, cursor, len);
        obj[len] = '\0';

        procs[count].pid = parse_json_int(obj, "pid", count + 1);
        parse_json_str(obj, "name", procs[count].name, sizeof(procs[count].name), "Px");
        if (strlen(procs[count].name) == 0) {
            parse_json_str(obj, "nombre", procs[count].name, sizeof(procs[count].name), "Px");
        }
        procs[count].arrival_time = parse_json_int(obj, "arrival_time", parse_json_int(obj, "llegada", 0));
        procs[count].burst_time = parse_json_int(obj, "burst_time", parse_json_int(obj, "rafaga", 1));
        procs[count].remaining_time = procs[count].burst_time;
        procs[count].priority = parse_json_int(obj, "priority", parse_json_int(obj, "prioridad", 0));
        procs[count].memory_kb = parse_json_int(obj, "memory_kb", parse_json_int(obj, "memoria", 128));
        procs[count].state = STATE_READY;
        procs[count].start_time = -1;
        procs[count].finish_time = -1;
        procs[count].waiting_time = 0;
        procs[count].turnaround_time = 0;
        procs[count].response_time = -1;

        count++;
        cursor = end + 1;
    }

    if (count == 0) {
        // Fallback default processes
        procs[0] = (Process){ .pid = 1, .arrival_time = 0, .burst_time = 5, .remaining_time = 5, .priority = 3, .memory_kb = 128 };
        strncpy(procs[0].name, "P1", sizeof(procs[0].name));
        procs[1] = (Process){ .pid = 2, .arrival_time = 1, .burst_time = 3, .remaining_time = 3, .priority = 4, .memory_kb = 64 };
        strncpy(procs[1].name, "P2", sizeof(procs[1].name));
        procs[2] = (Process){ .pid = 3, .arrival_time = 2, .burst_time = 8, .remaining_time = 8, .priority = 1, .memory_kb = 256 };
        strncpy(procs[2].name, "P3", sizeof(procs[2].name));
        procs[3] = (Process){ .pid = 4, .arrival_time = 3, .burst_time = 6, .remaining_time = 6, .priority = 2, .memory_kb = 128 };
        strncpy(procs[3].name, "P4", sizeof(procs[3].name));
        return 4;
    }
    return count;
}

// FCFS Algorithm
static void run_fcfs(Process procs[], int n, TimelineSegment timeline[], int *timeline_count) {
    int current_time = 0;
    int completed = 0;
    bool done[32] = {false};
    *timeline_count = 0;

    while (completed < n) {
        int best_idx = -1;
        int min_arrival = 100000;
        for (int i = 0; i < n; i++) {
            if (!done[i] && procs[i].arrival_time <= current_time) {
                if (procs[i].arrival_time < min_arrival) {
                    min_arrival = procs[i].arrival_time;
                    best_idx = i;
                } else if (procs[i].arrival_time == min_arrival && best_idx >= 0 && procs[i].pid < procs[best_idx].pid) {
                    best_idx = i;
                }
            }
        }

        if (best_idx == -1) {
            // Find next arriving process
            int next_arr = 100000;
            for (int i = 0; i < n; i++) {
                if (!done[i] && procs[i].arrival_time < next_arr) {
                    next_arr = procs[i].arrival_time;
                }
            }
            if (next_arr == 100000) break;
            current_time = next_arr;
            continue;
        }

        int start = current_time;
        int duration = procs[best_idx].burst_time;
        int end = start + duration;

        procs[best_idx].start_time = start;
        procs[best_idx].finish_time = end;
        procs[best_idx].response_time = start - procs[best_idx].arrival_time;
        procs[best_idx].turnaround_time = end - procs[best_idx].arrival_time;
        procs[best_idx].waiting_time = procs[best_idx].turnaround_time - procs[best_idx].burst_time;
        procs[best_idx].state = STATE_TERMINATED;

        timeline[*timeline_count] = (TimelineSegment){
            .pid = procs[best_idx].pid,
            .start = start,
            .end = end,
            .duration = duration
        };
        strncpy(timeline[*timeline_count].name, procs[best_idx].name, sizeof(timeline[*timeline_count].name));
        (*timeline_count)++;

        done[best_idx] = true;
        completed++;
        current_time = end;
    }
}

// SJF (Non-preemptive)
static void run_sjf(Process procs[], int n, TimelineSegment timeline[], int *timeline_count) {
    int current_time = 0;
    int completed = 0;
    bool done[32] = {false};
    *timeline_count = 0;

    while (completed < n) {
        int best_idx = -1;
        int min_burst = 100000;
        for (int i = 0; i < n; i++) {
            if (!done[i] && procs[i].arrival_time <= current_time) {
                if (procs[i].burst_time < min_burst) {
                    min_burst = procs[i].burst_time;
                    best_idx = i;
                } else if (procs[i].burst_time == min_burst && best_idx >= 0) {
                    if (procs[i].arrival_time < procs[best_idx].arrival_time) {
                        best_idx = i;
                    } else if (procs[i].arrival_time == procs[best_idx].arrival_time && procs[i].pid < procs[best_idx].pid) {
                        best_idx = i;
                    }
                }
            }
        }

        if (best_idx == -1) {
            int next_arr = 100000;
            for (int i = 0; i < n; i++) {
                if (!done[i] && procs[i].arrival_time < next_arr) {
                    next_arr = procs[i].arrival_time;
                }
            }
            if (next_arr == 100000) break;
            current_time = next_arr;
            continue;
        }

        int start = current_time;
        int duration = procs[best_idx].burst_time;
        int end = start + duration;

        procs[best_idx].start_time = start;
        procs[best_idx].finish_time = end;
        procs[best_idx].response_time = start - procs[best_idx].arrival_time;
        procs[best_idx].turnaround_time = end - procs[best_idx].arrival_time;
        procs[best_idx].waiting_time = procs[best_idx].turnaround_time - procs[best_idx].burst_time;
        procs[best_idx].state = STATE_TERMINATED;

        timeline[*timeline_count] = (TimelineSegment){
            .pid = procs[best_idx].pid,
            .start = start,
            .end = end,
            .duration = duration
        };
        strncpy(timeline[*timeline_count].name, procs[best_idx].name, sizeof(timeline[*timeline_count].name));
        (*timeline_count)++;

        done[best_idx] = true;
        completed++;
        current_time = end;
    }
}

// Priority (Non-preemptive, lower number = higher priority)
static void run_priority(Process procs[], int n, TimelineSegment timeline[], int *timeline_count) {
    int current_time = 0;
    int completed = 0;
    bool done[32] = {false};
    *timeline_count = 0;

    while (completed < n) {
        int best_idx = -1;
        int best_prio = 100000;
        for (int i = 0; i < n; i++) {
            if (!done[i] && procs[i].arrival_time <= current_time) {
                if (procs[i].priority < best_prio) {
                    best_prio = procs[i].priority;
                    best_idx = i;
                } else if (procs[i].priority == best_prio && best_idx >= 0) {
                    if (procs[i].arrival_time < procs[best_idx].arrival_time) {
                        best_idx = i;
                    } else if (procs[i].arrival_time == procs[best_idx].arrival_time && procs[i].pid < procs[best_idx].pid) {
                        best_idx = i;
                    }
                }
            }
        }

        if (best_idx == -1) {
            int next_arr = 100000;
            for (int i = 0; i < n; i++) {
                if (!done[i] && procs[i].arrival_time < next_arr) {
                    next_arr = procs[i].arrival_time;
                }
            }
            if (next_arr == 100000) break;
            current_time = next_arr;
            continue;
        }

        int start = current_time;
        int duration = procs[best_idx].burst_time;
        int end = start + duration;

        procs[best_idx].start_time = start;
        procs[best_idx].finish_time = end;
        procs[best_idx].response_time = start - procs[best_idx].arrival_time;
        procs[best_idx].turnaround_time = end - procs[best_idx].arrival_time;
        procs[best_idx].waiting_time = procs[best_idx].turnaround_time - procs[best_idx].burst_time;
        procs[best_idx].state = STATE_TERMINATED;

        timeline[*timeline_count] = (TimelineSegment){
            .pid = procs[best_idx].pid,
            .start = start,
            .end = end,
            .duration = duration
        };
        strncpy(timeline[*timeline_count].name, procs[best_idx].name, sizeof(timeline[*timeline_count].name));
        (*timeline_count)++;

        done[best_idx] = true;
        completed++;
        current_time = end;
    }
}

// Round Robin (Quantum = 4)
static void run_rr(Process procs[], int n, int quantum, TimelineSegment timeline[], int *timeline_count) {
    int queue[512];
    int q_head = 0, q_tail = 0;
    bool in_queue[32] = {false};
    bool visited[32] = {false};
    int current_time = 0;
    int completed = 0;
    *timeline_count = 0;

    // Initial arrivals at t=0
    for (int i = 0; i < n; i++) {
        if (procs[i].arrival_time <= current_time && !in_queue[i]) {
            queue[q_tail++] = i;
            in_queue[i] = true;
        }
    }

    while (completed < n) {
        if (q_head == q_tail) {
            // Find next arrival
            int next_arr = 100000;
            for (int i = 0; i < n; i++) {
                if (procs[i].remaining_time > 0 && procs[i].arrival_time < next_arr) {
                    next_arr = procs[i].arrival_time;
                }
            }
            if (next_arr == 100000) break;
            current_time = next_arr;
            for (int i = 0; i < n; i++) {
                if (procs[i].remaining_time > 0 && procs[i].arrival_time <= current_time && !in_queue[i]) {
                    queue[q_tail++] = i;
                    in_queue[i] = true;
                }
            }
            continue;
        }

        int idx = queue[q_head++];
        in_queue[idx] = false;

        if (!visited[idx]) {
            procs[idx].start_time = current_time;
            procs[idx].response_time = current_time - procs[idx].arrival_time;
            visited[idx] = true;
        }

        int slice = (procs[idx].remaining_time < quantum) ? procs[idx].remaining_time : quantum;
        int start = current_time;
        int end = current_time + slice;
        procs[idx].remaining_time -= slice;
        current_time = end;

        timeline[*timeline_count] = (TimelineSegment){
            .pid = procs[idx].pid,
            .start = start,
            .end = end,
            .duration = slice
        };
        strncpy(timeline[*timeline_count].name, procs[idx].name, sizeof(timeline[*timeline_count].name));
        (*timeline_count)++;

        // Enqueue new arrivals during this execution slice before re-enqueueing current process
        for (int i = 0; i < n; i++) {
            if (i != idx && procs[i].remaining_time > 0 && procs[i].arrival_time <= current_time && !in_queue[i]) {
                bool already_done = false;
                for (int k = q_head; k < q_tail; k++) {
                    if (queue[k] == i) { already_done = true; break; }
                }
                if (!already_done) {
                    queue[q_tail++] = i;
                    in_queue[i] = true;
                }
            }
        }

        if (procs[idx].remaining_time > 0) {
            queue[q_tail++] = idx;
            in_queue[idx] = true;
        } else {
            procs[idx].finish_time = current_time;
            procs[idx].turnaround_time = procs[idx].finish_time - procs[idx].arrival_time;
            procs[idx].waiting_time = procs[idx].turnaround_time - procs[idx].burst_time;
            procs[idx].state = STATE_TERMINATED;
            completed++;
        }
    }
}

int main(int argc, char *argv[]) {
    char algo[32] = "FCFS";
    char filename[256] = "data/procesos_semana4.json";
    int quantum = 4;

    for (int i = 1; i < argc; i++) {
        if (strcmp(argv[i], "--algo") == 0 && i + 1 < argc) {
            strncpy(algo, argv[++i], sizeof(algo) - 1);
        } else if (strcmp(argv[i], "--file") == 0 && i + 1 < argc) {
            strncpy(filename, argv[++i], sizeof(filename) - 1);
        } else if (strcmp(argv[i], "--quantum") == 0 && i + 1 < argc) {
            quantum = atoi(argv[++i]);
        }
    }

    // Convert algo to uppercase
    for (int i = 0; algo[i]; i++) {
        if (algo[i] >= 'a' && algo[i] <= 'z') algo[i] -= 32;
    }

    if (strcmp(algo, "FCFS") != 0 && strcmp(algo, "SJF") != 0 &&
        strcmp(algo, "RR") != 0 && strcmp(algo, "ROUNDROBIN") != 0 &&
        strcmp(algo, "PRIORIDAD") != 0 && strcmp(algo, "PRIORITY") != 0) {
        fprintf(stderr, "Error: Algoritmo no valido '%s'. Algoritmos permitidos: FCFS, SJF, RR, PRIORIDAD.\n", algo);
        return 1;
    }

    Process procs[32];
    int n = load_processes(filename, procs, 32);
    if (n <= 0) {
        fprintf(stderr, "Error: No se pudieron cargar procesos desde '%s'\n", filename);
        return 2;
    }

    TimelineSegment timeline[128];
    int timeline_count = 0;

    if (strcmp(algo, "FCFS") == 0) {
        run_fcfs(procs, n, timeline, &timeline_count);
    } else if (strcmp(algo, "SJF") == 0) {
        run_sjf(procs, n, timeline, &timeline_count);
    } else if (strcmp(algo, "RR") == 0 || strcmp(algo, "ROUNDROBIN") == 0) {
        run_rr(procs, n, quantum, timeline, &timeline_count);
    } else if (strcmp(algo, "PRIORIDAD") == 0 || strcmp(algo, "PRIORITY") == 0) {
        run_priority(procs, n, timeline, &timeline_count);
    }

    double total_wait = 0, total_turn = 0, total_resp = 0;
    int max_time = 0;
    for (int i = 0; i < n; i++) {
        total_wait += procs[i].waiting_time;
        total_turn += procs[i].turnaround_time;
        total_resp += procs[i].response_time;
        if (procs[i].finish_time > max_time) max_time = procs[i].finish_time;
    }
    double avg_wait = total_wait / n;
    double avg_turn = total_turn / n;
    double avg_resp = total_resp / n;

    // Print JSON output to stdout
    printf("{\n");
    printf("  \"algorithm\": \"%s\",\n", algo);
    printf("  \"total_time\": %d,\n", max_time);
    printf("  \"average_waiting_time\": %.2f,\n", avg_wait);
    printf("  \"average_turnaround_time\": %.2f,\n", avg_turn);
    printf("  \"average_response_time\": %.2f,\n", avg_resp);
    printf("  \"timeline\": [\n");
    for (int i = 0; i < timeline_count; i++) {
        printf("    {\"pid\": %d, \"name\": \"%s\", \"start\": %d, \"end\": %d, \"duration\": %d}%s\n",
               timeline[i].pid, timeline[i].name, timeline[i].start, timeline[i].end, timeline[i].duration,
               (i == timeline_count - 1) ? "" : ",");
    }
    printf("  ],\n");
    printf("  \"processes\": [\n");
    for (int i = 0; i < n; i++) {
        printf("    {\"pid\": %d, \"name\": \"%s\", \"arrival\": %d, \"burst\": %d, \"priority\": %d, \"memory_kb\": %d, \"waiting_time\": %d, \"turnaround_time\": %d, \"response_time\": %d}%s\n",
               procs[i].pid, procs[i].name, procs[i].arrival_time, procs[i].burst_time, procs[i].priority, procs[i].memory_kb,
               procs[i].waiting_time, procs[i].turnaround_time, procs[i].response_time,
               (i == n - 1) ? "" : ",");
    }
    printf("  ]\n");
    printf("}\n");

    return 0;
}
