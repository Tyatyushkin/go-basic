package models

import (
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
