package countries

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter(svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	module := NewModule(svc)
	module.RegisterRoutes(r)
	return r
}

func TestGetAllCountries_Handler(t *testing.T) {
	mock := &mockRepository{
		countries: []Country{
			{ID: 1, Name: "Argentina", IsoCode: "AR"},
			{ID: 2, Name: "Chile", IsoCode: "CL"},
		},
	}

	svc := NewService(mock)
	r := setupRouter(svc)

	req := httptest.NewRequest("GET", "/api/v1/countries/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var result []CountryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 countries, got %d", len(result))
	}

	if result[0].Name != "Argentina" {
		t.Errorf("expected name Argentina, got %s", result[0].Name)
	}
}

func TestGetAllCountries_Handler_Empty(t *testing.T) {
	mock := &mockRepository{
		countries: []Country{},
	}

	svc := NewService(mock)
	r := setupRouter(svc)

	req := httptest.NewRequest("GET", "/api/v1/countries/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var result []CountryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 countries, got %d", len(result))
	}
}

func TestGetAllCountries_Handler_Error(t *testing.T) {
	mock := &mockRepository{
		err: errors.New("database error"),
	}

	svc := NewService(mock)
	r := setupRouter(svc)

	req := httptest.NewRequest("GET", "/api/v1/countries/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
