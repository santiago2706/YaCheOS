CC = gcc
CFLAGS = -std=c11 -Wall -Wextra -Werror -O2
BUILD_DIR = ../../build
BIN_DIR = ../../bin

ifeq ($(OS),Windows_NT)
EXE = .exe
else
EXE =
POSIX_TARGETS = process_demo
endif

SCHED_SRCS = queue.c scheduler_nonpreemptive.c scheduler_preemptive.c scheduler.c
MEM_SRCS = memory_contigous.c memory_paging.c memory_segmentation.c memory.c
ENGINE_SRCS = $(SCHED_SRCS) $(MEM_SRCS)

.PHONY: all clean test

all: $(POSIX_TARGETS) scheduler$(EXE) memory_cli$(EXE) $(BUILD_DIR)/core_engine$(EXE) $(BIN_DIR)/scheduler$(EXE) $(BIN_DIR)/memory_translate$(EXE)

$(BIN_DIR) $(BUILD_DIR):
	@mkdir "$@"

$(BIN_DIR)/scheduler$(EXE): $(SCHED_SRCS) pcb.h | $(BIN_DIR)
	$(CC) $(CFLAGS) -o $@ $(SCHED_SRCS)

$(BIN_DIR)/memory_translate$(EXE): memory_translate.c | $(BIN_DIR)
	$(CC) $(CFLAGS) -o $@ memory_translate.c

process_demo: process_demo.c
	$(CC) $(CFLAGS) -o process_demo process_demo.c

scheduler$(EXE): $(SCHED_SRCS) pcb.h
	$(CC) $(CFLAGS) -o $@ $(SCHED_SRCS)

memory_cli$(EXE): $(ENGINE_SRCS) pcb.h
	$(CC) $(CFLAGS) -DCORE_ENGINE_UNIFIED -o $@ $(ENGINE_SRCS)

$(BUILD_DIR)/core_engine$(EXE): $(ENGINE_SRCS) pcb.h | $(BUILD_DIR)
	$(CC) $(CFLAGS) -DCORE_ENGINE_UNIFIED -o $@ $(ENGINE_SRCS)

clean:
	rm -f process_demo scheduler memory_cli cpu_timeline.json
	rm -rf $(BUILD_DIR)/core_engine
	rm -f $(BIN_DIR)/scheduler$(EXE) $(BIN_DIR)/memory_translate$(EXE)
