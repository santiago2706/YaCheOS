CC = gcc
CFLAGS = -std=c11 -Wall -Wextra -Werror -O2
BUILD_DIR = ../../build

SCHED_SRCS = queue.c scheduler_nonpreemptive.c scheduler_preemptive.c scheduler.c
MEM_SRCS = memory_contiguous.c memory_paging.c memory_segmentation.c memory.c
ENGINE_SRCS = queue.c scheduler_nonpreemptive.c scheduler_preemptive.c scheduler.c memory_contiguous.c memory_paging.c memory_segmentation.c memory.c

.PHONY: all clean test

all: process_demo scheduler memory_cli $(BUILD_DIR)/core_engine

process_demo: process_demo.c
	$(CC) $(CFLAGS) -o process_demo process_demo.c

scheduler: $(SCHED_SRCS) pcb.h
	$(CC) $(CFLAGS) -o scheduler $(SCHED_SRCS)

memory_cli: $(MEM_SRCS) pcb.h
	$(CC) $(CFLAGS) -DCORE_ENGINE_UNIFIED -o memory_cli $(ENGINE_SRCS)

$(BUILD_DIR)/core_engine: $(ENGINE_SRCS) pcb.h
	mkdir -p $(BUILD_DIR)
	$(CC) $(CFLAGS) -DCORE_ENGINE_UNIFIED -o $(BUILD_DIR)/core_engine $(ENGINE_SRCS)

clean:
	rm -f process_demo scheduler memory_cli cpu_timeline.json
	rm -rf $(BUILD_DIR)/core_engine