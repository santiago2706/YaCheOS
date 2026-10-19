# Documentación del Módulo de Sincronización y Concurrencia

Este documento detalla la arquitectura, el funcionamiento interno, los comandos de prueba y las validaciones del módulo de concurrencia y sincronización (`syncgo`), que forma parte de la infraestructura del servidor.

---

## 1. Comandos de Prueba y Resultados Esperados

Para validar la corrección de los componentes concurrentes, se dispone de una suite de pruebas unitarias e integrales en Go.

### Pruebas del Semáforo Contador
```bash
go test -v ./src/sync_go -run TestSemaphore