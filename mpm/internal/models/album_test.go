package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlbum_Validate(t *testing.T) {
	tests := []struct {
		name     string
		album    Album
		wantErr  bool
		errorMsg string
	}{
		{
			name: "Valid album",
			album: Album{
				Name:        "Летний отпуск",
				Description: "Фотографии с отпуска на море",
				Tags:        []string{"лето", "море", "отпуск"},
			},
			wantErr: false,
		},
		{
			name: "Empty name",
			album: Album{
				Name:        "",
				Description: "Описание альбома",
				Tags:        []string{"тег1", "тег2"},
			},
			wantErr:  true,
			errorMsg: "название альбома не может быть пустым",
		},
		{
			name: "Name too long",
			album: Album{
				Name:        string(make([]rune, 101)), // 101 символ
				Description: "Описание альбома",
				Tags:        []string{"тег1", "тег2"},
			},
			wantErr:  true,
			errorMsg: "название альбома слишком длинное",
		},
		{
			name: "Description too long",
			album: Album{
				Name:        "Нормальное название",
				Description: string(make([]rune, 501)), // 501 символ
				Tags:        []string{"тег1", "тег2"},
			},
			wantErr:  true,
			errorMsg: "описание альбома слишком длинное",
		},
		{
			name: "Invalid tags",
			album: Album{
				Name:        "Альбом с тегами",
				Description: "Описание",
				Tags:        []string{"tag with space", ""},
			},
			wantErr:  true,
			errorMsg: "теги содержат недопустимые символы",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.album.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAlbum_GetID(t *testing.T) {
	album := Album{ID: 42}
	assert.Equal(t, 42, album.GetID(), "GetID должен возвращать корректный ID альбома")
}

func TestAlbum_GetType(t *testing.T) {
	album := Album{}
	assert.Equal(t, "album", album.GetType(), "GetType должен возвращать строку 'album'")
}

func TestAlbum_Serialization(t *testing.T) {
	album := Album{
		ID:          1,
		Name:        "Летний отпуск",
		Description: "Фотографии с отпуска на море",
		Tags:        []string{"лето", "море", "отпуск"},
	}

	// Тест сериализации в JSON
	t.Run("Serialize to JSON", func(t *testing.T) {
		data, err := json.Marshal(album)
		assert.NoError(t, err, "Ошибка при сериализации альбома")
		assert.Contains(t, string(data), `"name":"Летний отпуск"`, "JSON должен содержать корректное название")
		assert.Contains(t, string(data), `"tags":["лето","море","отпуск"]`, "JSON должен содержать корректные теги")
	})

	// Тест десериализации из JSON
	t.Run("Deserialize from JSON", func(t *testing.T) {
		jsonData := `{"id":1,"name":"Летний отпуск","description":"Фотографии с отпуска на море","tags":["лето","море","отпуск"]}`
		var deserializedAlbum Album
		err := json.Unmarshal([]byte(jsonData), &deserializedAlbum)
		assert.NoError(t, err, "Ошибка при десериализации альбома")
		assert.Equal(t, album, deserializedAlbum, "Десериализованный альбом должен совпадать с оригиналом")
	})
}
