package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// resolveBaseDir busca el directorio raíz del proyecto que contiene la carpeta 'web'.
func resolveBaseDir() string {
	cwd, err := os.Getwd()
	if err == nil {
		if _, err := os.Stat(filepath.Join(cwd, "web", "index.html")); err == nil {
			return cwd
		}
		parent := filepath.Dir(cwd)
		if _, err := os.Stat(filepath.Join(parent, "web", "index.html")); err == nil {
			return parent
		}
		grandParent := filepath.Dir(parent)
		if _, err := os.Stat(filepath.Join(grandParent, "web", "index.html")); err == nil {
			return grandParent
		}
	}
	return "."
}

// setupRouter configura las rutas HTTP del servidor nativo.
func setupRouter(webDir, dataDir, baseDir string) http.Handler {
	mux := http.NewServeMux()

	// Inicializar manejadores del núcleo (C y Go)
	core := NewCoreHandlers(baseDir)

	// Servir archivos estáticos del frontend (escritorio, terminal, estilos y scripts)
	fsWeb := http.FileServer(http.Dir(webDir))
	mux.Handle("/", fsWeb)

	// Servir contratos de datos JSON para consumo de la interfaz y la terminal
	if _, err := os.Stat(dataDir); err == nil {
		fsData := http.StripPrefix("/data/", http.FileServer(http.Dir(dataDir)))
		mux.Handle("/data/", fsData)
	}
	mux.Handle("/docs/gantt/", http.StripPrefix("/docs/gantt/", http.FileServer(http.Dir(filepath.Join(baseDir, "docs", "gantt")))))

	// Endpoints de Integración M1 y M3 (Sprint 2 - T1.4)
	mux.HandleFunc("/api/v1/cpu/schedule", core.HandleCPUSchedule)
	mux.HandleFunc("/api/v1/process/demo", core.HandleProcessDemo)
	mux.HandleFunc("/api/v1/thread/demo", core.HandleThreadDemo)
	mux.HandleFunc("/api/v1/simulation", core.HandleTaskSimulation)
	mux.HandleFunc("/api/v1/memory/translate", core.HandleMemoryTranslate)
	mux.HandleFunc("/api/v1/memory/replace", core.HandlePageReplacementIntegrated)
	mux.HandleFunc("/api/v1/memory/map", core.HandleMemoryMap)

	// Endpoints de Sincronización M2 (Sprint 2 - T1.6)
	mux.HandleFunc("/api/v1/sync/status", core.HandleSyncStatus)
	mux.HandleFunc("/api/v1/sync/step", core.HandleSyncStep)

	// Endpoint de verificación de estado y bienvenida del sistema
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","system":"YaKcheOS","version":"v0.5.0-beta","sprint":2,"role":"Integrante 1: Lider Tecnico & Orquestador Web"}`))
	})

	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	handler := setupRouter(webDir, dataDir, baseDir)
	addr := fmt.Sprintf(":%s", port)

	log.Println("==================================================")
	log.Println("  YaKcheOS Web Server - UNI FIIS SO-2026")
	log.Println("  Integrante 1: Lider Tecnico & Orquestador Web")
	log.Println("  Version: v0.5.0-beta (Sprint 2 Completado)")
	log.Printf("  Servidor escuchando en: http://localhost:%s\n", port)
	log.Printf("  Directorio Web:        %s\n", webDir)
	log.Printf("  Directorio Data:       %s\n", dataDir)
	log.Println("  Endpoints Activos:")
	log.Println("    - POST/GET /api/v1/cpu/schedule")
	log.Println("    - POST/GET /api/v1/memory/translate")
	log.Println("    - POST/GET /api/v1/memory/replace")
	log.Println("    - GET      /api/v1/sync/status")
	log.Println("    - POST/GET /api/v1/sync/step")
	log.Println("==================================================")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}
