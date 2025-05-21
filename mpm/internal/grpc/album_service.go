package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"mpm/internal/repository"
	pb "mpm/proto/albums"
	"time"
)

type AlbumServer struct {
	pb.UnimplementedAlbumServiceServer
	repository *repository.Repository
}

func NewAlbumServer(repo *repository.Repository) *AlbumServer {
	return &AlbumServer{
		repository: repo,
	}
}

func (s *AlbumServer) GetAlbums(ctx context.Context, req *pb.GetAlbumsRequest) (*pb.GetAlbumsResponse, error) {
	albums, err := s.repository.GetAllAlbums(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка получения списка альбомов: %v", err)
	}

	result := &pb.GetAlbumsResponse{
		Albums: make([]*pb.Album, 0, len(albums)),
	}

	for _, album := range albums {
		result.Albums = append(result.Albums, &pb.Album{
			Id:          int32(album.ID),
			Name:        album.Name,
			Description: album.Description,
			CreatedAt:   album.CreatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}

func (s *AlbumServer) CreateAlbum(ctx context.Context, req *pb.CreateAlbumRequest) (*pb.Album, error) {
	// Создаем модель альбома для сохранения в репозитории
	album := models.Album{
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
	}

	// Добавляем альбом в репозиторий
	id, err := s.repository.AddAlbum(ctx, album)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка создания альбома: %v", err)
	}

	// Получаем только что созданный альбом для возврата полных данных
	createdAlbum, err := s.repository.FindAlbumByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка получения созданного альбома: %v", err)
	}

	// Преобразуем в формат proto и возвращаем
	return &pb.Album{
		Id:          int32(createdAlbum.ID),
		Name:        createdAlbum.Name,
		Description: createdAlbum.Description,
		CreatedAt:   createdAlbum.CreatedAt.Format(time.RFC3339),
	}, nil
}
