#include "pcb.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void extract_int_field(const char *line, const char *key, int *out_val) {
    const char *pos = strstr(line, key);
    if (!pos) return;
    pos = strchr(pos, ':');
    if (!pos) return;
    *out_val = atoi(pos + 1);
}

static void extract_str_field(const char *line, const char *key, char *out_str, size_t max_len) {
    const char *pos = strstr(line, key);
    if (!pos) return;
    pos = strchr(pos, ':');
    if (!pos) return;
    pos = strchr(pos, '"');
    if (!pos) return;
    pos++;
    const char *end = strchr(pos, '"');
    if (!end) return;
    size_t len = (size_t)(end - pos);
    if (len >= max_len) len = max_len - 1;
    memcpy(out_str, pos, len);
    out_str[len] = '\0';
}

int load_processes_from_json(const char *filepath, Process *out_procs, int max_procs) {
    FILE *fp = fopen(filepath, "r");
    if (!fp) return -1;

    char line[4096];
    size_t used = 0;
    int ch;
    bool in_object = false, in_string = false, escaped = false;
    int count = 0;
    while ((ch = fgetc(fp)) != EOF) {
        if (!in_object) {
            if (ch != '{') continue;
            in_object = true;
            used = 0;
        }
        if (used + 1 >= sizeof(line)) {
            fclose(fp);
            return -1;
        }
        line[used++] = (char)ch;
        if (in_string) {
            if (escaped) escaped = false;
            else if (ch == '\\') escaped = true;
            else if (ch == '"') in_string = false;
            continue;
        }
        if (ch == '"') in_string = true;
        if (ch != '}') continue;
        line[used] = '\0';
        in_object = false;
        if (strstr(line, "\"pid\"")) {
            if (count >= max_procs) {
                fclose(fp);
                return -1;
            }
            Process p;
            memset(&p, 0, sizeof(Process));
            p.priority = 1;
            p.memory_kb = 64;
            extract_str_field(line, "\"nombre\"", p.name, sizeof(p.name));
            extract_int_field(line, "\"llegada\"", &p.arrival_time);
            extract_int_field(line, "\"rafaga\"", &p.burst_time);
            extract_int_field(line, "\"prioridad\"", &p.priority);
            extract_int_field(line, "\"memoria\"", &p.memory_kb);
            extract_int_field(line, "\"pid\"", &p.pid);
            extract_str_field(line, "\"name\"", p.name, sizeof(p.name));
            extract_int_field(line, "\"arrival_time\"", &p.arrival_time);
            extract_int_field(line, "\"burst_time\"", &p.burst_time);
            extract_int_field(line, "\"priority\"", &p.priority);
            extract_int_field(line, "\"memory_kb\"", &p.memory_kb);
            if (p.pid <= 0 || p.arrival_time < 0 || p.burst_time <= 0 || p.priority <= 0) {
                fclose(fp);
                return -1;
            }
            for (int i = 0; i < count; i++) {
                if (out_procs[i].pid == p.pid) {
                    fclose(fp);
                    return -1;
                }
            }
            p.remaining_time = p.burst_time;
            p.effective_priority = p.priority;
            if (p.name[0] == '\0') {
                snprintf(p.name, sizeof(p.name), "P%d", p.pid);
            }
            out_procs[count++] = p;
        }
    }
    fclose(fp);
    return in_object ? -1 : count;
}

