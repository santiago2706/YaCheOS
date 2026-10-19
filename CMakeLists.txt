cmake_minimum_required(VERSION 4.0)
project(YaCheOS)

set(CMAKE_CXX_STANDARD 14)

include_directories(src/core_c)

add_executable(YaCheOS
        .idea/.gitignore
        .idea/workspace.xml
        data/procesos_semana4.json
        src/core_c/Makefile
        src/core_c/memory_contigous.c
        src/core_c/pcb.h
        src/core_c/scheduler_nonpreemptive.c
        src/sync_go/philosophers.go
        src/sync_go/prod_cons.go
        src/sync_go/semaphore.go
        web/css/desktop.css
        web/css/terminal.css
        web/js/app.js
        web/index.html
        .gitignore
        LICENSE
        src/core_c/process_demo.c
        src/core_c/memory_paging.c
        src/core_c/queue.c
        src/core_c/scheduler_preemptive.c
        src/core_c/scheduler.c
        src/core_c/memory_segmentation.c
        src/core_c/memory.c)
