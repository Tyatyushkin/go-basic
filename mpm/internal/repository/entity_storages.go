package repository

import "mpm/internal/models"

type EntityStorage interface {
	// Save Сохранить сущность
	Save(entity models.Entity) error

	// Load Загрузить данные из хранилища
	Load() error

	// Persist Сохранить текущее состояние в хранилище
	Persist() error
}
