package repository

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// Константы для типов хранилищ
const (
	StorageTypeJSON     = "json"
	StorageTypePostgres = "postgres"
	// В будущем можно добавить другие типы: redis, mysql и т.д.
)

// CreateStorage создает экземпляр хранилища нужного типа
func CreateStorage() (EntityStorage, error) {
	// Получаем тип хранилища из переменной окружения
	storageType := os.Getenv("MPM_STORAGE_TYPE")
	if storageType == "" {
		storageType = StorageTypeJSON // По умолчанию используем JSON
		log.Printf("Тип хранилища не указан, используется тип по умолчанию: %s", storageType)
	}

	// Приводим к нижнему регистру для удобства
	storageType = strings.ToLower(storageType)

	// Создаем хранилище в зависимости от указанного типа
	switch storageType {
	case StorageTypeJSON:
		// Получаем путь к директории данных
		dataDir := os.Getenv("MPM_DATA_PATH")
		if dataDir == "" {
			dataDir = getDefaultDataDir()
		}
		// Получаем интервал сохранения (в секундах)
		saveInterval := 30 * time.Second
		if intervalStr := os.Getenv("MPM_SAVE_INTERVAL"); intervalStr != "" {
			if interval, err := time.ParseDuration(intervalStr); err == nil {
				saveInterval = interval
			} else {
				log.Printf("Ошибка разбора интервала сохранения: %v, используется значение по умолчанию: %v", err, saveInterval)
			}
		}

		log.Printf("Используется JSON-хранилище с директорией %s и интервалом сохранения %v", dataDir, saveInterval)
		return NewJSONStorage(dataDir, saveInterval), nil

	case StorageTypePostgres:
		// Получаем строку подключения к PostgreSQL
		connStr := os.Getenv("MPM_DATABASE_URL")
		if connStr == "" {
			return nil, fmt.Errorf("не указана строка подключения к PostgreSQL (MPM_DATABASE_URL)")
		}

		log.Printf("Используется PostgreSQL-хранилище с подключением: %s", connStr)
		// Здесь будет вызов конструктора PostgreSQL-хранилища когда оно будет реализовано
		return nil, fmt.Errorf("хранилище типа %s пока не реализовано", storageType)

	default:
		return nil, fmt.Errorf("неизвестный тип хранилища: %s", storageType)
	}
}
