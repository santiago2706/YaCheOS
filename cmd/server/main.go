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
func setupRouter(webDir, dataDir string) http.Handler {
	mux := http.NewServeMux()

	// Servir archivos estáticos del frontend (escritorio, terminal, estilos y scripts)
	fsWeb := http.FileServer(http.Dir(webDir))
	mux.Handle("/", fsWeb)

	// Servir contratos de datos JSON para consumo de la interfaz y la terminal
	if _, err := os.Stat(dataDir); err == nil {
		fsData := http.StripPrefix("/data/", http.FileServer(http.Dir(dataDir)))
		mux.Handle("/data/", fsData)
	}

	// Endpoint de verificación de estado y bienvenida del sistema
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","system":"YaKcheOS","version":"0.1.0-alpha","sprint":1,"role":"Integrante 1: Lider Tecnico & Orquestador Web"}`))
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

	handler := setupRouter(webDir, dataDir)
	addr := fmt.Sprintf(":%s", port)

	log.Println("==================================================")
	log.Println("  YaKcheOS Web Server - UNI FIIS SO-2026")
	log.Println("  Integrante 1: Lider Tecnico & Orquestador Web")
	log.Printf("  Servidor escuchando en: http://localhost:%s\n", port)
	log.Printf("  Directorio Web:        %s\n", webDir)
	log.Printf("  Directorio Data:       %s\n", dataDir)
	log.Println("==================================================")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}
