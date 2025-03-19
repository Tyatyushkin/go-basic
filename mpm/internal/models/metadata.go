package models

// Metadata хранит различные метаданные для фотографий
type Metadata struct {
	Key   string `json:"key" db:"key"`     // Тип метаданных (например, "camera", "location", "date_taken")
	Value string `json:"value" db:"value"` // Значение метаданных
}
