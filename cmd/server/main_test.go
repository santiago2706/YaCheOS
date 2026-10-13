package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRootEndpointReturns200(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	// Verificar que index.html exista en web
	if _, err := os.Stat(filepath.Join(webDir, "index.html")); err != nil {
		t.Fatalf("web/index.html no fue encontrado en %s: %v", webDir, err)
	}

	router := setupRouter(webDir, dataDir)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Se esperaba codigo HTTP 200 para '/', pero se obtuvo %d", rr.Code)
	}
}

func TestHealthEndpointReturns200(t *testing.T) {
	baseDir := resolveBaseDir()
	webDir := filepath.Join(baseDir, "web")
	dataDir := filepath.Join(baseDir, "data")

	router := setupRouter(webDir, dataDir)
	req, err := http.NewRequest(http.MethodGet, "/api/health", nil)
	if err != nil {
		t.Fatalf("Error al crear request HTTP: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Se esperaba codigo HTTP 200 para '/api/health', pero se obtuvo %d", rr.Code)
	}
}
