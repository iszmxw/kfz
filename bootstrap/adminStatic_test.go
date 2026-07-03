package bootstrap

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"goapi/pkg/logger"
)

func TestAdminSPAFallbackServesBuiltAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger.Init()

	tempDir := t.TempDir()
	distDir := filepath.Join(tempDir, "admin", "dist")
	assetsDir := filepath.Join(distDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("create admin assets dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "index.html"), []byte("fake-admin-index"), 0o644); err != nil {
		t.Fatalf("write admin index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "app.js"), []byte("console.log('admin')"), 0o644); err != nil {
		t.Fatalf("write admin asset: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	router := SetupRoute(gin.New())

	for _, path := range []string{"/admin", "/admin/", "/admin/dashboard"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d; body = %s", path, recorder.Code, http.StatusOK, recorder.Body.String())
		}
		if !strings.Contains(recorder.Body.String(), "fake-admin-index") {
			t.Fatalf("GET %s body = %q, want admin index", path, recorder.Body.String())
		}
	}

	assetRecorder := httptest.NewRecorder()
	assetRequest := httptest.NewRequest(http.MethodGet, "/admin/assets/app.js", nil)
	router.ServeHTTP(assetRecorder, assetRequest)

	if assetRecorder.Code != http.StatusOK {
		t.Fatalf("GET /admin/assets/app.js status = %d, want %d", assetRecorder.Code, http.StatusOK)
	}
	if strings.TrimSpace(assetRecorder.Body.String()) != "console.log('admin')" {
		t.Fatalf("GET /admin/assets/app.js body = %q, want asset content", assetRecorder.Body.String())
	}

	apiRecorder := httptest.NewRecorder()
	apiRequest := httptest.NewRequest(http.MethodGet, "/admin/api/v1/not-found", nil)
	router.ServeHTTP(apiRecorder, apiRequest)

	if apiRecorder.Code != http.StatusNotFound {
		t.Fatalf("GET /admin/api/v1/not-found status = %d, want %d", apiRecorder.Code, http.StatusNotFound)
	}
	if strings.Contains(apiRecorder.Body.String(), "fake-admin-index") {
		t.Fatalf("GET /admin/api/v1/not-found returned admin index: %q", apiRecorder.Body.String())
	}
}
