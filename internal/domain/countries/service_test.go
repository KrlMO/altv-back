package countries

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

type mockRepository struct {
	countries []Country
	err       error
}

func (m *mockRepository) GetAllCountries(ctx context.Context) ([]Country, error) {
	return m.countries, m.err
}

func TestGetAllCountries_Success(t *testing.T) {
	mock := &mockRepository{
		countries: []Country{
			{ID: 1, Name: "Argentina", IsoCode: "AR", FlagPath: sql.NullString{String: "https://flag.com/ar.png", Valid: true}},
			{ID: 2, Name: "Brasil", IsoCode: "BR", FlagPath: sql.NullString{Valid: false}},
		},
	}

	svc := NewService(mock)
	result, err := svc.GetAllCountries(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 countries, got %d", len(result))
	}

	if result[0].Name != "Argentina" {
		t.Errorf("expected name Argentina, got %s", result[0].Name)
	}

	if result[0].FlagPath == nil || *result[0].FlagPath != "https://flag.com/ar.png" {
		t.Errorf("expected flag path to be set correctly")
	}

	if result[1].FlagPath != nil {
		t.Errorf("expected nil flag path for Brasil")
	}
}

func TestGetAllCountries_Empty(t *testing.T) {
	mock := &mockRepository{
		countries: []Country{},
	}

	svc := NewService(mock)
	result, err := svc.GetAllCountries(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 countries, got %d", len(result))
	}
}

func TestGetAllCountries_Error(t *testing.T) {
	mock := &mockRepository{
		err: errors.New("database connection failed"),
	}

	svc := NewService(mock)
	_, err := svc.GetAllCountries(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "database connection failed" {
		t.Errorf("expected error message 'database connection failed', got %s", err.Error())
	}
}
