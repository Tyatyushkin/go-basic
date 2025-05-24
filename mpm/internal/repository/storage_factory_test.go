package repository

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestCreateStorage_JSON(t *testing.T) {
	tempDir := t.TempDir()
	saveInterval := 30 * time.Second

	storage, err := CreateStorage("json", tempDir, saveInterval)
	if err != nil {
		t.Errorf("CreateStorage() error = %v", err)
	}

	if storage == nil {
		t.Error("Expected storage to be created")
	}

	jsonStorage, ok := storage.(*JSONStorage)
	if !ok {
		t.Errorf("Expected JSONStorage, got %T", storage)
	}

	if jsonStorage.dataDir != tempDir {
		t.Errorf("Expected dataDir to be %s, got %s", tempDir, jsonStorage.dataDir)
	}

	if jsonStorage.saveInterval != saveInterval {
		t.Errorf("Expected saveInterval to be %v, got %v", saveInterval, jsonStorage.saveInterval)
	}
}

func TestCreateStorage_JSON_CaseInsensitive(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []string{"JSON", "Json", "jSoN", "json"}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			storage, err := CreateStorage(tc, tempDir, time.Second)
			if err != nil {
				t.Errorf("CreateStorage() with '%s' error = %v", tc, err)
			}

			if storage == nil {
				t.Errorf("Expected storage to be created for '%s'", tc)
			}

			_, ok := storage.(*JSONStorage)
			if !ok {
				t.Errorf("Expected JSONStorage for '%s', got %T", tc, storage)
			}
		})
	}
}

func TestCreateStorage_EmptyType(t *testing.T) {
	tempDir := t.TempDir()
	saveInterval := 30 * time.Second

	storage, err := CreateStorage("", tempDir, saveInterval)
	if err != nil {
		t.Errorf("CreateStorage() with empty type error = %v", err)
	}

	if storage == nil {
		t.Error("Expected storage to be created with empty type")
	}

	_, ok := storage.(*JSONStorage)
	if !ok {
		t.Errorf("Expected JSONStorage as default, got %T", storage)
	}
}

func TestCreateStorage_Postgres(t *testing.T) {
	tempDir := t.TempDir()
	saveInterval := 30 * time.Second

	originalEnv := os.Getenv("MPM_DATABASE_URL")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("MPM_DATABASE_URL", originalEnv)
		} else {
			_ = os.Unsetenv("MPM_DATABASE_URL")
		}
	}()

	t.Run("Without connection string", func(t *testing.T) {
		_ = os.Unsetenv("MPM_DATABASE_URL")

		storage, err := CreateStorage("postgres", tempDir, saveInterval)
		if err == nil {
			t.Error("Expected error when MPM_DATABASE_URL is not set")
		}

		if storage != nil {
			t.Error("Expected storage to be nil when connection string is missing")
		}

		expectedError := "не указана строка подключения к PostgreSQL"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("With connection string", func(t *testing.T) {
		_ = os.Setenv("MPM_DATABASE_URL", "postgresql://user:pass@localhost/db")

		storage, err := CreateStorage("postgres", tempDir, saveInterval)
		if err == nil {
			t.Error("Expected error since PostgreSQL storage is not implemented")
		}

		if storage != nil {
			t.Error("Expected storage to be nil when PostgreSQL is not implemented")
		}

		expectedError := "пока не реализовано"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func TestCreateStorage_UnknownType(t *testing.T) {
	tempDir := t.TempDir()
	saveInterval := 30 * time.Second

	storage, err := CreateStorage("unknown", tempDir, saveInterval)
	if err == nil {
		t.Error("Expected error for unknown storage type")
		return
	}

	if storage != nil {
		t.Error("Expected storage to be nil for unknown type")
	}

	expectedError := "неизвестный тип хранилища"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCreateStorage_StorageConstants(t *testing.T) {
	// Проверяем, что константы определены правильно
	if StorageTypeJSON == "" {
		t.Error("StorageTypeJSON should not be empty")
	}
	if StorageTypeJSON != "json" {
		t.Errorf("Expected StorageTypeJSON to be 'json', got '%s'", StorageTypeJSON)
	}

	if StorageTypePostgres == "" {
		t.Error("StorageTypePostgres should not be empty")
	}
	if StorageTypePostgres != "postgres" {
		t.Errorf("Expected StorageTypePostgres to be 'postgres', got '%s'", StorageTypePostgres)
	}
}

func TestCreateStorage_PostgresCaseInsensitive(t *testing.T) {
	tempDir := t.TempDir()

	originalEnv := os.Getenv("MPM_DATABASE_URL")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("MPM_DATABASE_URL", originalEnv)
		} else {
			_ = os.Unsetenv("MPM_DATABASE_URL")
		}
	}()

	_ = os.Setenv("MPM_DATABASE_URL", "postgresql://user:pass@localhost/db")

	testCases := []string{"POSTGRES", "Postgres", "PostgreS", "postgres"}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			storage, err := CreateStorage(tc, tempDir, time.Second)
			if err == nil {
				t.Errorf("Expected error for '%s' (not implemented)", tc)
			}

			if storage != nil {
				t.Errorf("Expected storage to be nil for '%s' (not implemented)", tc)
			}

			expectedError := "пока не реализовано"
			if !strings.Contains(err.Error(), expectedError) {
				t.Errorf("Expected error to contain '%s' for '%s', got '%s'", expectedError, tc, err.Error())
			}
		})
	}
}