static void write_schedule_stream(FILE *fp, const ScheduleResult *res) {
    fprintf(fp, "{\n");
    fprintf(fp, "  \"algorithm\": \"%s\",\n", res->algorithm);
    int total_time = 0;
    for (int i = 0; i < res->process_count; i++) {
        if (res->processes[i].completion_time > total_time)
            total_time = res->processes[i].completion_time;
    }
    fprintf(fp, "  \"total_time\": %d,\n", total_time);
    fprintf(fp, "  \"average_waiting_time\": %.2f,\n", res->avg_waiting_time);
    fprintf(fp, "  \"average_turnaround_time\": %.2f,\n", res->avg_turnaround_time);
    fprintf(fp, "  \"average_response_time\": %.2f,\n", res->avg_response_time);
    if (res->quantum > 0) {
        fprintf(fp, "  \"quantum\": %d,\n", res->quantum);
    }
    if (res->aging_interval > 0) {
        fprintf(fp, "  \"aging_interval\": %d,\n", res->aging_interval);
    }
    fprintf(fp, "  \"timeline\": [\n");
    for (int i = 0; i < res->segment_count; i++) {
        const TimelineSegment *s = &res->segments[i];
        fprintf(fp, "    {\"pid\": %d, \"name\": \"%s\", \"start\": %d, \"end\": %d, \"duration\": %d}%s\n",
                s->pid, s->name, s->start_time, s->end_time, s->end_time - s->start_time,
                (i + 1 < res->segment_count) ? "," : "");
    }
    fprintf(fp, "  ],\n");
    fprintf(fp, "  \"processes\": [\n");
    for (int i = 0; i < res->process_count; i++) {
        const Process *p = &res->processes[i];
        fprintf(fp,
                "    {\"pid\": %d, \"name\": \"%s\", \"arrival_time\": %d, \"burst_time\": %d, "
                "\"arrival\": %d, \"burst\": %d, \"memory_kb\": %d, "
                "\"priority\": %d, \"start_time\": %d, \"completion_time\": %d, "
                "\"waiting_time\": %d, \"turnaround_time\": %d, \"response_time\": %d}%s\n",
                p->pid, p->name, p->arrival_time, p->burst_time,
                p->arrival_time, p->burst_time, p->memory_kb, p->priority,
                p->start_time, p->completion_time, p->waiting_time,
                p->turnaround_time, p->response_time,
                (i + 1 < res->process_count) ? "," : "");
    }
    fprintf(fp, "  ],\n");
    fprintf(fp, "  \"metrics\": {\n");
    fprintf(fp, "    \"avg_waiting_time\": %.2f,\n", res->avg_waiting_time);
    fprintf(fp, "    \"avg_turnaround_time\": %.2f,\n", res->avg_turnaround_time);
    fprintf(fp, "    \"avg_response_time\": %.2f\n", res->avg_response_time);
    fprintf(fp, "  }\n");
    fprintf(fp, "}\n");
}

void print_schedule_json(const ScheduleResult *res, const char *out_filepath) {
    write_schedule_stream(stdout, res);
    if (out_filepath && out_filepath[0] != '\0') {
        FILE *fp = fopen(out_filepath, "w");
        if (fp) {
            write_schedule_stream(fp, res);
            fclose(fp);
        }
    }
}

#ifndef CORE_ENGINE_UNIFIED
int main(int argc, char *argv[]) {
    const char *algo = "FCFS";
    const char *filepath = "data/procesos_semana4.json";
    const char *outpath = "cpu_timeline.json";
    int quantum = 4;
    int aging = 0;

    for (int i = 1; i < argc; i++) {
        if (strcmp(argv[i], "--algo") == 0 && i + 1 < argc) {
            algo = argv[++i];
        } else if (strcmp(argv[i], "--file") == 0 && i + 1 < argc) {
            filepath = argv[++i];
        } else if (strcmp(argv[i], "--quantum") == 0 && i + 1 < argc) {
            quantum = atoi(argv[++i]);
        } else if (strcmp(argv[i], "--aging") == 0 && i + 1 < argc) {
            aging = atoi(argv[++i]);
        } else if (strcmp(argv[i], "--out") == 0 && i + 1 < argc) {
            outpath = argv[++i];
        }
    }

    Process procs[MAX_PROCESSES];
    int n = load_processes_from_json(filepath, procs, MAX_PROCESSES);
    if (n <= 0) {
        fprintf(stderr, "{\"error\": \"No se pudo cargar el archivo %s\"}\n", filepath);
        return EXIT_FAILURE;
    }

    ScheduleResult result;
    if (strcmp(algo, "FCFS") == 0) {
        schedule_fcfs(procs, n, &result);
    } else if (strcmp(algo, "SJF") == 0) {
        schedule_sjf(procs, n, &result);
    } else if (strcmp(algo, "RR") == 0 || strcmp(algo, "ROUNDROBIN") == 0) {
        schedule_rr(procs, n, quantum, &result);
    } else if (strcmp(algo, "PRIORITY") == 0 || strcmp(algo, "PRIORIDAD") == 0) {
        schedule_priority_nonpreemptive(procs, n, &result);
    } else if (strcmp(algo, "PRIORITY_PREEMPTIVE") == 0) {
        schedule_priority_preemptive(procs, n, aging, &result);
    } else {
        fprintf(stderr, "{\"error\": \"Algoritmo no soportado: %s\"}\n", algo);
        return EXIT_FAILURE;
    }

    print_schedule_json(&result, outpath);
    return EXIT_SUCCESS;
}
#endif
