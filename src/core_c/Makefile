CC = gcc
CFLAGS = -Wall -Wextra -O2

BIN_DIR = ../../bin

all: $(BIN_DIR)/scheduler $(BIN_DIR)/memory_translate

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(BIN_DIR)/scheduler: scheduler.c pcb.h | $(BIN_DIR)
	$(CC) $(CFLAGS) -o $@ scheduler.c

$(BIN_DIR)/memory_translate: memory.c | $(BIN_DIR)
	$(CC) $(CFLAGS) -o $@ memory.c

clean:
	rm -f $(BIN_DIR)/scheduler $(BIN_DIR)/memory_translate
