#include "pcb.h"
#include <stdlib.h>

void queue_init(Queue *q) {
    if (!q) return;
    q->head = NULL;
    q->tail = NULL;
    q->size = 0;
}

bool queue_is_empty(const Queue *q) {
    return (q == NULL || q->head == NULL || q->size == 0);
}

void queue_enqueue(Queue *q, Process *p) {
    if (!q || !p) return;
    QueueNode *node = (QueueNode *)malloc(sizeof(QueueNode));
    if (!node) return;
    node->process = p;
    node->next = NULL;

    if (q->tail) {
        q->tail->next = node;
    } else {
        q->head = node;
    }
    q->tail = node;
    q->size++;
}

Process *queue_dequeue(Queue *q) {
    if (queue_is_empty(q)) return NULL;
    QueueNode *node = q->head;
    Process *p = node->process;
    q->head = node->next;
    if (q->head == NULL) {
        q->tail = NULL;
    }
    free(node);
    q->size--;
    return p;
}

void queue_clear(Queue *q) {
    while (!queue_is_empty(q)) {
        queue_dequeue(q);
    }
}