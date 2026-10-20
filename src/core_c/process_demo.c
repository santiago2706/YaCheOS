#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/wait.h>

int main(void) {
    pid_t pid = fork();

    if (pid < 0) {
        perror("Error en fork");
        return EXIT_FAILURE;
    }

    if (pid == 0) {
        /* Proceso Hijo */
        printf("{\n");
        printf("  \"event\": \"child_exec\",\n");
        printf("  \"child_pid\": %d,\n", (int)getpid());
        printf("  \"parent_pid\": %d,\n", (int)getppid());
        fflush(stdout);

        char *args[] = {"echo", "  \"child_message\": \"Subproceso real ejecutado con execvp y recolectado sin zombies\",", NULL};
        execvp(args[0], args);
        perror("Error en execvp");
        _exit(EXIT_FAILURE);
    } else {
        /* Proceso Padre */
        int status = 0;
        pid_t waited = waitpid(pid, &status, 0);
        if (waited < 0) {
            perror("Error en waitpid");
            return EXIT_FAILURE;
        }
        int exit_code = WIFEXITED(status) ? WEXITSTATUS(status) : -1;
        printf("  \"reaped_child_pid\": %d,\n", (int)waited);
        printf("  \"exit_status\": %d,\n", exit_code);
        printf("  \"zombie_prevented\": true,\n");
        printf("  \"note\": \"PID anfitrion gestionado por kernel Linux; PCB de YaCheOS es determinista en espacio de usuario.\"\n");
        printf("}\n");
    }

    return EXIT_SUCCESS;
}